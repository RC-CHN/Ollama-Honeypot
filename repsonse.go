package main

import "time"

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
		time.Sleep(100 * time.Millisecond)
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
