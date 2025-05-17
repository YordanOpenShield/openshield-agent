package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"openshield-agent/internal/config"
	"openshield-agent/internal/utils"

	"github.com/zalando/go-keyring"
)

// RegisterAgent registers the agent with the manager via a POST request and stores credentials in the OS keyring.
func RegisterAgent() error {
	config := config.GlobalConfig

	// Add device_id to agentInfo using GetDeviceID
	agentInfo := make(map[string]interface{})
	deviceID, err := utils.GetDeviceID()
	if err != nil {
		return fmt.Errorf("failed to get device ID: %w", err)
	}
	agentInfo["device_id"] = deviceID

	url := "http://" + config.MANAGER_ADDRESS + ":" + config.MANAGER_API_PORT + "/agents/register"
	body, err := json.Marshal(agentInfo)
	fmt.Printf("[AGENT] Registering agent with body: %s\n", string(body))
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		log.Println("[AGENT] Agent already registered (409 Conflict). Continuing as normal.")
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed: %s", resp.Status)
	}
	// Debug: print raw response body
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("[AGENT] Raw registration response: %s\n", string(bodyBytes))
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset for decoding
	var result struct {
		AgentID    string `json:"id"`
		AgentToken string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	// Store credentials in OS keyring
	if err := keyring.Set("openshield-agent", "agent_id", result.AgentID); err != nil {
		log.Printf("[AGENT] Failed to store agent_id in keyring: %v", err)
	}
	if err := keyring.Set("openshield-agent", "agent_token", result.AgentToken); err != nil {
		log.Printf("[AGENT] Failed to store agent_token in keyring: %v", err)
	}
	// Always save to file as fallback
	creds := utils.AgentCredentials{AgentID: result.AgentID, AgentToken: result.AgentToken}
	if err := utils.SaveCredentialsToFile(creds); err != nil {
		log.Printf("[AGENT] Failed to save credentials to file: %v", err)
	}
	log.Println("[AGENT] Registered with manager successfully and credentials stored.")
	return nil
}
