package app

import (
	"context"
	"fmt"
)

type Application struct {
	server *Server
}

func NewApplication(ctx context.Context) (*Application, error) {
	server, err := NewServer(ctx)
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
