package aihandlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/Deahesi/deahesi-os-backend/internal/ai"
	"github.com/Deahesi/deahesi-os-backend/internal/domain"
	"github.com/Deahesi/deahesi-os-backend/internal/httpapi/parsers"
	"github.com/Deahesi/deahesi-os-backend/internal/httpapi/responses"
)

type AIHandler struct {
	ai *ai.AIService
}

func NewAIHandler(ai *ai.AIService) *AIHandler {
	return &AIHandler{ai: ai}
}

func (h *AIHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		responses.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal Error", nil)
		return
	}

	text := r.FormValue("text")
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	if text == "" {
		responses.WriteError(w, http.StatusBadRequest, "invalid text", "Invalid text", nil)
		return
	}

	file, err := parsers.ParseFileFromRequest(r)
	if err != nil {
		responses.WriteError(w, http.StatusBadRequest, "invalid file", "Invalid file", err)
		return
	}

	var fileInput *ai.FileInput
	if file != nil {
		fileInput = &ai.FileInput{
			Data:        file.Data,
			ContentType: file.ContentType,
			Filename:    file.Filename,
		}
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	started := false
	var writeErr error
	err = h.ai.SendMessage(ctx, &ai.SendMessageInput{Text: text, File: fileInput},
		func(_ context.Context, chunk domain.ChatChunk) error {
			if !started {
				responses.WriteStreamHeaders(w)
			}
			event := streamChoice{
				Delta: streamDelta{Content: chunk.Text, Reasoning: chunk.Reasoning},
			}
			if !started {
				event.Delta.Role = "assistant"
			}
			started = true
			if chunk.FinishReason != "" {
				reason := finishReason(chunk.FinishReason)
				event.FinishReason = &reason
			}
			writeErr = responses.WriteStreamJSON(w, flusher, event)
			if writeErr != nil {
				cancel()
			}
			return writeErr
		},
	)
	if err == nil || writeErr != nil || r.Context().Err() != nil {
		return
	}
	if started {
		log.Printf("AI stream failed: %v", err)
		responses.WriteStreamError(w, flusher, "internal_error", "Internal Error")
		return
	}
	if errors.Is(err, domain.ErrInvalidMimeType) {
		responses.WriteError(w, http.StatusBadRequest, "invalid mime type", "Invalid mime type", err)
		return
	}
	responses.WriteError(w, http.StatusInternalServerError, "internal_error", "Internal Error", err)
}
