package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"secondbrain/src/entity"
	"secondbrain/src/mapper"
)

const (
	endpoint = "https://localhost:11434/api/generate"
)

type OllamaClient struct {
	Client *http.Client
}

func NewOllamaClient(client *http.Client) *OllamaClient {
	return &OllamaClient{
		Client: client,
	}
}

func (o *OllamaClient) Ask(ctx context.Context, req *entity.LLMRequest) (*entity.LLMResponse, error) {
	httpReq, err := json.Marshal(mapper.ToOllamaRequest(req))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal http request to Ollama: %s", err)
	}

	httpRes, err := o.Client.Post(endpoint, "application/json", bytes.NewReader(httpReq))
	if err != nil {
		return nil, fmt.Errorf("failed to send POST http request to Ollama: %s", err)
	}
	defer httpRes.Body.Close()

	bodyBytes, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP Response body from Ollama: %s", err)
	}

	var ollamaRes *entity.OllamaResponse
	err = json.Unmarshal(bodyBytes, ollamaRes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Ollama response: %s", err)
	}

	return mapper.ToLLMResponse([]*entity.OllamaResponse{ollamaRes}), nil
}
