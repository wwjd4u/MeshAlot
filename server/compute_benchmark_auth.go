package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"math"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const (
	computeBenchmarkSignatureMaxSkew = 5 * time.Minute
	computeBenchmarkMaxSamples       = 50
	computeBenchmarkMaxContextSize   = 1048576
)

var (
	ErrInvalidComputeBenchmarkSubmission = errors.New("invalid compute benchmark submission")

	ErrComputeBenchmarkSignature = errors.New("invalid compute benchmark signature")

	ErrComputeBenchmarkTimestamp = errors.New("compute benchmark timestamp outside allowed window")
)

func verifyComputeBenchmarkSignature(
	publicKeyValue string,
	signatureValue string,
	body []byte,
) error {
	publicKeyBytes, err := base64.RawStdEncoding.DecodeString(
		strings.TrimSpace(publicKeyValue),
	)
	if err != nil || len(publicKeyBytes) != ed25519.PublicKeySize {
		return ErrComputeBenchmarkSignature
	}

	signatureBytes, err := base64.RawStdEncoding.DecodeString(
		strings.TrimSpace(signatureValue),
	)
	if err != nil || len(signatureBytes) != ed25519.SignatureSize {
		return ErrComputeBenchmarkSignature
	}

	if !ed25519.Verify(
		ed25519.PublicKey(publicKeyBytes),
		protocol.ComputeBenchmarkSigningMessage(body),
		signatureBytes,
	) {
		return ErrComputeBenchmarkSignature
	}

	return nil
}

func validateComputeBenchmarkSubmission(
	submission protocol.ComputeBenchmarkSubmission,
	now time.Time,
) error {
	if submission.Version !=
		protocol.ComputeBenchmarkSubmissionVersion ||
		!validUUIDv4(submission.NodeID) ||
		!validUUIDv4(submission.ReportID) {
		return ErrInvalidComputeBenchmarkSubmission
	}

	if submission.SignedAt.IsZero() {
		return ErrComputeBenchmarkTimestamp
	}

	signedSkew := now.UTC().Sub(submission.SignedAt.UTC())
	if signedSkew < 0 {
		signedSkew = -signedSkew
	}

	if signedSkew > computeBenchmarkSignatureMaxSkew {
		return ErrComputeBenchmarkTimestamp
	}

	benchmark := submission.Benchmark

	if benchmark.CollectedAt.IsZero() {
		return ErrComputeBenchmarkTimestamp
	}

	collectionSkew := submission.SignedAt.UTC().Sub(
		benchmark.CollectedAt.UTC(),
	)
	if collectionSkew < 0 {
		collectionSkew = -collectionSkew
	}

	if collectionSkew > computeBenchmarkSignatureMaxSkew {
		return ErrComputeBenchmarkTimestamp
	}

	if benchmark.SchemaVersion !=
		protocol.ComputeBenchmarkSchemaVersion {
		return ErrInvalidComputeBenchmarkSubmission
	}

	if !validComputeBenchmarkText(
		benchmark.WorkloadID,
		128,
	) ||
		!validComputeBenchmarkText(
			benchmark.Condition,
			64,
		) ||
		!validComputeBenchmarkText(
			benchmark.Runtime,
			128,
		) ||
		!validComputeBenchmarkText(
			benchmark.Model,
			512,
		) {
		return ErrInvalidComputeBenchmarkSubmission
	}

	if len(benchmark.RuntimeVersion) > 128 ||
		len(benchmark.Quantization) > 128 {
		return ErrInvalidComputeBenchmarkSubmission
	}

	if benchmark.ContextSize <= 0 ||
		benchmark.ContextSize > computeBenchmarkMaxContextSize {
		return ErrInvalidComputeBenchmarkSubmission
	}

	if len(benchmark.Samples) == 0 ||
		len(benchmark.Samples) > computeBenchmarkMaxSamples {
		return ErrInvalidComputeBenchmarkSubmission
	}

	seenRuns := make(map[int]struct{}, len(benchmark.Samples))

	for _, sample := range benchmark.Samples {
		if sample.Run <= 0 {
			return ErrInvalidComputeBenchmarkSubmission
		}

		if _, exists := seenRuns[sample.Run]; exists {
			return ErrInvalidComputeBenchmarkSubmission
		}
		seenRuns[sample.Run] = struct{}{}

		if sample.PromptTokens < 0 ||
			sample.GeneratedTokens < 0 {
			return ErrInvalidComputeBenchmarkSubmission
		}

		if !finiteComputeBenchmarkNonNegative(
			sample.PromptTokensPerSecond,
		) ||
			!finiteComputeBenchmarkNonNegative(
				sample.GenerationTokensPerSecond,
			) ||
			!finiteComputeBenchmarkNonNegative(
				sample.TimeToFirstTokenMs,
			) {
			return ErrInvalidComputeBenchmarkSubmission
		}

		if sample.TemperatureCelsius != nil {
			temperature := *sample.TemperatureCelsius

			if math.IsNaN(temperature) ||
				math.IsInf(temperature, 0) ||
				temperature < -100 ||
				temperature > 250 {
				return ErrInvalidComputeBenchmarkSubmission
			}
		}

		errorText := strings.TrimSpace(sample.Error)

		if sample.Success {
			if sample.PromptTokens <= 0 ||
				sample.GeneratedTokens <= 0 ||
				sample.PromptTokensPerSecond <= 0 ||
				sample.GenerationTokensPerSecond <= 0 ||
				sample.PeakSystemRAMBytes == 0 ||
				errorText != "" {
				return ErrInvalidComputeBenchmarkSubmission
			}
		} else {
			if errorText == "" || len(sample.Error) > 1024 {
				return ErrInvalidComputeBenchmarkSubmission
			}
		}
	}

	return nil
}

func validComputeBenchmarkText(
	value string,
	maxLength int,
) bool {
	value = strings.TrimSpace(value)

	return value != "" &&
		len(value) <= maxLength
}

func finiteComputeBenchmarkNonNegative(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0
}
