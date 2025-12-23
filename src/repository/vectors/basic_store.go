package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"secondbrain/src/entity"

	"github.com/google/uuid"
)

type BasicStore struct {
	Path string
}

func NewBasicStore(path string) *BasicStore {
	return &BasicStore{
		Path: path,
	}
}

func (b *BasicStore) Create(ctx context.Context, id, content string) (*entity.Note, error) {
	if id == "" {
		id = uuid.NewString()
	}

	path := filepath.Join(b.Path, fmt.Sprintf("%s.txt", id))
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return nil, err
	}
	f.Sync()

	return &entity.Note{
		ID:      id,
		Content: content,
	}, nil
}

func (b *BasicStore) Query(ctx context.Context, query string, _ int32) ([]*entity.Note, error) {
	// We'll treat query as the name of the file
	path := filepath.Join(b.Path, query)
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}

	return []*entity.Note{
		{
			ID:      query,
			Content: string(dat),
		},
	}, nil
}
