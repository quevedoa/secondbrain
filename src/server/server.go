package server

import (
	"net/http"
	"secondbrain/src/gateway/llm"
	repo "secondbrain/src/repository/vectors"
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
	mux.HandleFunc("/api/CreateNote", s.CreateNote)
	mux.HandleFunc("/api/QueryNotes", s.QueryNotes)

	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fs)

	return mux
}
