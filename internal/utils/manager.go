package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/zalando/go-keyring"

	"openshield-agent/internal/executor"
)

// GetDeviceID retrieves the unique machine/device ID using the machineid package.
func GetDeviceID() (string, error) {
	id, err := machineid.ID()
	if err != nil {
		return "", fmt.Errorf("failed to get device ID: %w", err)
	}
	return id, nil
}

// Path to fallback credentials file
const credentialsFile = "config/agent_credentials.json"

type AgentCredentials struct {
	AgentID    string `json:"agent_id"`
	AgentToken string `json:"agent_token"`
}

// Save credentials to file
func saveCredentialsToFile(creds AgentCredentials) error {
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
func getAgentCredentials() (AgentCredentials, error) {
	id, errID := keyring.Get("openshield-agent", "agent_id")
	token, errToken := keyring.Get("openshield-agent", "agent_token")
	if errID == nil && errToken == nil {
		return AgentCredentials{AgentID: id, AgentToken: token}, nil
	}
	// Fallback to file
	return loadCredentialsFromFile()
}

// RegisterAgent registers the agent with the manager via a POST request and stores credentials in the OS keyring.
func RegisterAgent(managerURL string, agentInfo map[string]interface{}) error {
	// Add device_id to agentInfo using GetDeviceID
	deviceID, err := GetDeviceID()
	if err != nil {
		return fmt.Errorf("failed to get device ID: %w", err)
	}
	agentInfo["device_id"] = deviceID

	url := managerURL + "/agents/register"
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
	creds := AgentCredentials{AgentID: result.AgentID, AgentToken: result.AgentToken}
	if err := saveCredentialsToFile(creds); err != nil {
		log.Printf("[AGENT] Failed to save credentials to file: %v", err)
	}
	log.Println("[AGENT] Registered with manager successfully and credentials stored.")
	return nil
}

// GetLocalAddress retrieves the non-loopback local IP address of the machine.
// It first tries to use osquery via executor.RunOSQuery, and falls back to the old approach if that fails.
func GetLocalAddress() (string, error) {
	// Try using osquery
	query := "SELECT address FROM interface_addresses WHERE address NOT LIKE '127.%' AND address NOT LIKE '169.254.%' AND address LIKE '%.%' LIMIT 1;"
	results, err := executor.RunOSQuery(query)
	if err == nil && len(results) > 0 {
		if addr, ok := results[0]["address"]; ok {
			if addrStr, ok := addr.(string); ok && addrStr != "" {
				return addrStr, nil
			}
		}
	}
	// Fallback to old approach
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ip := ipnet.IP.To4()
			// Skip APIPA addresses (169.254.x.x)
			if ip[0] == 169 && ip[1] == 254 {
				continue
			}
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("no non-loopback address found")
}

// GetAllLocalAddresses returns all non-loopback, non-APIPA IPv4 addresses.
func GetAllLocalAddresses() ([]string, error) {
	var addresses []string

	// Try using osquery first
	query := "SELECT address FROM interface_addresses WHERE address NOT LIKE '127.%' AND address NOT LIKE '169.254.%' AND address LIKE '%.%';"
	results, err := executor.RunOSQuery(query)
	if err == nil && len(results) > 0 {
		for _, row := range results {
			if addr, ok := row["address"]; ok {
				if addrStr, ok := addr.(string); ok && addrStr != "" {
					addresses = append(addresses, addrStr)
				}
			}
		}
	}
	// Fallback to net.InterfaceAddrs if osquery fails or returns nothing
	if len(addresses) == 0 {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return nil, fmt.Errorf("failed to get network interfaces: %w", err)
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ip := ipnet.IP.To4()
				// Skip APIPA addresses (169.254.x.x)
				if ip[0] == 169 && ip[1] == 254 {
					continue
				}
				addresses = append(addresses, ip.String())
			}
		}
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no non-loopback addresses found")
	}
	return addresses, nil
}

// StartHeartbeat starts a goroutine that sends a heartbeat to the manager every interval.
func StartHeartbeat(managerURL string, interval time.Duration) {
	go func() {
		for {
			creds, err := getAgentCredentials()
			if err != nil {
				log.Printf("[AGENT] Failed to retrieve credentials: %v", err)
				time.Sleep(interval)
				agentInfo := make(map[string]interface{})
				RegisterAgent(managerURL, agentInfo)
				continue
			}

			addresses, err := GetAllLocalAddresses()
			if err != nil {
				log.Printf("[AGENT] Failed to get local addresses: %v", err)
				addresses = []string{}
			}

			payload := map[string]interface{}{
				"id":        creds.AgentID,
				"addresses": addresses,
			}
			body, _ := json.Marshal(payload)
			url := managerURL + "/agents/heartbeat"

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
			if err != nil {
				log.Printf("[AGENT] Heartbeat request creation error: %v", err)
				time.Sleep(interval)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Agent-Token", creds.AgentToken)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("[AGENT] Heartbeat error: %v", err)
			} else {
				resp.Body.Close()
				if resp.StatusCode == http.StatusUnauthorized {
					log.Printf("[AGENT] Heartbeat unauthorized (401), re-registering agent.")
					agentInfo := make(map[string]interface{})
					RegisterAgent(managerURL, agentInfo)
				} else if resp.StatusCode != http.StatusOK {
					log.Printf("[AGENT] Heartbeat failed: %s", resp.Status)
				} else {
					log.Println("[AGENT] Heartbeat sent.")
				}
			}
			time.Sleep(interval)
		}
	}()
}
