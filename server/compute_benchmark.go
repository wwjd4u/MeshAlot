package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const maxComputeBenchmarkSubmissionBodyBytes int64 = 1024 * 1024

func decodeComputeBenchmarkSubmissionBody(
	body []byte,
) (protocol.ComputeBenchmarkSubmission, error) {
	var submission protocol.ComputeBenchmarkSubmission

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&submission); err != nil {
		return protocol.ComputeBenchmarkSubmission{},
			ErrInvalidComputeBenchmarkSubmission
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return protocol.ComputeBenchmarkSubmission{},
			ErrInvalidComputeBenchmarkSubmission
	}

	return submission, nil
}

func (s *Service) computeBenchmarkSubmission(
	w http.ResponseWriter,
	r *http.Request,
) {
	if s.postgres == nil {
		writeError(
			w,
			http.StatusServiceUnavailable,
			"database unavailable",
		)
		return
	}

	body, err := io.ReadAll(
		http.MaxBytesReader(
			w,
			r.Body,
			maxComputeBenchmarkSubmissionBodyBytes,
		),
	)

	if err != nil {
		var maxBytesError *http.MaxBytesError

		if errors.As(err, &maxBytesError) {
			writeError(
				w,
				http.StatusRequestEntityTooLarge,
				"compute benchmark submission too large",
			)
			return
		}

		writeError(
			w,
			http.StatusBadRequest,
			"invalid compute benchmark submission",
		)
		return
	}

	submission, err :=
		decodeComputeBenchmarkSubmissionBody(body)

	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid compute benchmark submission",
		)
		return
	}

	now := time.Now().UTC()

	if err := validateComputeBenchmarkSubmission(
		submission,
		now,
	); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid compute benchmark submission",
		)
		return
	}

	ctx, cancel := context.WithTimeout(
		r.Context(),
		3*time.Second,
	)
	defer cancel()

	// Reuse the enrolled node Ed25519 public identity established by M6/M7.
	publicKey, err := s.postgres.InventoryPublicKey(
		ctx,
		submission.NodeID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		writeError(
			w,
			http.StatusUnauthorized,
			"compute benchmark authentication failed",
		)
		return
	}

	if err != nil {
		s.databaseError(w, err)
		return
	}

	if err := verifyComputeBenchmarkSignature(
		publicKey,
		r.Header.Get(
			protocol.ComputeBenchmarkSignatureHeader,
		),
		body,
	); err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			"compute benchmark authentication failed",
		)
		return
	}

	benchmark := submission.Benchmark

	// The control server calculates the authoritative score from authenticated
	// raw measurements. An agent-provided score is never accepted.
	score := CalculateComputeScore(benchmark)

	payload, err := json.Marshal(benchmark)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid compute benchmark submission",
		)
		return
	}

	err = s.postgres.InsertComputeBenchmark(
		ctx,
		submission.ReportID,
		submission.NodeID,
		payload,
		score.Score,
		now,
	)

	if errors.Is(err, ErrComputeBenchmarkReplay) {
		writeError(
			w,
			http.StatusConflict,
			"compute benchmark report already accepted",
		)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		writeError(
			w,
			http.StatusUnauthorized,
			"compute benchmark authentication failed",
		)
		return
	}

	if err != nil {
		s.databaseError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		protocol.ComputeBenchmarkSubmissionResponse{
			ReportID: submission.ReportID,
			Accepted: true,
		},
	)
}
