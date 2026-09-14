package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type computeBenchmarkHTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type llamaTTFTRequest struct {
	Prompt       string  `json:"prompt"`
	NPredict     int     `json:"n_predict"`
	Stream       bool    `json:"stream"`
	CachePrompt  bool    `json:"cache_prompt"`
	ReturnTokens bool    `json:"return_tokens"`
	Seed         int     `json:"seed"`
	Temperature  float64 `json:"temperature"`
}

type llamaTTFTStreamEvent struct {
	Content string `json:"content"`
	Tokens  []int  `json:"tokens"`
	Stop    bool   `json:"stop"`
}

// measureLlamaServerTTFT measures request submission to arrival of the first
// generated token from a benchmark-only local llama.cpp server.
//
// M9 deliberately restricts this helper to loopback addresses. It is not a
// general remote runtime adapter.
func measureLlamaServerTTFT(
	ctx context.Context,
	client computeBenchmarkHTTPDoer,
	baseURL string,
	prompt string,
	now func() time.Time,
) (time.Duration, error) {
	if client == nil {
		return 0, errors.New(
			"TTFT HTTP client is nil",
		)
	}

	if now == nil {
		return 0, errors.New(
			"TTFT clock is nil",
		)
	}

	if strings.TrimSpace(prompt) == "" {
		return 0, errors.New(
			"TTFT prompt is empty",
		)
	}

	baseURL, err := validateLocalBenchmarkBaseURL(
		baseURL,
	)
	if err != nil {
		return 0, err
	}

	payload := llamaTTFTRequest{
		Prompt:       prompt,
		NPredict:     1,
		Stream:       true,
		CachePrompt:  false,
		ReturnTokens: true,
		Seed:         1,
		Temperature:  0,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		baseURL+"/completion",
		bytes.NewReader(body),
	)
	if err != nil {
		return 0, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	startedAt := now()

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf(
			"llama.cpp TTFT request failed: %w",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 ||
		resp.StatusCode > 299 {
		return 0, fmt.Errorf(
			"llama.cpp TTFT HTTP status %d",
			resp.StatusCode,
		)
	}

	if err := waitForFirstLlamaToken(
		resp.Body,
	); err != nil {
		return 0, err
	}

	finishedAt := now()

	duration := finishedAt.Sub(startedAt)

	if duration <= 0 {
		return 0, errors.New(
			"TTFT duration is invalid",
		)
	}

	return duration, nil
}

func waitForFirstLlamaToken(
	reader io.Reader,
) error {
	scanner := bufio.NewScanner(reader)

	scanner.Buffer(
		make([]byte, 64*1024),
		1024*1024,
	)

	for scanner.Scan() {
		line := strings.TrimSpace(
			scanner.Text(),
		)

		if !strings.HasPrefix(
			line,
			"data:",
		) {
			continue
		}

		data := strings.TrimSpace(
			strings.TrimPrefix(
				line,
				"data:",
			),
		)

		if data == "" ||
			data == "[DONE]" {
			continue
		}

		var event llamaTTFTStreamEvent

		if err := json.Unmarshal(
			[]byte(data),
			&event,
		); err != nil {
			return errors.New(
				"invalid llama.cpp streaming event",
			)
		}

		if len(event.Tokens) > 0 ||
			event.Content != "" {
			return nil
		}

		if event.Stop {
			return errors.New(
				"llama.cpp stream stopped before first token",
			)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf(
			"read llama.cpp stream: %w",
			err,
		)
	}

	return errors.New(
		"llama.cpp stream ended before first token",
	)
}

func validateLocalBenchmarkBaseURL(
	value string,
) (string, error) {
	value = strings.TrimSpace(value)

	parsed, err := url.Parse(value)
	if err != nil {
		return "", errors.New(
			"invalid benchmark server URL",
		)
	}

	if parsed.Scheme != "http" ||
		parsed.Host == "" ||
		parsed.User != nil ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" ||
		(parsed.Path != "" &&
			parsed.Path != "/") {
		return "", errors.New(
			"benchmark server must be a local HTTP base URL",
		)
	}

	host := strings.ToLower(
		parsed.Hostname(),
	)

	switch host {
	case "127.0.0.1",
		"localhost",
		"::1":
	default:
		return "", errors.New(
			"benchmark server must use a loopback address",
		)
	}

	return strings.TrimRight(
		value,
		"/",
	), nil
}
