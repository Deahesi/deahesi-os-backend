package llm

import (
	"context"

	"github.com/Deahesi/deahesi-os-backend/internal/domain"
	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/models/operations"
)

type LLMService struct {
	systemMessage components.ChatSystemMessage
	model         string
	sdk           *openrouter.OpenRouter
}

func NewLLMService(Sdk *openrouter.OpenRouter, systemMessage components.ChatSystemMessage, model string) *LLMService {
	return &LLMService{
		sdk:           Sdk,
		systemMessage: systemMessage,
		model:         model,
	}
}

func (s *LLMService) SendMessage(ctx context.Context, message *domain.ChatUserMessage) (*operations.SendChatCompletionRequestResponse, error) {
	content := []components.ChatContentItems{
		components.CreateChatContentItemsText(
			components.ChatContentText{Text: message.Text},
		),
	}

	if message.FileData != nil && *message.FileData != "" {
		content = append(content, components.CreateChatContentItemsFile(
			components.ChatContentFile{
				File: components.ChatContentFileFile{
					Filename: message.Filename,
					FileData: message.FileData,
				},
			},
		))
	}

	res, err := s.sdk.Chat.Send(ctx, components.ChatRequest{
		Model: openrouter.String(s.model),
		Messages: []components.ChatMessages{
			components.CreateChatMessagesSystem(
				s.systemMessage,
			),
			components.CreateChatMessagesUser(
				components.ChatUserMessage{
					Content: components.CreateChatUserMessageContentArrayOfChatContentItems(
						content,
					),
				},
			),
		},
		Tools:  tools,
		Stream: new(true),
	}, nil)

	if err != nil {
		return nil, err
	}

	return res, nil
}
