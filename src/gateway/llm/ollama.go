package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"secondbrain/src/entity"
	"secondbrain/src/mapper"
	"strings"
)

const (
	endpoint = "http://localhost:11434/api/generate"
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

	var ollamaRes entity.OllamaResponse
	err = json.Unmarshal(bodyBytes, &ollamaRes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Ollama response: %s", err)
	}

	return mapper.ToLLMResponse(&ollamaRes), nil
}

func (o *OllamaClient) AskStream(ctx context.Context, req *entity.LLMRequest) <-chan entity.LLMResponse {
	chunkChannel := make(chan entity.LLMResponse)

	go func() {
		httpReq, err := json.Marshal(mapper.ToOllamaRequest(req))
		if err != nil {
			sendToChannel(ctx, chunkChannel, entity.LLMResponse{Error: fmt.Errorf("failed to marshal http request to Ollama: %s", err)})
			return
		}

		httpRes, err := o.Client.Post(endpoint, "application/json", bytes.NewReader(httpReq))
		if err != nil {
			sendToChannel(ctx, chunkChannel, entity.LLMResponse{Error: fmt.Errorf("failed to send POST http request to Ollama: %s", err)})
			return
		}
		defer httpRes.Body.Close()

		sc := bufio.NewScanner(httpRes.Body)
		sc.Buffer(make([]byte, 64*1024), 1024*1024)

		for sc.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}

			var ollamaRes entity.OllamaResponse
			err = json.Unmarshal([]byte(line), &ollamaRes)
			if err != nil {
				sendToChannel(ctx, chunkChannel, entity.LLMResponse{Error: fmt.Errorf("failed to unmarshal Ollama response: %s", err)})
				return
			}

			chunk := mapper.ToLLMResponse(&ollamaRes)
			sendToChannel(ctx, chunkChannel, *chunk)
		}
	}()

	return chunkChannel
}

func sendToChannel(ctx context.Context, chunkChannel chan<- entity.LLMResponse, r entity.LLMResponse) {
	select {
	case <-ctx.Done():
		return
	case chunkChannel <- r:
	}
}
