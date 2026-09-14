package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	m9StandardWorkloadID = "m9-standard-v1"

	m9StandardPromptTokens    = 512
	m9StandardGeneratedTokens = 128
	m9StandardRepetitions     = 5
	m9StandardContextSize     = 4096

	// Each llama-bench process represents exactly one M9 repetition.
	// The benchmark runner performs m9StandardRepetitions separate processes.
	m9LlamaBenchInvocationRepetitions = 1
)

type llamaBenchJSONResult struct {
	BuildCommit   string    `json:"build_commit"`
	BuildNumber   int       `json:"build_number"`
	Backends      string    `json:"backends"`
	ModelFilename string    `json:"model_filename"`
	ModelType     string    `json:"model_type"`
	NPrompt       int       `json:"n_prompt"`
	NGen          int       `json:"n_gen"`
	AvgTS         float64   `json:"avg_ts"`
	SamplesTS     []float64 `json:"samples_ts"`
}

type llamaBenchThroughput struct {
	RuntimeVersion string
	ModelFilename  string
	ModelType      string
	Backend        string

	PromptTokens    int
	GeneratedTokens int

	PromptTokensPerSecond     []float64
	GenerationTokensPerSecond []float64
}

// m9LlamaBenchArgs defines one standardized M9 throughput repetition.
//
// The complete M9 benchmark executes this command five separate times so that
// each throughput measurement can be paired with its own TTFT and resource
// telemetry. TTFT and resource telemetry are collected separately.
func m9LlamaBenchArgs(
	modelPath string,
) ([]string, error) {
	modelPath = strings.TrimSpace(modelPath)

	if modelPath == "" {
		return nil, errors.New(
			"llama-bench model path is required",
		)
	}

	return []string{
		"-m", modelPath,
		"-p", fmt.Sprintf("%d", m9StandardPromptTokens),
		"-n", fmt.Sprintf("%d", m9StandardGeneratedTokens),
		"-r", fmt.Sprintf("%d", m9LlamaBenchInvocationRepetitions),
		"-o", "json",
	}, nil
}

func parseLlamaBenchJSON(
	raw []byte,
) (llamaBenchThroughput, error) {
	var result llamaBenchThroughput

	if strings.TrimSpace(string(raw)) == "" {
		return result, errors.New(
			"llama-bench output is empty",
		)
	}

	var rows []llamaBenchJSONResult

	if err := json.Unmarshal(raw, &rows); err != nil {
		return result, errors.New(
			"invalid llama-bench JSON",
		)
	}

	var (
		promptRow     *llamaBenchJSONResult
		generationRow *llamaBenchJSONResult
	)

	for i := range rows {
		row := &rows[i]

		switch {
		case row.NPrompt == m9StandardPromptTokens &&
			row.NGen == 0:
			if promptRow != nil {
				return result, errors.New(
					"duplicate M9 prompt result",
				)
			}
			promptRow = row

		case row.NPrompt == 0 &&
			row.NGen == m9StandardGeneratedTokens:
			if generationRow != nil {
				return result, errors.New(
					"duplicate M9 generation result",
				)
			}
			generationRow = row
		}
	}

	if promptRow == nil {
		return result, errors.New(
			"M9 prompt result is missing",
		)
	}

	if generationRow == nil {
		return result, errors.New(
			"M9 generation result is missing",
		)
	}

	if promptRow.ModelFilename == "" ||
		promptRow.ModelFilename != generationRow.ModelFilename {
		return result, errors.New(
			"llama-bench model mismatch",
		)
	}

	if promptRow.ModelType == "" ||
		promptRow.ModelType != generationRow.ModelType {
		return result, errors.New(
			"llama-bench model type mismatch",
		)
	}

	if promptRow.BuildCommit == "" ||
		promptRow.BuildCommit != generationRow.BuildCommit ||
		promptRow.BuildNumber != generationRow.BuildNumber {
		return result, errors.New(
			"llama-bench runtime version mismatch",
		)
	}

	if promptRow.Backends == "" ||
		promptRow.Backends != generationRow.Backends {
		return result, errors.New(
			"llama-bench backend mismatch",
		)
	}

	if err := validateLlamaBenchSamples(
		promptRow.SamplesTS,
	); err != nil {
		return result, fmt.Errorf(
			"prompt throughput: %w",
			err,
		)
	}

	if err := validateLlamaBenchSamples(
		generationRow.SamplesTS,
	); err != nil {
		return result, fmt.Errorf(
			"generation throughput: %w",
			err,
		)
	}

	if len(promptRow.SamplesTS) !=
		len(generationRow.SamplesTS) {
		return result, errors.New(
			"llama-bench repetition counts do not match",
		)
	}

	if len(promptRow.SamplesTS) !=
		m9LlamaBenchInvocationRepetitions {
		return result, fmt.Errorf(
			"llama-bench invocation repetition count = %d, want %d",
			len(promptRow.SamplesTS),
			m9LlamaBenchInvocationRepetitions,
		)
	}

	result = llamaBenchThroughput{
		RuntimeVersion: fmt.Sprintf(
			"%s-%d",
			promptRow.BuildCommit,
			promptRow.BuildNumber,
		),
		ModelFilename: promptRow.ModelFilename,
		ModelType:     promptRow.ModelType,
		Backend:       promptRow.Backends,

		PromptTokens:    m9StandardPromptTokens,
		GeneratedTokens: m9StandardGeneratedTokens,

		PromptTokensPerSecond: append([]float64(nil), promptRow.SamplesTS...),

		GenerationTokensPerSecond: append([]float64(nil), generationRow.SamplesTS...),
	}

	return result, nil
}

func validateLlamaBenchSamples(
	values []float64,
) error {
	if len(values) == 0 {
		return errors.New(
			"llama-bench samples are empty",
		)
	}

	for _, value := range values {
		if !finiteNonNegative(value) ||
			value <= 0 {
			return errors.New(
				"llama-bench sample is invalid",
			)
		}
	}

	return nil
}
