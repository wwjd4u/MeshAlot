package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type fakeTTFTDoer struct {
	statusCode int
	body       string

	requestURL  string
	requestBody []byte
}

func (f *fakeTTFTDoer) Do(
	req *http.Request,
) (*http.Response, error) {
	f.requestURL = req.URL.String()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	f.requestBody = body

	return &http.Response{
		StatusCode: f.statusCode,
		Body: io.NopCloser(
			strings.NewReader(f.body),
		),
		Header: make(http.Header),
	}, nil
}

func TestMeasureLlamaServerTTFT(
	t *testing.T,
) {
	stream := "" +
		"data: {\"prompt_progress\":{\"total\":512,\"processed\":512}}\n\n" +
		"data: {\"content\":\"A\",\"tokens\":[123],\"stop\":false}\n\n" +
		"data: {\"content\":\"\",\"tokens\":[],\"stop\":true}\n\n"

	client := &fakeTTFTDoer{
		statusCode: http.StatusOK,
		body:       stream,
	}

	start := time.Date(
		2026,
		9,
		14,
		22,
		0,
		0,
		0,
		time.UTC,
	)

	times := []time.Time{
		start,
		start.Add(125 * time.Millisecond),
	}

	index := 0

	clock := func() time.Time {
		value := times[index]

		if index < len(times)-1 {
			index++
		}

		return value
	}

	duration, err := measureLlamaServerTTFT(
		context.Background(),
		client,
		"http://127.0.0.1:8080",
		"MeshAlot M9 benchmark prompt",
		clock,
	)
	if err != nil {
		t.Fatal(err)
	}

	if duration != 125*time.Millisecond {
		t.Fatalf(
			"TTFT = %v, want 125ms",
			duration,
		)
	}

	if client.requestURL !=
		"http://127.0.0.1:8080/completion" {
		t.Fatalf(
			"request URL = %q",
			client.requestURL,
		)
	}

	var payload llamaTTFTRequest

	if err := json.Unmarshal(
		client.requestBody,
		&payload,
	); err != nil {
		t.Fatal(err)
	}

	if !payload.Stream ||
		payload.NPredict != 1 ||
		payload.CachePrompt ||
		!payload.ReturnTokens ||
		payload.Seed != 1 ||
		payload.Temperature != 0 {
		t.Fatalf(
			"unexpected TTFT request: %#v",
			payload,
		)
	}
}

func TestWaitForFirstLlamaTokenRejectsStop(
	t *testing.T,
) {
	stream :=
		"data: {\"content\":\"\",\"tokens\":[],\"stop\":true}\n\n"

	if err := waitForFirstLlamaToken(
		strings.NewReader(stream),
	); err == nil {
		t.Fatal(
			"stream stopping before a token was accepted",
		)
	}
}

func TestMeasureLlamaServerTTFTRejectsRemoteURL(
	t *testing.T,
) {
	client := &fakeTTFTDoer{
		statusCode: http.StatusOK,
	}

	_, err := measureLlamaServerTTFT(
		context.Background(),
		client,
		"https://example.com",
		"benchmark prompt",
		time.Now,
	)

	if err == nil {
		t.Fatal(
			"remote benchmark server URL was accepted",
		)
	}

	if client.requestURL != "" {
		t.Fatal(
			"remote benchmark URL was contacted",
		)
	}
}

func TestValidateLocalBenchmarkBaseURL(
	t *testing.T,
) {
	valid := []string{
		"http://127.0.0.1:8080",
		"http://localhost:8080/",
		"http://[::1]:8080",
	}

	for _, value := range valid {
		if _, err :=
			validateLocalBenchmarkBaseURL(
				value,
			); err != nil {
			t.Fatalf(
				"local URL %q rejected: %v",
				value,
				err,
			)
		}
	}

	invalid := []string{
		"https://127.0.0.1:8080",
		"http://192.168.1.10:8080",
		"http://example.com:8080",
		"http://127.0.0.1:8080/completion",
	}

	for _, value := range invalid {
		if _, err :=
			validateLocalBenchmarkBaseURL(
				value,
			); err == nil {
			t.Fatalf(
				"invalid URL %q accepted",
				value,
			)
		}
	}
}
