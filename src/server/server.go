package server

import (
	"net/http"
	"secondbrain/src/gateway/llm"
	repo "secondbrain/src/repository/vector_repository"
)

type Server struct {
	VectorRepo repo.VectorRepository
	LLMGateway llm.LLMGateway
}

func New(vectorRepo repo.VectorRepository, llmGateway llm.LLMGateway) *Server {
	return &Server{
		VectorRepo: vectorRepo,
		LLMGateway: llmGateway,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/CreateNote", s.CreateNote)
	mux.HandleFunc("/QueryNotes", s.QueryNotes)
	return mux
}
