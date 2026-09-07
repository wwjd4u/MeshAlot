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

const maxInventorySubmissionBodyBytes int64 = 1024 * 1024

func decodeInventorySubmissionBody(
	body []byte,
) (protocol.InventorySubmission, error) {

	var submission protocol.InventorySubmission

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&submission); err != nil {
		return protocol.InventorySubmission{},
			ErrInvalidInventorySubmission
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return protocol.InventorySubmission{},
			ErrInvalidInventorySubmission
	}

	return submission, nil
}

func (s *Service) inventorySubmission(
	w http.ResponseWriter,
	r *http.Request,
) {

	if s.postgres == nil {
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	body, err := io.ReadAll(
		http.MaxBytesReader(
			w,
			r.Body,
			maxInventorySubmissionBodyBytes,
		),
	)

	if err != nil {
		var maxBytesError *http.MaxBytesError

		if errors.As(err, &maxBytesError) {
			writeError(
				w,
				http.StatusRequestEntityTooLarge,
				"inventory submission too large",
			)
			return
		}

		writeError(
			w,
			http.StatusBadRequest,
			"invalid inventory submission",
		)
		return
	}

	submission, err := decodeInventorySubmissionBody(body)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid inventory submission",
		)
		return
	}

	now := time.Now().UTC()

	if err := validateInventorySubmission(
		submission,
		now,
	); err != nil {

		writeError(
			w,
			http.StatusBadRequest,
			"invalid inventory submission",
		)
		return
	}

	ctx, cancel := context.WithTimeout(
		r.Context(),
		3*time.Second,
	)
	defer cancel()

	publicKey, err := s.postgres.InventoryPublicKey(
		ctx,
		submission.NodeID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		writeError(
			w,
			http.StatusUnauthorized,
			"inventory authentication failed",
		)
		return
	}

	if err != nil {
		s.databaseError(w, err)
		return
	}

	if err := verifyInventorySignature(
		publicKey,
		r.Header.Get(protocol.InventorySignatureHeader),
		body,
	); err != nil {

		writeError(
			w,
			http.StatusUnauthorized,
			"inventory authentication failed",
		)
		return
	}

	payload, err := json.Marshal(submission.Inventory)
	if err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid inventory submission",
		)
		return
	}

	err = s.postgres.InsertInventory(
		ctx,
		submission.ReportID,
		submission.NodeID,
		payload,
		now,
	)

	if errors.Is(err, ErrInventoryReplay) {
		writeError(
			w,
			http.StatusConflict,
			"inventory report already accepted",
		)
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		writeError(
			w,
			http.StatusUnauthorized,
			"inventory authentication failed",
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
		protocol.InventorySubmissionResponse{
			ReportID: submission.ReportID,
			Accepted: true,
		},
	)
}

func (s *Service) accountNodeInventory(
	w http.ResponseWriter,
	r *http.Request,
) {

	user := currentUser(r)

	ctx, cancel := context.WithTimeout(
		r.Context(),
		3*time.Second,
	)
	defer cancel()

	report, err := s.postgres.LatestInventoryForUser(
		ctx,
		user.ID,
		r.PathValue("nodeID"),
	)

	if errors.Is(err, sql.ErrNoRows) {
		writeError(
			w,
			http.StatusNotFound,
			"inventory not found",
		)
		return
	}

	if err != nil {
		s.databaseError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		report,
	)
}
