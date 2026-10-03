package app

import (
	"github.com/Deahesi/deahesi-os-backend/internal/ai"
	"github.com/Deahesi/deahesi-os-backend/internal/config"
	"github.com/Deahesi/deahesi-os-backend/internal/llm"
	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
)

type Services struct {
	aiService  *ai.AIService
	llmService *llm.LLMService
}

func NewServices(cfg *config.Config) *Services {
	sdk := openrouter.New(
		openrouter.WithSecurity(cfg.LLM.OpenRouterAPI),
	)

	llm := llm.NewLLMService(sdk, components.ChatSystemMessage{
		Content: components.CreateChatSystemMessageContentStr(cfg.LLM.SystemPrompt),
	}, cfg.LLM.Model)

	ai := ai.NewAIService(llm)

	return &Services{
		aiService:  ai,
		llmService: llm,
	}
}
