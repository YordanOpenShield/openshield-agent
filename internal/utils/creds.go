package utils

import (
	"encoding/json"
	"openshield-agent/internal/config"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const credentialsFilename = "agent_credentials.json"

type AgentCredentials struct {
	AgentID    string `json:"id"`
	AgentToken string `json:"token"`
}

// Save credentials to keyring
func SaveCredentialsToKeyring(creds AgentCredentials) error {
	if err := keyring.Set("openshield-agent", "agent_id", creds.AgentID); err != nil {
		return err
	}
	if err := keyring.Set("openshield-agent", "agent_token", creds.AgentToken); err != nil {
		return err
	}
	return nil
}

// Save credentials to file
func SaveCredentialsToFile(creds AgentCredentials) error {
	var credentialsFile = config.ConfigPath + "/" + credentialsFilename
	_ = os.MkdirAll(filepath.Dir(credentialsFile), 0700)
	f, err := os.OpenFile(credentialsFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(&creds)
}

// Save credentials to both keyring and file
func SaveAgentCredentials(creds AgentCredentials) error {
	if err := SaveCredentialsToKeyring(creds); err != nil {
		return SaveCredentialsToFile(creds)
	}
	return SaveCredentialsToFile(creds)
}

// Load credentials from file
func loadCredentialsFromFile() (AgentCredentials, error) {
	var credentialsFile = config.ConfigPath + "/" + credentialsFilename
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
