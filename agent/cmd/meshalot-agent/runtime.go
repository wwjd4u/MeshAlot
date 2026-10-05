package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/wwjd4u/MeshAlot/agent"
)

const defaultRuntimeURL = "http://127.0.0.1:8080"

func runRuntime(args []string) error {
	return runRuntimeWithOutput(args, io.Discard)
}

func runRuntimeWithOutput(args []string, output io.Writer) error {
	if len(args) < 1 {
		return errors.New("runtime action is required: health, models, or complete")
	}

	switch args[0] {
	case "health":
		return runRuntimeHealth(args[1:], output)
	case "models":
		return runRuntimeModels(args[1:], output)
	case "complete":
		return runRuntimeComplete(args[1:], output)
	default:
		return fmt.Errorf("unsupported runtime action %q", args[0])
	}
}

func runRuntimeHealth(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("runtime health", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	runtimeURL := flags.String("url", defaultRuntimeURL, "localhost llama.cpp base URL")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("runtime health accepts no positional arguments")
	}

	adapter, err := newRuntimeAdapter(*runtimeURL)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status, err := adapter.Health(ctx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "runtime=%s health=%s\n", adapter.Kind(), status)
	return err
}

func runRuntimeModels(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("runtime models", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	runtimeURL := flags.String("url", defaultRuntimeURL, "localhost llama.cpp base URL")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("runtime models accepts no positional arguments")
	}

	adapter, err := newRuntimeAdapter(*runtimeURL)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	capabilities, err := adapter.Capabilities(ctx)
	if err != nil {
		return err
	}
	if len(capabilities.Models) == 0 {
		_, err = fmt.Fprintln(output, "no models detected")
		return err
	}

	for _, model := range capabilities.Models {
		operations := make([]string, 0, len(model.Operations))
		for _, operation := range model.Operations {
			operations = append(operations, string(operation))
		}
		if _, err := fmt.Fprintf(
			output,
			"model=%s context_tokens=%d operations=%s\n",
			model.ID,
			model.ContextWindowTokens,
			strings.Join(operations, ","),
		); err != nil {
			return err
		}
	}
	return nil
}

func runRuntimeComplete(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("runtime complete", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	runtimeURL := flags.String("url", defaultRuntimeURL, "localhost llama.cpp base URL")
	model := flags.String("model", "", "runtime model identifier")
	prompt := flags.String("prompt", "", "completion prompt")
	maxTokens := flags.Int("max-tokens", 32, "maximum generated tokens")
	temperature := flags.Float64("temperature", 0, "sampling temperature")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("runtime complete accepts no positional arguments")
	}

	adapter, err := newRuntimeAdapter(*runtimeURL)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	response, err := adapter.Complete(
		ctx,
		agent.CompletionRequest{
			Model:       *model,
			Prompt:      *prompt,
			MaxTokens:   *maxTokens,
			Temperature: *temperature,
		},
	)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(output, response.Text); err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		output,
		"prompt_tokens=%d completion_tokens=%d\n",
		response.PromptTokens,
		response.CompletionTokens,
	)
	return err
}

func newRuntimeAdapter(runtimeURL string) (*agent.LlamaCppAdapter, error) {
	return agent.NewLlamaCppAdapter(
		agent.RuntimeConfig{
			Kind:    agent.RuntimeKindLlamaCpp,
			BaseURL: runtimeURL,
		},
	)
}
