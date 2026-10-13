package domain

import "errors"

var ErrInvalidMimeType = errors.New("invalid mime type")

type ChatChunk struct {
	Text         string `json:"text,omitempty"`
	Reasoning    string `json:"reasoning,omitempty"`
	FinishReason string `json:"finishReason,omitempty"`
}
