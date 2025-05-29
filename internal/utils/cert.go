package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"openshield-agent/internal/config"
)

// CertResponse represents the expected JSON structure from the manager.
type CertResponse struct {
	CA   string `json:"ca"`
	Cert string `json:"cert"`
}

// RequestCSRSigning sends a CSR to the manager for signing and returns the signed certificates.
func RequestCSRSigning(agentToken string, csr []byte) (*CertResponse, error) {
	req, err := http.NewRequest("POST", config.GlobalConfig.MANAGER_ADDRESS, bytes.NewReader(csr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Agent-Token", agentToken)
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var certResp CertResponse
	if err := json.NewDecoder(resp.Body).Decode(&certResp); err != nil {
		return nil, err
	}
	return &certResp, nil
}
