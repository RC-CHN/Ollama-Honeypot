# Ollama-Honeypot
The Honeypot of Ollama

Inference API均已经实现

```

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

```


可以使用OAI和Ollama的API接口进行“推理”

![效果展示](893b4ed6c2a11a2fbd5466fae5ee318c.png)