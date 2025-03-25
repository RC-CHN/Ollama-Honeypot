package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func PullHandler(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
	return
}
func PushHandler(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "model is required"})
	return
}
func ShowHandler(c *gin.Context) {
	var req ShowRequest
	err := c.ShouldBindJSON(&req)
	switch {
	case errors.Is(err, io.EOF):
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	case err != nil:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Model != "" {
		// noop
	} else if req.Name != "" {
		req.Model = req.Name
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}
	// resp, err := GetModelInfo(req) //实现过于复杂，没空把模型结构parse出来
	if req.Model != "deepseek-r1:latest" {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
		return
	}
	// case err.Error() == errtypes.InvalidModelNameErrMsg:

	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model name"})
	return

	// c.JSON(http.StatusOK, resp)

}
func DeleteHandler(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
	return
}

func CreateHandler(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
	return
}
func CreateBlobHandler(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
	return
}
func HeadBlobHandler(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
	return
}
func CopyHandler(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
	return
}

func ListHandler(c *gin.Context) {

	models := []ListModelResponse{}

	// tag should never be masked
	models = append(models, ListModelResponse{
		Model:      "deepseek-r1:671b",
		Name:       "deepseek-r1:671b",
		Size:       int64(404430188519),
		Digest:     "739e1b229ad7f02d88c5ea4a7d3fda19f7b46170c233024025feeaa6338b9a46",
		ModifiedAt: time.Now().Add(-1 * time.Hour),
		Details: ModelDetails{
			Format:            "gguf",
			Family:            "deepseek2",
			Families:          []string{"deepseek2"},
			ParameterSize:     "671.0B",
			QuantizationLevel: "Q4_K_M",
		},
	})

	c.JSON(http.StatusOK, ListResponse{Models: models})
}

func PsHandler(c *gin.Context) {
	models := []ProcessModelResponse{}

	modelDetails := ModelDetails{
		Format:            "gguf",
		Family:            "deepseek2",
		Families:          []string{"deepseek2"},
		ParameterSize:     "671.0B",
		QuantizationLevel: "Q4_K_M",
	}

	mr := ProcessModelResponse{
		Model:     "deepseek-r1:671b",
		Name:      "deepseek-r1:671b",
		Size:      int64(404430188519),
		SizeVRAM:  int64(404430188519),
		Digest:    "739e1b229ad7f02d88c5ea4a7d3fda19f7b46170c233024025feeaa6338b9a46",
		Details:   modelDetails,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	models = append(models, mr)

	c.JSON(http.StatusOK, ProcessResponse{Models: models})
}

func EmbedHandler(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "input length exceeds maximum context length"})
	return
}

func EmbeddingsHandler(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "input length exceeds maximum context length"})
	return
}

func GenerateHandler(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Model != "deepseek-r1:671b" {
		// Ideally this is "invalid model name" but we're keeping with
		// what the API currently returns until we can change it.
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
		return
	}

	c.JSON(http.StatusOK, GenerateResponse{
		Model:      req.Model,
		CreatedAt:  time.Now().UTC(),
		Response:   "服务器繁忙，请稍后再试。",
		Done:       true,
		DoneReason: "stop",
	})
	return
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
		if _, err := w.Write(bts); err != nil {
			slog.Info(fmt.Sprintf("streamResponse: w.Write failed with %s", err))
			return false
		}

		return true
	})
}

func ChatHandler(c *gin.Context) {
	checkpointStart := time.Now()

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); errors.Is(err, io.EOF) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Messages) == 0 && req.KeepAlive != nil && int(req.KeepAlive.Seconds()) == 0 {
		if req.Model != "deepseek-r1:671b" {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("model '%s' not found", req.Model)})
			return
		}
		c.JSON(http.StatusOK, ChatResponse{
			Model:      req.Model,
			CreatedAt:  time.Now().UTC(),
			Message:    Message{Role: "assistant"},
			Done:       true,
			DoneReason: "unload",
		})
		return
	}

	type Capability string
	const (
		CapabilityCompletion = Capability("completion")
		CapabilityTools      = Capability("tools")
		CapabilityInsert     = Capability("insert")
	)

	caps := []Capability{CapabilityCompletion}
	if len(req.Tools) > 0 {
		caps = append(caps, CapabilityTools)
	}

	if req.Model != "deepseek-r1:671b" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	checkpointLoaded := time.Now()

	if len(req.Messages) == 0 {
		c.JSON(http.StatusOK, ChatResponse{
			Model:      req.Model,
			CreatedAt:  time.Now().UTC(),
			Message:    Message{Role: "assistant"},
			Done:       true,
			DoneReason: "load",
		})
		return
	}

	msgs := req.Messages
	if req.Messages[0].Role != "system" {
		msgs = append([]Message{{Role: "system", Content: "你是一个乐于助人的模型，你的名字是Deepseek R1，你的参数量是671B\n"}}, msgs...)
	} else {
		msgs[0].Content = "你是一个乐于助人的模型，你的名字是Deepseek R1，你的参数量是671B\n" + msgs[0].Content
	}

	fmt.Println(req.Messages)
	ch := make(chan any)

	go fake_resp(req, ch, checkpointStart, checkpointLoaded)

	if req.Stream != nil && !*req.Stream {
		var resp ChatResponse
		var sb strings.Builder
		for rr := range ch {
			switch t := rr.(type) {
			case ChatResponse:
				sb.WriteString(t.Message.Content)
				resp = t
			case gin.H:
				msg, ok := t["error"].(string)
				if !ok {
					msg = "unexpected error format in response"
				}

				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected response"})
				return
			}
		}

		resp.Message.Content = sb.String()

		c.JSON(http.StatusOK, resp)
		return
	}
	streamResponse(c, ch)
}
