package httpapi

import (
	"net/http"

	"github.com/Deahesi/deahesi-os-backend/internal/ai"
	aihandlers "github.com/Deahesi/deahesi-os-backend/internal/httpapi/handlers/ai-handlers"
	"github.com/Deahesi/deahesi-os-backend/internal/llm"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	router *chi.Mux
}

func NewServer(
	aiService *ai.AIService,
	llmService *llm.LLMService,
) *Server {
	router := chi.NewRouter()

	aiHandler := aihandlers.NewAIHandler(aiService)
	// authHandler := NewAuthHandler(authService)

	router.Route("/ai", func(r chi.Router) {
		r.Post("/message", aiHandler.SendMessage)
	})

	return &Server{
		router: router,
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}
