package responses

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
	err error,
) {
	log.Println(err)
	WriteJSON(
		w,
		status,
		ErrorResponse{
			Code:    code,
			Message: message,
		},
	)
}

func WriteStreamHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Заголовок для обхода буферизации в Nginx/Cloudflare
	w.Header().Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(http.StatusOK)
}

func WriteStreamJSON(w http.ResponseWriter, flusher http.Flusher, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "data: %s\n\n", jsonData)
	if err != nil {
		return err
	}

	flusher.Flush()
	return nil
}

func WriteStreamError(w http.ResponseWriter, flusher http.Flusher, code string, message string) {
	_ = WriteStreamJSON(w, flusher, ErrorResponse{
		Code:    code,
		Message: message,
	})
}
