package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"secondbrain/src/entity"
)

func (s *Server) CreateNote(w http.ResponseWriter, r *http.Request) {
	var note entity.Note
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdNote, err := s.VectorRepo.Create(r.Context(), "", note.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Created note %s with content %s", createdNote.ID, createdNote.Content)

	res, err := json.Marshal(createdNote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}
