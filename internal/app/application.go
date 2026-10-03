package app

import (
	"context"
	"fmt"

	"github.com/Deahesi/deahesi-os-backend/internal/config"
)

type Application struct {
	server *Server
}

func NewApplication() (*Application, error) {
	cfg := config.NewConfig()
	services := NewServices(cfg)

	server, err := NewServer(
		services.aiService,
		services.llmService,
	)

	if err != nil {
		return nil, fmt.Errorf("server: %w", err)
	}

	return &Application{
		server: server,
	}, nil
}

func (a *Application) Run(ctx context.Context) error {
	return a.server.Serve(ctx)
}

func (a *Application) Close() {
	a.server.server.Close()
}
