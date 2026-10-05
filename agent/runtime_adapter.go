package agent

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// RuntimeKind identifies a local AI runtime implementation supported by the agent.
type RuntimeKind string

const (
	RuntimeKindLlamaCpp RuntimeKind = "llama.cpp"
)

// RuntimeOperation identifies an AI operation the marketplace may ask the
// local runtime adapter to perform.
//
// Milestone 12 intentionally starts with completion only. The contract exposes
// no generic command, process, shell, or arbitrary execution operation.
type RuntimeOperation string

const (
	RuntimeOperationCompletion RuntimeOperation = "completion"
)

// RuntimeHealthStatus is the agent's normalized view of local runtime health.
type RuntimeHealthStatus string

const (
	RuntimeHealthUnknown   RuntimeHealthStatus = "unknown"
	RuntimeHealthHealthy   RuntimeHealthStatus = "healthy"
	RuntimeHealthUnhealthy RuntimeHealthStatus = "unhealthy"
)

// RuntimeConfig describes one local runtime endpoint.
//
// BaseURL must refer only to loopback on the provider machine. The runtime is
// not permitted to be reached through a public or LAN address.
type RuntimeConfig struct {
	Kind    RuntimeKind
	BaseURL string
}

// RuntimeModel describes a model discovered from the local runtime.
type RuntimeModel struct {
	ID                  string
	ContextWindowTokens int
	Operations          []RuntimeOperation
}

// RuntimeCapabilities describes the operations and models currently available
// from the local runtime.
type RuntimeCapabilities struct {
	Operations []RuntimeOperation
	Models     []RuntimeModel
}

// CompletionRequest is the only marketplace inference request permitted by the
// initial Milestone 12 runtime contract.
type CompletionRequest struct {
	Model       string
	Prompt      string
	MaxTokens   int
	Temperature float64
}

// CompletionResponse is the normalized result returned to the agent.
type CompletionResponse struct {
	Text             string
	PromptTokens     int
	CompletionTokens int
}

// RuntimeAdapter is the agent-side boundary around a supported local AI
// runtime. It deliberately exposes only explicit AI operations.
type RuntimeAdapter interface {
	Kind() RuntimeKind
	Health(context.Context) (RuntimeHealthStatus, error)
	Models(context.Context) ([]RuntimeModel, error)
	Capabilities(context.Context) (RuntimeCapabilities, error)
	Complete(context.Context, CompletionRequest) (CompletionResponse, error)
}

// ValidateRuntimeConfig enforces the Milestone 12 runtime boundary before an
// adapter is constructed.
func ValidateRuntimeConfig(config RuntimeConfig) error {
	if config.Kind != RuntimeKindLlamaCpp {
		return fmt.Errorf("unsupported runtime kind %q", config.Kind)
	}
	if err := ValidateLoopbackRuntimeURL(config.BaseURL); err != nil {
		return fmt.Errorf("invalid runtime base URL: %w", err)
	}
	return nil
}

// ValidateLoopbackRuntimeURL accepts only an explicit HTTP loopback endpoint
// with a concrete TCP port and no credentials, query, fragment, or subpath.
func ValidateLoopbackRuntimeURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("runtime URL is required")
	}

	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("parse runtime URL: %w", err)
	}
	if u.Scheme != "http" {
		return fmt.Errorf("runtime URL scheme must be http")
	}
	if u.User != nil {
		return fmt.Errorf("runtime URL must not contain credentials")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("runtime URL must not contain query or fragment")
	}
	if u.Path != "" && u.Path != "/" {
		return fmt.Errorf("runtime URL must not contain a subpath")
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("runtime URL host is required")
	}
	if !strings.EqualFold(host, "localhost") {
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return fmt.Errorf("runtime URL host must be loopback")
		}
	}

	port := u.Port()
	if port == "" {
		return fmt.Errorf("runtime URL must include an explicit port")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("runtime URL port must be between 1 and 65535")
	}

	return nil
}

// IsPermittedRuntimeOperation reports whether the operation is part of the
// marketplace's explicit Milestone 12 allow-list.
func IsPermittedRuntimeOperation(operation RuntimeOperation) bool {
	return operation == RuntimeOperationCompletion
}

// ValidateCompletionRequest checks the normalized completion request before it
// is sent to the local runtime.
func ValidateCompletionRequest(request CompletionRequest) error {
	if strings.TrimSpace(request.Model) == "" {
		return fmt.Errorf("model is required")
	}
	if strings.TrimSpace(request.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	if request.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be greater than zero")
	}
	if request.Temperature < 0 || request.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	return nil
}
