package main

import (
	"context"
	"fmt"
	"net/http"
	"secondbrain/src/gateway/llm"
	"secondbrain/src/middleware"
	vectorrepo "secondbrain/src/repository/vectors"
	"secondbrain/src/server"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
)

func main() {
	ctx := context.Background()
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	vectorRepo := createVectorRepo(ctx)
	if vectorRepo == nil {
		return
	}
	defer vectorRepo.Close()

	llmGateway := createLLMGateway(httpClient)
	if llmGateway == nil {
		return
	}

	server := server.New(
		vectorRepo,
		llmGateway,
	)

	fmt.Println("Listening on port 8080...")
	http.ListenAndServe(":8080", middleware.WithCORS(server.Routes()))
}

func createVectorRepo(ctx context.Context) *vectorrepo.ChromaStore {
	fmt.Println("Creating Chroma HTTP client...")
	chromaClient, err := chroma.NewHTTPClient()
	if err != nil {
		fmt.Printf("failed to create Chroma client: %s", err)
		return nil
	}

	fmt.Println("Creating Chroma collection...")
	chromaCollection, err := chromaClient.GetOrCreateCollection(ctx, "secondbrain_chroma_db")
	if err != nil {
		fmt.Printf("failed to create Chroma collection: %s", err)
		return nil
	}

	fmt.Println("Creating Chroma store...")
	return vectorrepo.NewChromaStore(chromaClient, chromaCollection)
}

func createLLMGateway(client *http.Client) *llm.OllamaClient {
	return llm.NewOllamaClient(client)
}
