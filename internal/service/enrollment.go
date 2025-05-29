package service

import (
	"context"
	"log"
	"openshield-agent/internal/config"
	agentgrpc "openshield-agent/internal/grpc"
	"openshield-agent/internal/utils"
	"os"
	"path/filepath"
	"time"
)

func registerAgent() error {
	log.Printf("[AGENT] Registering agent with manager at %s", config.GlobalConfig.MANAGER_ADDRESS)
	// Create client for registration
	client, err := agentgrpc.NewRegistrationClient(config.GlobalConfig.MANAGER_ADDRESS)
	if err != nil {
		log.Printf("[AGENT] Could not create client for manager: %v", err)
		return err
	}
	// Register the agent
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.RegisterAgent(ctx)
	if err != nil {
		log.Printf("[AGENT] Agent registration failed: %v", err)
		return err
	}

	return generateCerts(resp.Id)
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

func EnrollAgent() error {
	// Register the agent
	err := registerAgent()
	if err != nil {
		return err
	}

	return nil
}
