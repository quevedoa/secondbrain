package llm

import (
	"context"
	"secondbrain/src/entity"
)

type LLMGateway interface {
	Ask(ctx context.Context, req *entity.LLMRequest) (*entity.LLMResponse, error)
}
