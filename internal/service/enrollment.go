package service

import (
	"context"
	"errors"
	"log"
	"openshield-agent/internal/config"
	agentgrpc "openshield-agent/internal/grpc"
	"openshield-agent/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func registerAgent() (string, error) {
	log.Printf("[AGENT] Registering agent with manager at %s", config.GlobalConfig.MANAGER_ADDRESS)
	// Create client for registration
	client, err := agentgrpc.NewRegistrationClient(config.GlobalConfig.MANAGER_ADDRESS)
	if err != nil {
		log.Printf("[AGENT] Could not create client for manager: %v", err)
		return "", err
	}
	// Register the agent
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.RegisterAgent(ctx)
	if err != nil {
		log.Printf("[AGENT] Agent registration failed: %v", err)
		return "", err
	}

	return resp.Id, nil
}

func generateCerts(agentId string) error {
	// Generate private key and CSR (Certificate Signing Request) for the agent
	err := utils.GeneratePrivateKey()
	if err != nil {
		log.Printf("[AGENT] Failed to generate private key: %v", err)
		return err
	}
	err = utils.GenerateCSR(agentId)
	if err != nil {
		log.Printf("[AGENT] Failed to generate CSR: %v", err)
		return err
	}

	// Read the agent.csr file from the certs directory
	csrPath := filepath.Join(config.CertsPath, "agent.csr")
	csrData, err := os.ReadFile(csrPath)
	if err != nil {
		return err
	}
	// Request certificate from the manager
	certsResp, err := utils.RequestCSRSigning([]byte(csrData))
	if err != nil {
		log.Printf("[AGENT] Failed to request certificate signing: %v", err)
		return err
	}
	// Save the signed certificate and CA certificate
	err = utils.SaveCertificates(certsResp)
	if err != nil {
		log.Printf("[AGENT] Failed to save certificates: %v", err)
		return err
	}

	return nil
}

// isAlreadyEnrolled checks if the agent already has credentials stored
func isAlreadyEnrolled() (bool, utils.AgentCredentials, error) {
	creds, err := utils.GetAgentCredentials()
	if err != nil {
		return false, creds, err
	}
	return creds.AgentID != "", creds, nil
}

func EnrollAgent() error {
	// Check if already enrolled
	enrolled, creds, err := isAlreadyEnrolled()
	if err != nil {
		log.Printf("[AGENT] No existing credentials found, proceeding with registration")
	}

	var agentId string
	if enrolled {
		log.Printf("[AGENT] Agent already enrolled with ID: %s", creds.AgentID)
		agentId = creds.AgentID
	} else {
		// Try to register the agent
		agentId, err = registerAgent()
		if err != nil {
			// Check if the error is because agent already exists
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
				log.Printf("[AGENT] Agent is already enrolled on manager, attempting to retrieve credentials")
				// Try to load credentials again - they might have been saved
				enrolled, creds, err = isAlreadyEnrolled()
				if err != nil || !enrolled {
					return errors.New("agent exists on manager but local credentials not found - manual cleanup required")
				}
				agentId = creds.AgentID
			} else {
				return err
			}
		}
	}

	// Generate certificates for the agent
	return generateCerts(agentId)
}
