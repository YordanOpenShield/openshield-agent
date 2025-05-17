package utils

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

// Path to fallback credentials file
const credentialsFile = "config/agent_credentials.json"

type AgentCredentials struct {
	AgentID    string `json:"agent_id"`
	AgentToken string `json:"agent_token"`
}

// Save credentials to file
func SaveCredentialsToFile(creds AgentCredentials) error {
	_ = os.MkdirAll(filepath.Dir(credentialsFile), 0700)
	f, err := os.OpenFile(credentialsFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(&creds)
}

// Load credentials from file
func loadCredentialsFromFile() (AgentCredentials, error) {
	var creds AgentCredentials
	f, err := os.Open(credentialsFile)
	if err != nil {
		return creds, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&creds)
	return creds, err
}

// Try to get credentials from keyring, fallback to file
func GetAgentCredentials() (AgentCredentials, error) {
	id, errID := keyring.Get("openshield-agent", "agent_id")
	token, errToken := keyring.Get("openshield-agent", "agent_token")
	if errID == nil && errToken == nil {
		return AgentCredentials{AgentID: id, AgentToken: token}, nil
	}
	// Fallback to file
	return loadCredentialsFromFile()
}
