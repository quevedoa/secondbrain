package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"secondbrain/src/entity"
)

func (s *Server) QueryNotes(w http.ResponseWriter, r *http.Request) {
	var query entity.Query
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	queriedNotes, err := s.VectorRepo.Query(r.Context(), query.Text, query.NumResults)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	prompt := fmt.Sprintf(promptOutline, query.Text, context)

	llmRes, err := s.LLMGateway.Ask(r.Context(), &entity.LLMRequest{
		Model: "llama3",
		Prompts: []entity.Prompt{
			{
				Content: prompt,
			},
		},
		Stream: false,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(llmRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusFound)
	w.Write(res)
}
