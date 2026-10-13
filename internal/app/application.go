package app

import (
	"context"
	"fmt"

	"github.com/Deahesi/deahesi-os-backend/internal/config"
)

type Application struct {
	server *Server
}

func NewApplication(ctx context.Context) (*Application, error) {
	cfg := config.NewConfig()
	services, err := NewServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	server, err := NewServer(
		services.aiService,
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
