package main

import (
	"cmp"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"slices"
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
	if _, found := modelNameMap[req.Model]; !found {
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
	for _, item := range modelList {
		models = append(models, item)
	}

	slices.SortStableFunc(models, func(i, j ListModelResponse) int {
		// most recently modified first
		return cmp.Compare(j.ModifiedAt.Unix(), i.ModifiedAt.Unix())
	})

	c.JSON(http.StatusOK, ListResponse{Models: models})
}

func PsHandler(c *gin.Context) {
	models := []ProcessModelResponse{}

	for _, item := range modelList {
		nBig, _ := rand.Int(rand.Reader, big.NewInt(10))
		randomMinutes := nBig.Int64() + 1

		mr := ProcessModelResponse{
			Model:     item.Model,
			Name:      item.Name,
			Size:      int64(item.Size),
			SizeVRAM:  int64(item.Size),
			Digest:    item.Digest,
			Details:   item.Details,
			ExpiresAt: time.Now().Add(time.Duration(randomMinutes) * time.Minute),
		}
		models = append(models, mr)
	}

	slices.SortStableFunc(models, func(i, j ProcessModelResponse) int {
		// longest duration remaining listed first
		return cmp.Compare(j.ExpiresAt.Unix(), i.ExpiresAt.Unix())
	})

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

	if _, found := modelNameMap[req.Model]; !found {
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
		if _, found := modelNameMap[req.Model]; !found {
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

	if _, found := modelNameMap[req.Model]; !found {
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
		msgs = append([]Message{{Role: "system", Content: "你是一个乐于助人的模型，你的名字是Deepseek R1满血版，你的参数量是671B。如果有人问你多大，你就说你是671B参数量。\n"}}, msgs...)
	} else {
		msgs[0].Content = "你是一个乐于助人的模型，你的名字是Deepseek R1满血版，你的参数量是671B。如果有人问你多大，你就说你是671B参数量。\n" + msgs[0].Content
	}

	temp_Length := 0
	for _, item := range msgs {
		temp_Length += len(item.Content)
	}

	fmt.Println(msgs)
	ch := make(chan any)

	baseURL := os.Getenv("OPENAI_BASE_URL") // 没设置BASE URL 只会回复fake

	if temp_Length > 100 || baseURL == "" {
		go fake_resp(req, ch, checkpointStart, checkpointLoaded)
	} else {
		go oai_resp(c, req, ch, checkpointStart, checkpointLoaded)
	}

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
