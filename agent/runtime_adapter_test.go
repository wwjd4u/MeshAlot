package agent

import "testing"

func TestValidateRuntimeConfigAcceptsLlamaCppLoopback(t *testing.T) {
	configs := []RuntimeConfig{
		{Kind: RuntimeKindLlamaCpp, BaseURL: "http://127.0.0.1:8080"},
		{Kind: RuntimeKindLlamaCpp, BaseURL: "http://localhost:8080/"},
		{Kind: RuntimeKindLlamaCpp, BaseURL: "http://[::1]:8080"},
	}

	for _, config := range configs {
		if err := ValidateRuntimeConfig(config); err != nil {
			t.Fatalf("ValidateRuntimeConfig(%+v) returned error: %v", config, err)
		}
	}
}

func TestValidateRuntimeConfigRejectsUnsupportedRuntime(t *testing.T) {
	err := ValidateRuntimeConfig(RuntimeConfig{
		Kind:    RuntimeKind("ollama"),
		BaseURL: "http://127.0.0.1:8080",
	})
	if err == nil {
		t.Fatal("expected unsupported runtime kind to be rejected")
	}
}

func TestValidateLoopbackRuntimeURLRejectsNonLoopbackAndUnsafeForms(t *testing.T) {
	cases := []string{
		"http://0.0.0.0:8080",
		"http://192.168.1.50:8080",
		"http://10.1.2.3:8080",
		"http://example.com:8080",
		"https://127.0.0.1:8080",
		"http://user:pass@127.0.0.1:8080",
		"http://127.0.0.1:8080/v1",
		"http://127.0.0.1:8080?x=1",
		"http://127.0.0.1:8080#fragment",
		"http://127.0.0.1",
		"http://127.0.0.1:0",
		"http://127.0.0.1:70000",
		"",
	}

	for _, raw := range cases {
		if err := ValidateLoopbackRuntimeURL(raw); err == nil {
			t.Fatalf("expected %q to be rejected", raw)
		}
	}
}

func TestPermittedRuntimeOperationsAreExplicit(t *testing.T) {
	if !IsPermittedRuntimeOperation(RuntimeOperationCompletion) {
		t.Fatal("completion should be permitted")
	}
	for _, operation := range []RuntimeOperation{
		"shell",
		"command",
		"process",
		"exec",
		RuntimeOperation(""),
	} {
		if IsPermittedRuntimeOperation(operation) {
			t.Fatalf("operation %q must not be permitted", operation)
		}
	}
}

func TestValidateCompletionRequest(t *testing.T) {
	valid := CompletionRequest{
		Model:       "qwen3",
		Prompt:      "hello",
		MaxTokens:   64,
		Temperature: 0.7,
	}
	if err := ValidateCompletionRequest(valid); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}

	invalid := []CompletionRequest{
		{Prompt: "hello", MaxTokens: 64, Temperature: 0.7},
		{Model: "qwen3", MaxTokens: 64, Temperature: 0.7},
		{Model: "qwen3", Prompt: "hello", MaxTokens: 0, Temperature: 0.7},
		{Model: "qwen3", Prompt: "hello", MaxTokens: 64, Temperature: -0.1},
		{Model: "qwen3", Prompt: "hello", MaxTokens: 64, Temperature: 2.1},
	}

	for i, request := range invalid {
		if err := ValidateCompletionRequest(request); err == nil {
			t.Fatalf("invalid request %d was accepted: %+v", i, request)
		}
	}
}
