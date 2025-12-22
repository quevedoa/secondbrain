package repository

import (
	"context"
	"secondbrain/src/entity"
)

type VectorRepository interface {
	Create(ctx context.Context, id, content string) (*entity.Note, error)
	Query(ctx context.Context, query string, numResults int32) ([]*entity.Note, error)
}
