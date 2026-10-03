package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Deahesi/portfolio-backend/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// cfg, err := config.Load()
	// if err != nil {
	// 	log.Fatalf("config: %v", err)
	// }

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func run(ctx context.Context) error {
	application, err := app.NewApplication(ctx)
	if err != nil {
		return err
	}

	defer application.Close()
	return application.Run(ctx)
}
