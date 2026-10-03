package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Deahesi/portfolio-backend/internal/httpapi"
)

type Server struct {
	server *http.Server
}

func NewServer(ctx context.Context) (*Server, error) {
	api := httpapi.NewServer()

	server := &http.Server{
		Addr:    ":8080",
		Handler: api.Handler(),
	}

	return &Server{
		server: server,
	}, nil
}

func (server *Server) Serve(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- server.server.ListenAndServe()
	}()

	fmt.Printf("Server serve on %s\n", server.server.Addr)

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, close := context.WithTimeout(context.Background(), 5000)
	defer close()

	if err := server.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	return nil
}

func (server *Server) Close() {
	server.server.Close()
}
