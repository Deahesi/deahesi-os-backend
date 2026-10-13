package ai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/openai/openai-go/option"
)

type fileContextKey struct{}

type attachment struct {
	Filename string `json:"filename"`
	DataURL  string `json:"file_data"`
}

// compat_oai v1.13.1 sends all media as image_url; restore OpenRouter file uploads.
func fileMiddleware(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
	file, ok := req.Context().Value(fileContextKey{}).(attachment)
	if !ok || req.Body == nil {
		return next(req)
	}
	var body map[string]any
	decoder := json.NewDecoder(req.Body)
	decoder.UseNumber()
	err := decoder.Decode(&body)
	_ = req.Body.Close()
	if err != nil {
		return nil, err
	}
	messages, _ := body["messages"].([]any)
	for _, value := range messages {
		message, _ := value.(map[string]any)
		parts, _ := message["content"].([]any)
		for i, value := range parts {
			part, _ := value.(map[string]any)
			image, _ := part["image_url"].(map[string]any)
			if image["url"] == file.DataURL {
				parts[i] = map[string]any{"type": "file", "file": file}
			}
		}
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(data)), nil }
	req.Body, _ = req.GetBody()
	req.ContentLength = int64(len(data))
	return next(req)
}
