package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewLlamaCppAdapterRejectsNonLoopbackEndpoint(t *testing.T) {
	_, err := NewLlamaCppAdapter(RuntimeConfig{
		Kind:    RuntimeKindLlamaCpp,
		BaseURL: "http://192.168.1.25:8080",
	})
	if err == nil {
		t.Fatal("expected non-loopback runtime endpoint to be rejected")
	}
}

func TestLlamaCppHealth(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if r.URL.Path != "/health" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		adapter := testLlamaCppAdapter(t, server)

		status, err := adapter.Health(context.Background())
		if err != nil {
			t.Fatalf("Health returned error: %v", err)
		}
		if status != RuntimeHealthHealthy {
			t.Fatalf("status = %q, want %q", status, RuntimeHealthHealthy)
		}
	})

	t.Run("loading is unhealthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":{"message":"Loading model"}}`))
		}))
		defer server.Close()

		adapter := testLlamaCppAdapter(t, server)

		status, err := adapter.Health(context.Background())
		if err != nil {
			t.Fatalf("Health returned error: %v", err)
		}
		if status != RuntimeHealthUnhealthy {
			t.Fatalf("status = %q, want %q", status, RuntimeHealthUnhealthy)
		}
	})
}

func TestLlamaCppModelsAndCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"object":"list",
			"data":[
				{
					"id":"qwen3-30b.gguf",
					"meta":{"n_ctx_train":65536}
				}
			]
		}`))
	}))
	defer server.Close()

	adapter := testLlamaCppAdapter(t, server)

	models, err := adapter.Models(context.Background())
	if err != nil {
		t.Fatalf("Models returned error: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("len(models) = %d, want 1", len(models))
	}
	if models[0].ID != "qwen3-30b.gguf" {
		t.Fatalf("model ID = %q", models[0].ID)
	}
	if models[0].ContextWindowTokens != 65536 {
		t.Fatalf(
			"context window = %d, want 65536",
			models[0].ContextWindowTokens,
		)
	}
	if len(models[0].Operations) != 1 ||
		models[0].Operations[0] != RuntimeOperationCompletion {
		t.Fatalf("unexpected model operations: %+v", models[0].Operations)
	}

	capabilities, err := adapter.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("Capabilities returned error: %v", err)
	}
	if len(capabilities.Operations) != 1 ||
		capabilities.Operations[0] != RuntimeOperationCompletion {
		t.Fatalf("unexpected operations: %+v", capabilities.Operations)
	}
	if len(capabilities.Models) != 1 {
		t.Fatalf("len(capabilities.Models) = %d, want 1", len(capabilities.Models))
	}
}

func TestLlamaCppCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/completions" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q", got)
		}

		var payload struct {
			Model       string  `json:"model"`
			Prompt      string  `json:"prompt"`
			MaxTokens   int     `json:"max_tokens"`
			Temperature float64 `json:"temperature"`
			Stream      bool    `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Model != "qwen3" ||
			payload.Prompt != "hello" ||
			payload.MaxTokens != 32 ||
			payload.Temperature != 0.2 ||
			payload.Stream {
			t.Fatalf("unexpected completion payload: %+v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices":[{"text":" world"}],
			"usage":{"prompt_tokens":1,"completion_tokens":1}
		}`))
	}))
	defer server.Close()

	adapter := testLlamaCppAdapter(t, server)

	response, err := adapter.Complete(
		context.Background(),
		CompletionRequest{
			Model:       "qwen3",
			Prompt:      "hello",
			MaxTokens:   32,
			Temperature: 0.2,
		},
	)
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if response.Text != " world" {
		t.Fatalf("Text = %q", response.Text)
	}
	if response.PromptTokens != 1 || response.CompletionTokens != 1 {
		t.Fatalf("unexpected usage: %+v", response)
	}
}

func TestLlamaCppCompletionRejectsRuntimeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		http.Error(w, "model unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	adapter := testLlamaCppAdapter(t, server)

	_, err := adapter.Complete(
		context.Background(),
		CompletionRequest{
			Model:       "qwen3",
			Prompt:      "hello",
			MaxTokens:   32,
			Temperature: 0.2,
		},
	)
	if err == nil {
		t.Fatal("expected runtime error")
	}
	if !strings.Contains(err.Error(), "HTTP 503") {
		t.Fatalf("error = %q, want HTTP 503", err)
	}
}

func TestLlamaCppAdapterDoesNotFollowRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		http.Redirect(
			w,
			r,
			"http://example.com/",
			http.StatusTemporaryRedirect,
		)
	}))
	defer server.Close()

	adapter, err := NewLlamaCppAdapter(RuntimeConfig{
		Kind:    RuntimeKindLlamaCpp,
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("NewLlamaCppAdapter returned error: %v", err)
	}

	status, err := adapter.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}
	if status != RuntimeHealthUnhealthy {
		t.Fatalf("status = %q, want unhealthy", status)
	}
}

func testLlamaCppAdapter(
	t *testing.T,
	server *httptest.Server,
) *LlamaCppAdapter {
	t.Helper()

	adapter, err := newLlamaCppAdapterWithClient(
		RuntimeConfig{
			Kind:    RuntimeKindLlamaCpp,
			BaseURL: server.URL,
		},
		server.Client(),
	)
	if err != nil {
		t.Fatalf("newLlamaCppAdapterWithClient returned error: %v", err)
	}
	return adapter
}
