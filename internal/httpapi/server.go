package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	router *chi.Mux
}

func NewServer() *Server {
	router := chi.NewRouter()

	// authHandler := NewAuthHandler(authService)

	// router.Route("/auth", func(r chi.Router) {
	// 	r.Post("/register", authHandler.Register)
	// 	r.Post("/login", authHandler.Login)
	// })

	return &Server{
		router: router,
	}
}

func (s *Server) Handler() http.Handler {
	return s.router
}
