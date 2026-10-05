package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunRuntimeHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.URL.Path != "/health" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	var output bytes.Buffer
	if err := runRuntimeWithOutput(
		[]string{"health", "--url", server.URL},
		&output,
	); err != nil {
		t.Fatalf("runRuntimeWithOutput returned error: %v", err)
	}
	if got := output.String(); !strings.Contains(got, "health=healthy") {
		t.Fatalf("output = %q", got)
	}
}

func TestRunRuntimeModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		_, _ = w.Write([]byte(`{
			"data":[
				{"id":"qwen3","meta":{"n_ctx_train":65536}}
			]
		}`))
	}))
	defer server.Close()

	var output bytes.Buffer
	if err := runRuntimeWithOutput(
		[]string{"models", "--url", server.URL},
		&output,
	); err != nil {
		t.Fatalf("runRuntimeWithOutput returned error: %v", err)
	}
	got := output.String()
	if !strings.Contains(got, "model=qwen3") ||
		!strings.Contains(got, "context_tokens=65536") ||
		!strings.Contains(got, "operations=completion") {
		t.Fatalf("output = %q", got)
	}
}

func TestRunRuntimeComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.URL.Path != "/v1/completions" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"choices":[{"text":"mesh response"}],
			"usage":{"prompt_tokens":2,"completion_tokens":2}
		}`))
	}))
	defer server.Close()

	var output bytes.Buffer
	err := runRuntimeWithOutput(
		[]string{
			"complete",
			"--url", server.URL,
			"--model", "qwen3",
			"--prompt", "hello",
			"--max-tokens", "16",
			"--temperature", "0",
		},
		&output,
	)
	if err != nil {
		t.Fatalf("runRuntimeWithOutput returned error: %v", err)
	}

	got := output.String()
	if !strings.Contains(got, "mesh response") ||
		!strings.Contains(got, "prompt_tokens=2") ||
		!strings.Contains(got, "completion_tokens=2") {
		t.Fatalf("output = %q", got)
	}
}

func TestRunRuntimeRejectsRemoteURL(t *testing.T) {
	var output bytes.Buffer
	err := runRuntimeWithOutput(
		[]string{"health", "--url", "http://192.168.1.10:8080"},
		&output,
	)
	if err == nil {
		t.Fatal("expected remote runtime URL to be rejected")
	}
	if !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("error = %q", err)
	}
}
