package app

import (
	"context"
	"fmt"

	"github.com/Deahesi/deahesi-os-backend/internal/ai"
	"github.com/Deahesi/deahesi-os-backend/internal/config"
)

type Services struct {
	aiService *ai.AIService
}

func NewServices(ctx context.Context, cfg *config.Config) (*Services, error) {
	aiService, err := ai.NewAIService(ctx, cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("initialize Genkit: %w", err)
	}
	return &Services{aiService: aiService}, nil
}
