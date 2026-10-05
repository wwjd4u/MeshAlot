package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultRuntimeHTTPTimeout = 30 * time.Second

// LlamaCppAdapter implements RuntimeAdapter against llama-server's documented
// loopback HTTP API.
type LlamaCppAdapter struct {
	baseURL    string
	httpClient *http.Client
}

// NewLlamaCppAdapter constructs a llama.cpp runtime adapter after enforcing the
// Milestone 12 loopback-only runtime boundary.
func NewLlamaCppAdapter(config RuntimeConfig) (*LlamaCppAdapter, error) {
	if err := ValidateRuntimeConfig(config); err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: defaultRuntimeHTTPTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return newLlamaCppAdapterWithClient(config, client)
}

func newLlamaCppAdapterWithClient(
	config RuntimeConfig,
	client *http.Client,
) (*LlamaCppAdapter, error) {
	if err := ValidateRuntimeConfig(config); err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("runtime HTTP client is required")
	}

	return &LlamaCppAdapter{
		baseURL:    strings.TrimRight(config.BaseURL, "/"),
		httpClient: client,
	}, nil
}

// Kind reports the concrete runtime implemented by this adapter.
func (a *LlamaCppAdapter) Kind() RuntimeKind {
	return RuntimeKindLlamaCpp
}

// Health queries llama-server's public health endpoint.
//
// A reachable runtime that is not ready yet is normalized as unhealthy without
// converting the health state itself into a transport error.
func (a *LlamaCppAdapter) Health(
	ctx context.Context,
) (RuntimeHealthStatus, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		a.baseURL+"/health",
		nil,
	)
	if err != nil {
		return RuntimeHealthUnknown, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return RuntimeHealthUnknown, fmt.Errorf("llama.cpp health request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RuntimeHealthUnhealthy, nil
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := decodeRuntimeJSON(resp.Body, &payload); err != nil {
		return RuntimeHealthUnknown, fmt.Errorf("decode llama.cpp health response: %w", err)
	}
	if payload.Status != "ok" {
		return RuntimeHealthUnhealthy, nil
	}

	return RuntimeHealthHealthy, nil
}

// Models returns models visible through llama-server's OpenAI-compatible model
// endpoint.
func (a *LlamaCppAdapter) Models(
	ctx context.Context,
) ([]RuntimeModel, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		a.baseURL+"/v1/models",
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llama.cpp models request: %w", err)
	}
	defer resp.Body.Close()

	if err := requireRuntimeStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var payload struct {
		Data []struct {
			ID   string `json:"id"`
			Meta *struct {
				ContextWindowTokens int `json:"n_ctx_train"`
			} `json:"meta"`
		} `json:"data"`
	}
	if err := decodeRuntimeJSON(resp.Body, &payload); err != nil {
		return nil, fmt.Errorf("decode llama.cpp models response: %w", err)
	}

	models := make([]RuntimeModel, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}

		contextWindow := 0
		if item.Meta != nil {
			contextWindow = item.Meta.ContextWindowTokens
		}

		models = append(models, RuntimeModel{
			ID:                  item.ID,
			ContextWindowTokens: contextWindow,
			Operations: []RuntimeOperation{
				RuntimeOperationCompletion,
			},
		})
	}

	return models, nil
}

// Capabilities reports the explicit operation allow-list together with the
// models currently visible from llama-server.
func (a *LlamaCppAdapter) Capabilities(
	ctx context.Context,
) (RuntimeCapabilities, error) {
	models, err := a.Models(ctx)
	if err != nil {
		return RuntimeCapabilities{}, err
	}

	return RuntimeCapabilities{
		Operations: []RuntimeOperation{
			RuntimeOperationCompletion,
		},
		Models: models,
	}, nil
}

// Complete sends a normalized non-streaming completion request through
// llama-server's OpenAI-compatible completions endpoint.
func (a *LlamaCppAdapter) Complete(
	ctx context.Context,
	request CompletionRequest,
) (CompletionResponse, error) {
	if err := ValidateCompletionRequest(request); err != nil {
		return CompletionResponse{}, err
	}

	body, err := json.Marshal(struct {
		Model       string  `json:"model"`
		Prompt      string  `json:"prompt"`
		MaxTokens   int     `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
		Stream      bool    `json:"stream"`
	}{
		Model:       request.Model,
		Prompt:      request.Prompt,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
		Stream:      false,
	})
	if err != nil {
		return CompletionResponse{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.baseURL+"/v1/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return CompletionResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("llama.cpp completion request: %w", err)
	}
	defer resp.Body.Close()

	if err := requireRuntimeStatus(resp, http.StatusOK); err != nil {
		return CompletionResponse{}, err
	}

	var payload struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := decodeRuntimeJSON(resp.Body, &payload); err != nil {
		return CompletionResponse{}, fmt.Errorf("decode llama.cpp completion response: %w", err)
	}
	if len(payload.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("llama.cpp completion response contained no choices")
	}

	return CompletionResponse{
		Text:             payload.Choices[0].Text,
		PromptTokens:     payload.Usage.PromptTokens,
		CompletionTokens: payload.Usage.CompletionTokens,
	}, nil
}

func requireRuntimeStatus(
	resp *http.Response,
	expected int,
) error {
	if resp.StatusCode == expected {
		return nil
	}

	const maxErrorBody = 4096
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = http.StatusText(resp.StatusCode)
	}

	return fmt.Errorf(
		"llama.cpp returned HTTP %d: %s",
		resp.StatusCode,
		message,
	)
}

func decodeRuntimeJSON(body io.Reader, target any) error {
	decoder := json.NewDecoder(io.LimitReader(body, 1<<20))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}
