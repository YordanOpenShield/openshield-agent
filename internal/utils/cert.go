package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"openshield-agent/internal/config"
	"os"
	"path/filepath"
)

// GeneratePrivateKey generates a new RSA private key and saves it to the specified directory.
func GeneratePrivateKey() error {
	keyPath := filepath.Join(config.CertsPath, "agent.key")
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyFile.Close()
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return err
	}
	return nil
}

// GenerateCSR generates a Certificate Signing Request (CSR) using the provided private key and common name.
func GenerateCSR(commonName string) error {
	csrPath := filepath.Join(config.CertsPath, "agent.csr")
	keyFile, err := os.ReadFile(config.CertsPath + "/agent.key")
	if err != nil {
		return err
	}
	block, _ := pem.Decode(keyFile)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block from key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: commonName,
		},
	}
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		return err
	}
	csrFile, err := os.OpenFile(csrPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer csrFile.Close()
	if err := pem.Encode(csrFile, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}); err != nil {
		return err
	}
	return nil
}

// CertResponse represents the expected JSON structure from the manager.
type CertResponse struct {
	Cert string `json:"agent_cert"`
	CA   string `json:"ca_cert"`
}

// RequestCSRSigning sends a CSR to the manager for signing and returns the signed certificates.
func RequestCSRSigning(csr []byte) (*CertResponse, error) {
	// Fetch agent credentials
	creds, err := GetAgentCredentials()
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request to the manager
	req, err := http.NewRequest("POST", "http://"+config.GlobalConfig.MANAGER_ADDRESS+":"+config.GlobalConfig.MANAGER_API_PORT+"/api/certs/sign", bytes.NewReader(csr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Agent-Token", creds.AgentToken)
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

// SaveCertificates saves the signed agent and CA certificates to disk.
func SaveCertificates(certsResp *CertResponse) error {
	agentCertPath := filepath.Join(config.CertsPath, "agent.crt")
	if err := os.WriteFile(agentCertPath, []byte(certsResp.Cert), 0644); err != nil {
		log.Printf("[AGENT] Failed to save agent certificate: %v", err)
		return err
	}
	caCertPath := filepath.Join(config.CertsPath, "ca.crt")
	if err := os.WriteFile(caCertPath, []byte(certsResp.CA), 0644); err != nil {
		log.Printf("[AGENT] Failed to save CA certificate: %v", err)
		return err
	}
	return nil
}

func LoadClientTLSCredentials() (*tls.Config, error) {
	// Load TLS credentials from the provided files
	cert, err := tls.LoadX509KeyPair(config.CertsPath+"/agent.crt", config.CertsPath+"/agent.key")
	if err != nil {
		return nil, err
	}
	caCert, err := os.ReadFile(config.CertsPath + "/ca.crt")
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}, nil

}

func LoadServerTLSCredentials() (*tls.Config, error) {
	// Load TLS credentials from the provided files
	cert, err := tls.LoadX509KeyPair(config.CertsPath+"/agent.crt", config.CertsPath+"/agent.key")
	if err != nil {
		return nil, err
	}
	caCert, err := os.ReadFile(config.CertsPath + "/ca.crt")
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}
