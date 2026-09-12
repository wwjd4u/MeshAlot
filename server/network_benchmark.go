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

const maxNetworkBenchmarkSubmissionBodyBytes int64 = 1024 * 1024

func decodeNetworkBenchmarkSubmissionBody(
	body []byte,
) (protocol.NetworkBenchmarkSubmission, error) {
	var submission protocol.NetworkBenchmarkSubmission

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&submission); err != nil {
		return protocol.NetworkBenchmarkSubmission{},
			ErrInvalidNetworkBenchmarkSubmission
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return protocol.NetworkBenchmarkSubmission{},
			ErrInvalidNetworkBenchmarkSubmission
	}

	return submission, nil
}

func (s *Service) networkBenchmarkSubmission(
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
			maxNetworkBenchmarkSubmissionBodyBytes,
		),
	)
	if err != nil {
		var maxBytesError *http.MaxBytesError

		if errors.As(err, &maxBytesError) {
			writeError(
				w,
				http.StatusRequestEntityTooLarge,
				"network benchmark submission too large",
			)
			return
		}

		writeError(
			w,
			http.StatusBadRequest,
			"invalid network benchmark submission",
		)
		return
	}

	submission, err := decodeNetworkBenchmarkSubmissionBody(body)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid network benchmark submission",
		)
		return
	}

	now := time.Now().UTC()

	if err := validateNetworkBenchmarkSubmission(
		submission,
		now,
	); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid network benchmark submission",
		)
		return
	}

	ctx, cancel := context.WithTimeout(
		r.Context(),
		3*time.Second,
	)
	defer cancel()

	// Reuse the same enrolled node Ed25519 public key used by the
	// authenticated M7 inventory submission path.
	publicKey, err := s.postgres.InventoryPublicKey(
		ctx,
		submission.NodeID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		writeError(
			w,
			http.StatusUnauthorized,
			"network benchmark authentication failed",
		)
		return
	}

	if err != nil {
		s.databaseError(w, err)
		return
	}

	if err := verifyNetworkBenchmarkSignature(
		publicKey,
		r.Header.Get(protocol.NetworkBenchmarkSignatureHeader),
		body,
	); err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			"network benchmark authentication failed",
		)
		return
	}

	benchmark := submission.Benchmark

	// The score is authoritative only when calculated here on the server
	// from the authenticated raw measurements.
	score := CalculateNetworkScore(
		benchmark.IPv4Available,
		benchmark.IPv6Available,
		benchmark.DownloadMbps,
		benchmark.UploadMbps,
		benchmark.LatencyMs,
		benchmark.JitterMs,
		benchmark.PacketLossPercent,
	)

	payload, err := json.Marshal(benchmark)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid network benchmark submission",
		)
		return
	}

	err = s.postgres.InsertNetworkBenchmark(
		ctx,
		submission.ReportID,
		submission.NodeID,
		payload,
		score,
		now,
	)

	if errors.Is(err, ErrNetworkBenchmarkReplay) {
		writeError(
			w,
			http.StatusConflict,
			"network benchmark report already accepted",
		)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		writeError(
			w,
			http.StatusUnauthorized,
			"network benchmark authentication failed",
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
		protocol.NetworkBenchmarkSubmissionResponse{
			ReportID: submission.ReportID,
			Accepted: true,
		},
	)
}
