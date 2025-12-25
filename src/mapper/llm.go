package mapper

import (
	"secondbrain/src/entity"
)

func ToOllamaRequest(req *entity.LLMRequest) *entity.OllamaRequest {
	return &entity.OllamaRequest{
		Model:  req.Model,
		Prompt: req.Prompt.Content,
		Stream: req.Stream,
	}
}

func ToLLMResponse(ollamaResponse *entity.OllamaResponse) *entity.LLMResponse {
	return &entity.LLMResponse{
		Response:   ollamaResponse.Response,
		Done:       ollamaResponse.Done,
		DoneReason: ollamaResponse.DoneReason,
	}
}
