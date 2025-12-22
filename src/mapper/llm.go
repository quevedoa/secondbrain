package mapper

import (
	"secondbrain/src/entity"
)

func ToOllamaRequest(req *entity.LLMRequest) []*entity.OllamaRequest {
	ollamaReqs := make([]*entity.OllamaRequest, len(req.Prompts))
	for idx, p := range req.Prompts {
		ollamaReqs[idx] = &entity.OllamaRequest{
			Model:  req.Model,
			Prompt: p.Content,
			Stream: req.Stream,
		}
	}
	return ollamaReqs
}

func ToLLMResponse(ollamaResponses []*entity.OllamaResponse) *entity.LLMResponse {
	responses := make([]string, len(ollamaResponses))
	for idx, ollamaRes := range ollamaResponses {
		responses[idx] = ollamaRes.Response
	}

	return &entity.LLMResponse{
		Responses: responses,
	}
}
