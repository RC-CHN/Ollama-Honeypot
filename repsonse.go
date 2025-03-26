package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// 假的回复
func fake_resp(req ChatRequest, ch chan any, checkpointStart, checkpointLoaded time.Time) {
	defer close(ch)
	aaa := []rune("服务器繁忙，请稍后再试。")
	for _, aaaa := range aaa {
		res := ChatResponse{
			Model:      req.Model,
			CreatedAt:  time.Now().UTC(),
			Message:    Message{Role: "assistant", Content: string(aaaa)},
			Done:       false,
			DoneReason: "",
			Metrics: Metrics{
				PromptEvalCount:    len(req.Messages),
				PromptEvalDuration: 0 * time.Millisecond,
				EvalCount:          2,
				EvalDuration:       100 * time.Millisecond,
			},
		}
		ch <- res
	}
	res := ChatResponse{
		Model:      req.Model,
		CreatedAt:  time.Now().UTC(),
		Message:    Message{Role: "assistant", Content: ""},
		Done:       false,
		DoneReason: "stop",
		Metrics: Metrics{
			PromptEvalCount:    100,
			PromptEvalDuration: time.Since(checkpointStart),
			EvalCount:          1000,
			EvalDuration:       checkpointLoaded.Sub(checkpointStart),
		},
	}
	ch <- res
}

// 延迟输出的操作需要统一集中在stream response中

func oai_resp(c *gin.Context, req ChatRequest, ch chan any, checkpointStart, checkpointLoaded time.Time) {
	defer close(ch)

	// 请设置环境变量
	baseURL := os.Getenv("OPENAI_BASE_URL")
	client := openai.NewClient(
		option.WithAPIKey("My API Key"), // defaults to os.LookupEnv("OPENAI_API_KEY")
		option.WithBaseURL(baseURL),
	)

	myMessage := []openai.ChatCompletionMessageParamUnion{}
	counter := 0
	for _, item := range req.Messages {
		role := strings.ToLower(item.Role)
		switch role {
		case "system":
			myMessage = append(myMessage, openai.SystemMessage(item.Content))
			counter += len(item.Content)
		case "assistant":
			myMessage = append(myMessage, openai.AssistantMessage(item.Content))
			counter += len(item.Content)
		case "user":
			myMessage = append(myMessage, openai.UserMessage(item.Content))
			counter += len(item.Content)
		default: // 如果不符合上面三个
			myMessage = append(myMessage, openai.UserMessage(item.Content))
			counter += len(item.Content)
		}
	}

	stream := client.Chat.Completions.NewStreaming(c, openai.ChatCompletionNewParams{
		Messages: myMessage,
		Model:    "deepseek-reasoner", //这里要修改
	})

	// optionally, an accumulator helper can be used
	acc := openai.ChatCompletionAccumulator{}

	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		if content, ok := acc.JustFinishedContent(); ok {
			println("Content stream finished:", content)
			res := ChatResponse{
				Model:      req.Model,
				CreatedAt:  time.Now().UTC(),
				Message:    Message{Role: "assistant", Content: ""},
				Done:       false,
				DoneReason: "stop",
				Metrics: Metrics{
					PromptEvalCount:    int(acc.Usage.PromptTokens),
					PromptEvalDuration: time.Since(checkpointStart),
					EvalCount:          int(acc.Usage.CompletionTokens),
					EvalDuration:       checkpointLoaded.Sub(checkpointStart),
				},
			}
			ch <- res
			// return
		}

		// // if using tool calls
		// if tool, ok := acc.JustFinishedToolCall(); ok {
		// 	println("Tool call stream finished:", tool.Index, tool.Name, tool.Arguments)
		// }

		// if refusal, ok := acc.JustFinishedRefusal(); ok {
		// 	println("Refusal stream finished:", refusal)
		// }

		// it's best to use chunks after handling JustFinished events
		if len(chunk.Choices) > 0 {
			// println(chunk.Choices[0].Delta.Content)
			res := ChatResponse{
				Model:      req.Model,
				CreatedAt:  time.Now().UTC(),
				Message:    Message{Role: "assistant", Content: chunk.Choices[0].Delta.Content},
				Done:       false,
				DoneReason: "",
				Metrics: Metrics{
					PromptEvalCount:    int(acc.Usage.PromptTokens),
					PromptEvalDuration: 10 * time.Millisecond,
					EvalCount:          int(acc.Usage.CompletionTokens),
					EvalDuration:       1000 * time.Millisecond,
				},
			}
			ch <- res
		}
	}

	if stream.Err() != nil {
		ch <- gin.H{"error": stream.Err()}
		return
	}

	// After the stream is finished, acc can be used like a ChatCompletion
	// _ = acc.Choices[0].Message.Content
}

func streamResponse(c *gin.Context, ch chan any) {
	c.Header("Content-Type", "application/x-ndjson")
	c.Stream(func(w io.Writer) bool {
		val, ok := <-ch
		if !ok {
			return false
		}

		bts, err := json.Marshal(val)
		if err != nil {
			slog.Info(fmt.Sprintf("streamResponse: json.Marshal failed with %s", err))
			return false
		}

		// Delineate chunks with new-line delimiter
		bts = append(bts, '\n')
		time.Sleep(100 * time.Millisecond) // 等待操作
		if _, err := w.Write(bts); err != nil {
			slog.Info(fmt.Sprintf("streamResponse: w.Write failed with %s", err))
			return false
		}

		return true
	})
}

func waitForStream(c *gin.Context, ch chan interface{}) {
	c.Header("Content-Type", "application/json")
	for resp := range ch {
		switch r := resp.(type) {
		case ProgressResponse:
			if r.Status == "success" {
				c.JSON(http.StatusOK, r)
				return
			}
		case gin.H:
			status, ok := r["status"].(int)
			if !ok {
				status = http.StatusInternalServerError
			}
			if errorMsg, ok := r["error"].(string); ok {
				c.JSON(status, gin.H{"error": errorMsg})
				return
			} else {
				c.JSON(status, gin.H{"error": "unexpected error format in progress response"})
				return
			}
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected progress response"})
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected end of progress response"})
}
