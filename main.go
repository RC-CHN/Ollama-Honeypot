package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowWildcard = true
	corsConfig.AllowBrowserExtensions = true
	corsConfig.AllowHeaders = []string{
		"Authorization",
		"Content-Type",
		"User-Agent",
		"Accept",
		"X-Requested-With",

		// OpenAI compatibility headers
		"x-stainless-lang",
		"x-stainless-package-version",
		"x-stainless-os",
		"x-stainless-arch",
		"x-stainless-retry-count",
		"x-stainless-runtime",
		"x-stainless-runtime-version",
		"x-stainless-async",
		"x-stainless-helper-method",
		"x-stainless-poll-helper",
		"x-stainless-custom-poll-interval",
		"x-stainless-timeout",
	}
	corsConfig.AllowAllOrigins = true

	r := gin.Default()
	r.Use(
		cors.New(corsConfig),
		// allowedHostsMiddleware(s.addr),
	)

	// General
	r.HEAD("/", func(c *gin.Context) { c.String(http.StatusOK, "Ollama is running") })
	r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "Ollama is running") })
	r.HEAD("/api/version", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"version": "0.6.0"}) }) //version.Version
	r.GET("/api/version", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"version": "0.6.0"}) })  // version.Version

	// Local model cache management (new implementation is at end of function)
	r.POST("/api/pull", PullHandler)
	r.POST("/api/push", PushHandler)
	r.HEAD("/api/tags", ListHandler)
	r.GET("/api/tags", ListHandler)
	r.POST("/api/show", ShowHandler)
	r.DELETE("/api/delete", DeleteHandler)

	// Create
	r.POST("/api/create", CreateHandler)
	r.POST("/api/blobs/:digest", CreateBlobHandler)
	r.HEAD("/api/blobs/:digest", HeadBlobHandler)
	r.POST("/api/copy", CopyHandler)

	// Inference
	r.GET("/api/ps", PsHandler)
	r.POST("/api/generate", GenerateHandler)
	r.POST("/api/chat", ChatHandler)
	r.POST("/api/embed", EmbedHandler)
	r.POST("/api/embeddings", EmbeddingsHandler)

	// Inference (OpenAI compatibility)
	r.POST("/v1/chat/completions", ChatMiddleware(), ChatHandler)
	r.POST("/v1/completions", GenerateHandler)
	r.POST("/v1/embeddings", EmbedHandler)
	r.GET("/v1/models", ListMiddleware(), ListHandler)
	r.GET("/v1/models/:model", RetrieveMiddleware(), ShowHandler)

	r.Run("0.0.0.0:11434")
}
