package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"

	"github.com/Deahesi/deahesi-os-backend/internal/domain"
	"github.com/Deahesi/deahesi-os-backend/internal/llm"
	"github.com/OpenRouterTeam/go-sdk/models/operations"
)

type AIService struct {
	llm *llm.LLMService
}

func NewAIService(LLM *llm.LLMService) *AIService {

	return &AIService{
		llm: LLM,
	}
}

func (s *AIService) SendMessage(ctx context.Context, input *SendMessageInput) (*operations.SendChatCompletionRequestResponse, error) {
	var (
		fileDataUrl string
		fileName    string
	)

	if input.File != nil {
		mediaType, _, err := mime.ParseMediaType(input.File.ContentType)
		if err != nil {
			return nil, domain.ErrInvalidMimeType
		}

		fileDataUrl = fmt.Sprintf("data:%s;base64,%s", mediaType, base64.StdEncoding.EncodeToString(input.File.Data))
		fileName = input.File.Filename
	}

	stream, err := s.llm.SendMessage(ctx, &domain.ChatUserMessage{
		Text:     input.Text,
		Filename: &fileName,
		FileData: &fileDataUrl,
	})

	if err != nil {
		return nil, err
	}

	if stream == nil {
		return nil, domain.ErrNoStream
	}

	return stream, nil
}
