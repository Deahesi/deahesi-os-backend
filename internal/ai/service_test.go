package ai

import (
	"context"
	"testing"

	"github.com/Deahesi/deahesi-os-backend/internal/config"
)

func TestNewAIServiceRejectsMissingConfiguration(t *testing.T) {
	for _, cfg := range []config.LLMConfig{
		{Model: "test/model"},
		{OpenRouterAPI: "   ", Model: "test/model"},
		{OpenRouterAPI: "test-key"},
		{OpenRouterAPI: "test-key", Model: "openrouter/"},
	} {
		service, err := NewAIService(context.Background(), cfg)
		if err == nil || service != nil {
			t.Errorf("expected startup error, got service=%v err=%v", service, err)
		}
	}
}
