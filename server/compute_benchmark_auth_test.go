package server

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

func TestComputeBenchmarkSignatureAndValidation(
	t *testing.T,
) {
	publicKey, privateKey, err :=
		ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(
		2026,
		9,
		14,
		20,
		0,
		0,
		0,
		time.UTC,
	)

	submission :=
		validComputeBenchmarkSubmissionFixture(now)

	body, err := json.Marshal(submission)
	if err != nil {
		t.Fatal(err)
	}

	signature := ed25519.Sign(
		privateKey,
		protocol.ComputeBenchmarkSigningMessage(body),
	)

	signatureValue :=
		base64.RawStdEncoding.EncodeToString(signature)

	publicKeyValue :=
		base64.RawStdEncoding.EncodeToString(publicKey)

	if err := verifyComputeBenchmarkSignature(
		publicKeyValue,
		signatureValue,
		body,
	); err != nil {
		t.Fatalf(
			"valid signature rejected: %v",
			err,
		)
	}

	if err := validateComputeBenchmarkSubmission(
		submission,
		now,
	); err != nil {
		t.Fatalf(
			"valid submission rejected: %v",
			err,
		)
	}

	changed := append([]byte(nil), body...)
	changed[len(changed)-1] ^= 1

	if !errors.Is(
		verifyComputeBenchmarkSignature(
			publicKeyValue,
			signatureValue,
			changed,
		),
		ErrComputeBenchmarkSignature,
	) {
		t.Fatal(
			"modified compute benchmark body was not rejected",
		)
	}
}

func TestComputeBenchmarkTimestampWindow(
	t *testing.T,
) {
	now := time.Date(
		2026,
		9,
		14,
		20,
		0,
		0,
		0,
		time.UTC,
	)

	old := validComputeBenchmarkSubmissionFixture(
		now.Add(-10 * time.Minute),
	)

	if !errors.Is(
		validateComputeBenchmarkSubmission(old, now),
		ErrComputeBenchmarkTimestamp,
	) {
		t.Fatal(
			"stale compute benchmark timestamp was not rejected",
		)
	}

	future := validComputeBenchmarkSubmissionFixture(
		now.Add(10 * time.Minute),
	)

	if !errors.Is(
		validateComputeBenchmarkSubmission(future, now),
		ErrComputeBenchmarkTimestamp,
	) {
		t.Fatal(
			"future compute benchmark timestamp was not rejected",
		)
	}

	collectionOld :=
		validComputeBenchmarkSubmissionFixture(now)

	collectionOld.Benchmark.CollectedAt =
		now.Add(-10 * time.Minute)

	if !errors.Is(
		validateComputeBenchmarkSubmission(
			collectionOld,
			now,
		),
		ErrComputeBenchmarkTimestamp,
	) {
		t.Fatal(
			"stale collection timestamp was not rejected",
		)
	}
}

func TestComputeBenchmarkValidationRejectsBadValues(
	t *testing.T,
) {
	now := time.Date(
		2026,
		9,
		14,
		20,
		0,
		0,
		0,
		time.UTC,
	)

	tests := []struct {
		name   string
		mutate func(
			*protocol.ComputeBenchmarkSubmission,
		)
	}{
		{
			name: "bad schema",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.SchemaVersion =
					"wrong"
			},
		},
		{
			name: "empty workload",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.WorkloadID = ""
			},
		},
		{
			name: "empty runtime",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.Runtime = ""
			},
		},
		{
			name: "empty model",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.Model = ""
			},
		},
		{
			name: "zero context",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.ContextSize = 0
			},
		},
		{
			name: "no samples",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.Samples = nil
			},
		},
		{
			name: "duplicate run",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.Samples =
					append(
						v.Benchmark.Samples,
						v.Benchmark.Samples[0],
					)
			},
		},
		{
			name: "nan generation rate",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.Samples[0].
					GenerationTokensPerSecond =
					math.NaN()
			},
		},
		{
			name: "successful run with error",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.Samples[0].Error =
					"unexpected"
			},
		},
		{
			name: "failed run without error",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				v.Benchmark.Samples[0].Success =
					false

				v.Benchmark.Samples[0].Error =
					""
			},
		},
		{
			name: "impossible temperature",
			mutate: func(
				v *protocol.ComputeBenchmarkSubmission,
			) {
				temperature := 500.0

				v.Benchmark.Samples[0].
					TemperatureCelsius =
					&temperature
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			submission :=
				validComputeBenchmarkSubmissionFixture(
					now,
				)

			test.mutate(&submission)

			if !errors.Is(
				validateComputeBenchmarkSubmission(
					submission,
					now,
				),
				ErrInvalidComputeBenchmarkSubmission,
			) {
				t.Fatal(
					"invalid compute benchmark was accepted",
				)
			}
		})
	}
}

func TestComputeBenchmarkAllowsRecordedFailure(
	t *testing.T,
) {
	now := time.Date(
		2026,
		9,
		14,
		20,
		0,
		0,
		0,
		time.UTC,
	)

	submission :=
		validComputeBenchmarkSubmissionFixture(now)

	submission.Benchmark.Samples[0] =
		protocol.ComputeBenchmarkSample{
			Run:     1,
			Success: false,
			Error:   "runtime unavailable",
		}

	if err := validateComputeBenchmarkSubmission(
		submission,
		now,
	); err != nil {
		t.Fatalf(
			"recorded benchmark failure rejected: %v",
			err,
		)
	}
}

func validComputeBenchmarkSubmissionFixture(
	now time.Time,
) protocol.ComputeBenchmarkSubmission {
	vram := uint64(
		8 * 1024 * 1024 * 1024,
	)

	temperature := 71.5
	throttled := false

	return protocol.ComputeBenchmarkSubmission{
		Version: protocol.ComputeBenchmarkSubmissionVersion,

		NodeID: "123e4567-e89b-42d3-a456-426614174000",

		ReportID: "123e4567-e89b-42d3-a456-426614174001",

		SignedAt: now,

		Benchmark: protocol.ComputeBenchmark{
			SchemaVersion: protocol.ComputeBenchmarkSchemaVersion,

			CollectedAt: now,

			WorkloadID: "m9-standard-v1",

			Condition: "baseline",

			Runtime: "llama.cpp",

			RuntimeVersion: "test-version",

			Model: "test-model",

			Quantization: "Q4_K_M",

			ContextSize: 4096,

			Samples: []protocol.ComputeBenchmarkSample{
				{
					Run: 1,

					PromptTokens: 512,

					GeneratedTokens: 128,

					PromptTokensPerSecond: 120.25,

					GenerationTokensPerSecond: 31.75,

					TimeToFirstTokenMs: 185.5,

					PeakSystemRAMBytes: 12 * 1024 * 1024 * 1024,

					PeakVRAMBytes: &vram,

					TemperatureCelsius: &temperature,

					Throttled: &throttled,

					Success: true,
				},
			},
		},
	}
}
