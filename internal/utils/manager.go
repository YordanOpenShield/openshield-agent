package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/zalando/go-keyring"
)

// GetDeviceID retrieves the unique machine/device ID using the machineid package.
func GetDeviceID() (string, error) {
	id, err := machineid.ID()
	if err != nil {
		return "", fmt.Errorf("failed to get device ID: %w", err)
	}
	return id, nil
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
		return fmt.Errorf("failed to store agent_id in keyring: %w", err)
	}
	if err := keyring.Set("openshield-agent", "agent_token", result.AgentToken); err != nil {
		return fmt.Errorf("failed to store agent_token in keyring: %w", err)
	}
	log.Println("[AGENT] Registered with manager successfully and credentials stored in keyring.")
	return nil
}

// GetLocalAddress retrieves the non-loopback local IP address of the machine.
func GetLocalAddress() (string, error) {
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

// StartHeartbeat starts a goroutine that sends a heartbeat to the manager every interval.
func StartHeartbeat(managerURL string, interval time.Duration) {
	go func() {
		for {
			agentID, err := keyring.Get("openshield-agent", "agent_id")
			if err != nil {
				log.Printf("[AGENT] Failed to retrieve agent_id from keyring: %v", err)
				time.Sleep(interval)
				// Register again if heartbeat fails
				agentInfo := make(map[string]interface{})
				RegisterAgent(managerURL, agentInfo)
				continue
			}
			agentToken, err := keyring.Get("openshield-agent", "agent_token")
			if err != nil {
				log.Printf("[AGENT] Failed to retrieve agent_token from keyring: %v", err)
				time.Sleep(interval)
				// Register again if heartbeat fails
				agentInfo := make(map[string]interface{})
				RegisterAgent(managerURL, agentInfo)
				continue
			}

			localAddr, err := GetLocalAddress()
			if err != nil {
				log.Printf("[AGENT] Failed to get local address: %v", err)
				localAddr = ""
			}

			payload := map[string]string{
				"id":      agentID,
				"address": localAddr,
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
			req.Header.Set("X-Agent-Token", agentToken)

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
