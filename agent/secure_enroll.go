package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const SecureEnrollmentVersion = "0.1.0-m6"

func EnrollSecure(ctx context.Context, serverURL string, identity Identity, enrollmentCode string) error {
	baseURL := strings.TrimRight(strings.TrimSpace(serverURL), "/")
	if baseURL == "" {
		return errors.New("server URL is required")
	}
	if err := validateIdentity(identity); err != nil {
		return err
	}
	code := strings.TrimSpace(enrollmentCode)
	if code == "" {
		return errors.New("enrollment code is required")
	}
	requestBody := protocol.SecureEnrollRequest{
		NodeID:         identity.NodeID,
		EnrollmentCode: code,
		AgentVersion:   SecureEnrollmentVersion,
		PublicKey:      identity.PublicKey,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/agent/enroll", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("contact enrollment service: %w", err)
	}
	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 32<<10))
	if resp.StatusCode != http.StatusCreated {
		var apiError map[string]string
		_ = json.Unmarshal(responseBody, &apiError)
		if message := strings.TrimSpace(apiError["error"]); message != "" {
			return fmt.Errorf("enrollment rejected: %s", message)
		}
		return fmt.Errorf("enrollment rejected: %s", resp.Status)
	}
	var result protocol.EnrollResponse
	if err = json.Unmarshal(responseBody, &result); err != nil {
		return errors.New("enrollment service returned an invalid response")
	}
	if !result.Accepted || result.NodeID != identity.NodeID {
		return errors.New("enrollment service did not confirm this node identity")
	}
	return nil
}
