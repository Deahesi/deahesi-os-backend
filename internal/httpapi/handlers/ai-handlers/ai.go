package aihandlers

import (
	"errors"
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

	return &AIHandler{
		ai: ai,
	}
}

func (h *AIHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		responses.WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"Internal Error",
			nil,
		)
		return
	}

	text := r.FormValue("text")
	if text == "" {
		responses.WriteError(
			w,
			http.StatusInternalServerError,
			"invalid text",
			"Invalid text",
			nil,
		)
		return
	}

	file, err := parsers.ParseFileFromRequest(r)
	if err != nil {
		responses.WriteError(
			w,
			http.StatusBadRequest,
			"invalid file",
			"Invalid file",
			err,
		)
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

	stream, err := h.ai.SendMessage(
		r.Context(),
		&ai.SendMessageInput{
			Text: text,
			File: fileInput,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidMimeType):
			responses.WriteError(
				w,
				http.StatusBadRequest,
				"invalid mime type",
				"Invalid mime type",
				err,
			)
			return
		case errors.Is(err, domain.ErrNoStream):
			responses.WriteError(
				w,
				http.StatusInternalServerError,
				"internal_error",
				"Internal Error",
				err,
			)
			return
		default:
			responses.WriteError(
				w,
				http.StatusInternalServerError,
				"internal_error",
				"Internal Error",
				err,
			)
			return
		}
	}

	responses.WriteStreamHeaders(w)

	defer stream.EventStream.Close()
	for stream.EventStream.Next() {
		event := stream.EventStream.Value()
		if len(event.Data.Choices) == 0 {
			continue
		}

		delta := event.Data.Choices[0]

		responses.WriteStreamJSON(w, flusher, delta)
	}
}
