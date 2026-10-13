package httpapi_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/Deahesi/deahesi-os-backend/internal/ai"
	"github.com/Deahesi/deahesi-os-backend/internal/config"
	"github.com/Deahesi/deahesi-os-backend/internal/httpapi"
)

const prompt = "Explain {{this}} at 100% accuracy"

func newAPI(t *testing.T, upstream http.HandlerFunc, model string) http.Handler {
	t.Helper()
	provider := httptest.NewServer(upstream)
	t.Cleanup(provider.Close)
	t.Setenv("GENKIT_ENV", "prod")
	t.Setenv("OPENROUTER_BASE_URL", provider.URL+"/v1")
	service, err := ai.NewAIService(context.Background(), config.LLMConfig{
		OpenRouterAPI: "test-key",
		SystemPrompt:  "Keep {{braces}} and 100% literal",
		Model:         model,
	})
	if err != nil {
		t.Fatal(err)
	}
	return httpapi.NewServer(service).Handler()
}

func messageRequest(t *testing.T, text, filename, contentType string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	if err := form.WriteField("text", text); err != nil {
		t.Fatal(err)
	}
	if filename != "" {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
		header.Set("Content-Type", contentType)
		part, err := form.CreatePart(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte("file contents")); err != nil {
			t.Fatal(err)
		}
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/ai/message", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	return req
}

func writeChunk(w http.ResponseWriter, delta string, finish any) {
	chunk := map[string]any{
		"id": "chat-test", "object": "chat.completion.chunk", "created": 1, "model": "test/model",
		"choices": []any{map[string]any{"index": 0, "delta": json.RawMessage(delta), "finish_reason": finish}},
	}
	data, _ := json.Marshal(chunk)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.(http.Flusher).Flush()
}

func TestMessageThroughGenkit(t *testing.T) {
	for _, tc := range []struct {
		name, model, filename, mime, finish string
	}{
		{name: "text", model: "test/model", finish: "stop"},
		{name: "pdf", model: "openrouter/test/model", filename: "document.pdf", mime: "application/pdf; charset=binary", finish: "length"},
		{name: "image", model: "test/model", filename: "photo.png", mime: "image/png", finish: "stop"},
		{name: "content filter", model: "test/model", finish: "content_filter"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
					t.Errorf("unexpected provider request: %s", r.URL.Path)
				}
				var body struct {
					Model    string `json:"model"`
					Stream   bool   `json:"stream"`
					Messages []struct {
						Role    string          `json:"role"`
						Content json.RawMessage `json:"content"`
					} `json:"messages"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if body.Model != "test/model" || !body.Stream || len(body.Messages) != 2 {
					t.Errorf("unexpected model/stream/messages: %+v", body)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				var system string
				_ = json.Unmarshal(body.Messages[0].Content, &system)
				if body.Messages[0].Role != "system" || system != "Keep {{braces}} and 100% literal" {
					t.Errorf("system prompt changed: %q", system)
				}
				var parts []struct {
					Type string `json:"type"`
					Text string `json:"text"`
					File struct {
						Filename string `json:"filename"`
						Data     string `json:"file_data"`
					} `json:"file"`
				}
				if err := json.Unmarshal(body.Messages[1].Content, &parts); err != nil {
					t.Error(err)
				}
				if body.Messages[1].Role != "user" || len(parts) == 0 || parts[0].Text != prompt {
					t.Errorf("user prompt changed: %s", body.Messages[1].Content)
				}
				wantParts := 1
				if tc.filename != "" {
					wantParts = 2
					if len(parts) == 2 {
						mime := strings.Split(tc.mime, ";")[0]
						wantData := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString([]byte("file contents"))
						if parts[1].Type != "file" || parts[1].File.Filename != tc.filename || parts[1].File.Data != wantData {
							t.Errorf("attachment changed: %+v", parts[1])
						}
					}
				}
				if len(parts) != wantParts {
					t.Errorf("got %d parts, want %d", len(parts), wantParts)
				}
				w.Header().Set("Content-Type", "text/event-stream")
				writeChunk(w, `{"reasoning":"Thinking"}`, nil)
				writeChunk(w, `{"role":"assistant","content":"Hello "}`, nil)
				writeChunk(w, `{"content":"world"}`, nil)
				writeChunk(w, `{}`, tc.finish)
				fmt.Fprint(w, "data: [DONE]\n\n")
			}, tc.model)
			response := httptest.NewRecorder()
			api.ServeHTTP(response, messageRequest(t, prompt, tc.filename, tc.mime))
			if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "text/event-stream" {
				t.Fatalf("unexpected response: %d %s", response.Code, response.Body)
			}
			var text, reasoning, finish string
			events := strings.Split(strings.TrimSpace(response.Body.String()), "\n\n")
			for i, event := range events {
				var choice struct {
					Delta struct {
						Content, Reasoning, Role string
					} `json:"delta"`
					Finish *string `json:"finish_reason"`
					Index  int     `json:"index"`
				}
				if err := json.Unmarshal([]byte(strings.TrimPrefix(event, "data: ")), &choice); err != nil {
					t.Fatal(err)
				}
				if choice.Index != 0 || (i == 0 && choice.Delta.Role != "assistant") {
					t.Errorf("SSE choice format changed: %s", event)
				}
				text += choice.Delta.Content
				reasoning += choice.Delta.Reasoning
				if choice.Finish != nil {
					finish = *choice.Finish
				}
			}
			if text != "Hello world" || reasoning != "Thinking" || finish != tc.finish {
				t.Errorf("unexpected stream: text=%q reasoning=%q finish=%q", text, reasoning, finish)
			}
		})
	}
}

func TestMessageProviderErrors(t *testing.T) {
	for _, afterChunk := range []bool{false, true} {
		t.Run(fmt.Sprintf("afterChunk=%t", afterChunk), func(t *testing.T) {
			api := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
				if !afterChunk {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Fprint(w, `{"error":{"message":"secret provider detail","type":"authentication_error"}}`)
					return
				}
				w.Header().Set("Content-Type", "text/event-stream")
				writeChunk(w, `{"content":"Partial"}`, nil)
				fmt.Fprint(w, "data: {\"error\":{\"code\":502,\"message\":\"secret provider detail\"}}\n\n")
			}, "test/model")
			response := httptest.NewRecorder()
			api.ServeHTTP(response, messageRequest(t, prompt, "", ""))
			wantStatus := http.StatusInternalServerError
			if afterChunk {
				wantStatus = http.StatusOK
			}
			body := response.Body.String()
			if response.Code != wantStatus || !strings.Contains(body, `"code":"internal_error"`) || strings.Contains(body, "secret provider detail") {
				t.Errorf("unexpected error response: %d %s", response.Code, body)
			}
			if afterChunk && (!strings.Contains(body, "Partial") || !strings.Contains(body, "data: {\"code\"")) {
				t.Errorf("missing SSE error after partial response: %s", body)
			}
		})
	}
}

func TestMessageStreamsAndCancelsUpstream(t *testing.T) {
	canceled := make(chan struct{})
	api := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		writeChunk(w, `{"content":"First"}`, nil)
		<-r.Context().Done()
		close(canceled)
	}, "test/model")
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)
	req := messageRequest(t, prompt, "", "")
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(server.URL, "http://")
	req.RequestURI = ""
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	line, err := bufio.NewReader(response.Body).ReadString('\n')
	if err != nil || !strings.Contains(line, `"content":"First"`) {
		t.Fatalf("first chunk was not streamed: %q %v", line, err)
	}
	response.Body.Close()
	select {
	case <-canceled:
	case <-time.After(5 * time.Second):
		t.Fatal("upstream request was not canceled after client disconnect")
	}
}

func TestMessageInvalidInput(t *testing.T) {
	api := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("invalid input reached provider")
	}, "test/model")
	for _, req := range []*http.Request{
		messageRequest(t, "", "", ""),
		messageRequest(t, prompt, "document.pdf", "invalid mime type"),
	} {
		response := httptest.NewRecorder()
		api.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", response.Code, response.Body)
		}
	}
}

type brokenStreamWriter struct {
	*httptest.ResponseRecorder
	writes int
}

func (w *brokenStreamWriter) Write([]byte) (int, error) {
	w.writes++
	return 0, io.ErrClosedPipe
}

func TestMessageWriteFailureCancelsUpstream(t *testing.T) {
	canceled := make(chan struct{})
	api := newAPI(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		writeChunk(w, `{"content":"First"}`, nil)
		<-r.Context().Done()
		close(canceled)
	}, "test/model")
	response := &brokenStreamWriter{ResponseRecorder: httptest.NewRecorder()}
	api.ServeHTTP(response, messageRequest(t, prompt, "", ""))
	if response.writes != 1 {
		t.Errorf("continued writing after failure: %d writes", response.writes)
	}
	select {
	case <-canceled:
	case <-time.After(5 * time.Second):
		t.Fatal("upstream request was not canceled after write failure")
	}
}
