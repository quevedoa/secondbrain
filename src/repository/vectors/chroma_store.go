package repository

import (
	"context"
	"fmt"
	"secondbrain/src/entity"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/google/uuid"
)

type ChromaStore struct {
	Client     chroma.Client
	Collection chroma.Collection
}

func NewChromaStore(client chroma.Client, collection chroma.Collection) *ChromaStore {
	return &ChromaStore{
		Client:     client,
		Collection: collection,
	}
}

func (c *ChromaStore) Create(ctx context.Context, id, content string) (*entity.Note, error) {
	if id == "" {
		id = uuid.NewString()
	}

	err := c.Collection.Add(ctx,
		chroma.WithIDs(chroma.DocumentID(id)),
		chroma.WithTexts(content),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add to chroma collection: %s", err)
	}

	return &entity.Note{
		ID:      id,
		Content: content,
	}, nil
}

func (c *ChromaStore) Query(ctx context.Context, query string, numResults int32) ([]*entity.Note, error) {
	qr, err := c.Collection.Query(ctx, chroma.WithQueryTexts(query))
	if err != nil {
		return nil, fmt.Errorf("failed to query chroma collection: %s", err)
	}

	docs := qr.GetDocumentsGroups()
	ids := qr.GetIDGroups()
	notes := make([]*entity.Note, len(docs[0]))

	if len(docs) != len(ids) {
		return nil, fmt.Errorf("failed due to different sized document and id slices")
	}

	for idx, d := range docs[0] { // First index because its only one query
		notes[idx] = &entity.Note{
			ID:      string(ids[0][idx]),
			Content: d.ContentString(),
		}
	}

	return notes, nil
}

func (c *ChromaStore) Close() error {
	if c.Client == nil {
		return nil
	}
	return c.Client.Close()
}
