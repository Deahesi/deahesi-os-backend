package ai

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"strings"

	"github.com/Deahesi/deahesi-os-backend/internal/config"
	"github.com/Deahesi/deahesi-os-backend/internal/domain"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/openrouter"
	"github.com/openai/openai-go/option"
)

type AIService struct {
	genkit       *genkit.Genkit
	systemPrompt string
	flow         *core.Flow[SendMessageInput, string, domain.ChatChunk]
}

func NewAIService(ctx context.Context, cfg config.LLMConfig) (*AIService, error) {
	if strings.TrimSpace(cfg.OpenRouterAPI) == "" {
		return nil, errors.New("OPENROUTER_API_KEY or OPEN_ROUTER_API_KEY is required")
	}
	model := strings.TrimPrefix(strings.TrimSpace(cfg.Model), "openrouter/")
	if model == "" {
		return nil, errors.New("LLM_MODEL is required")
	}

	g := genkit.Init(ctx, genkit.WithPlugins(&openrouter.OpenRouter{
		APIKey: cfg.OpenRouterAPI,
		Opts:   []option.RequestOption{option.WithMiddleware(fileMiddleware), option.WithMaxRetries(0)},
	}), genkit.WithDefaultModel("openrouter/"+model))

	s := &AIService{
		genkit:       g,
		systemPrompt: cfg.SystemPrompt,
	}

	s.flow = genkit.DefineStreamingFlow(g, "sendMessage", s.generate)
	return s, nil
}

func (s *AIService) SendMessage(ctx context.Context, input *SendMessageInput, onChunk func(context.Context, domain.ChatChunk) error) error {
	_, err := s.flow.Action.Run(ctx, *input, onChunk)
	return err
}

func (s *AIService) generate(ctx context.Context, input SendMessageInput, send func(context.Context, domain.ChatChunk) error) (string, error) {
	parts := []*ai.Part{ai.NewTextPart(input.Text)}
	if input.File != nil {
		mediaType, _, err := mime.ParseMediaType(input.File.ContentType)
		if err != nil {
			return "", domain.ErrInvalidMimeType
		}
		file := attachment{input.File.Filename, "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(input.File.Data)}
		parts = append(parts, ai.NewMediaPart(mediaType, file.DataURL))
		ctx = context.WithValue(ctx, fileContextKey{}, file)
	}
	result, err := genkit.Generate(ctx, s.genkit,
		ai.WithSystem(s.systemPrompt),
		ai.WithPromptParts(parts...),
		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			return send(ctx, domain.ChatChunk{Text: chunk.Text(), Reasoning: chunk.Reasoning()})
		}),
	)
	if err != nil && !(result != nil && errors.Is(err, ai.ErrGenerationBlocked)) {
		return "", fmt.Errorf("generate message: %w", err)
	}
	return result.Text(), send(ctx, domain.ChatChunk{FinishReason: string(result.FinishReason)})
}
