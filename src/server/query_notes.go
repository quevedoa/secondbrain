package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"secondbrain/src/entity"
	"secondbrain/src/utils"
)

func (s *Server) QueryNotes(w http.ResponseWriter, r *http.Request) {

	// Unmarshal request's query
	var query entity.Query
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch related context from vector repo
	queriedNotes, err := s.VectorRepo.Query(r.Context(), query.Text, query.NumResults)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Stream LLM response from query and context
	chunkChannel := s.LLMGateway.AskStream(r.Context(), &entity.LLMRequest{
		Model: "llama3",
		Prompt: entity.Prompt{
			Content: buildPrompt(queriedNotes, query.Text),
		},
		Stream: true,
	})

	for chunk := range chunkChannel {
		if chunk.Error != nil {
			utils.SendSSE(w, "error", chunk.Error.Error())
			flusher.Flush()
			return
		}

		if chunk.Done {
			utils.SendSSE(w, "done", chunk.DoneReason)
			flusher.Flush()
			return
		}

		if chunk.Response != "" {
			utils.SendSSE(w, "response", chunk.Response)
			flusher.Flush()
		}
	}
}

func buildPrompt(queriedNotes []*entity.Note, query string) string {
	context := ""
	for i, note := range queriedNotes {
		context += fmt.Sprintf("Note %d: %s\n", i, note.Content)
	}

	promptOutline := `
	You are a helpful assistant. Answer the user's question using ONLY the notes provided below.
	If the notes do not contain the answer, say you don't have enough information and end the conversation.
	Keep responses as short as possible.

	User question:
	%s

	Notes:
	%s
	`

	return fmt.Sprintf(promptOutline, query, context)
}
