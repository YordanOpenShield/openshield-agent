package main

import (
	"flag"
	"log"
	"openshield-agent/internal/config"
	agentgrpc "openshield-agent/internal/grpc"
	"openshield-agent/internal/utils"
	"os/exec"
	"strings"
	"time"

	"openshield-agent/internal/service"
)

func main() {
	// Parse command-line arguments
	managerAddr := flag.String("manager", "", "Manager address (hostname or IP)")
	configPath := flag.String("config", config.ConfigPath, "Path to configuration file")
	scriptsPath := flag.String("scripts", config.ScriptsPath, "Path to scripts directory")
	certsPath := flag.String("certs", config.CertsPath, "Path to certificates directory")
	flag.Parse()
	config.ConfigPath = *configPath
	config.ScriptsPath = *scriptsPath
	config.CertsPath = *certsPath

	// Create the config directory if it doesn't exist
	utils.CreateConfig(config.ConfigPath, *managerAddr)
	// Create the scripts directory if it doesn't exist
	utils.CreateScriptsDir(config.ScriptsPath)
	// Create the certs directory if it doesn't exist
	utils.CreateCertsDir(config.CertsPath)

	// Check if osqueryi is installed
	_, err := exec.LookPath("osqueryi")
	if err != nil {
		log.Printf("osqueryi is not installed or not in PATH. Please install osquery using your system's package manager.")
	}

	// Load the configuration file
	err = config.LoadAndSetConfig(config.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Enroll agent
	err = service.EnrollAgent()
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			log.Printf("Agent is already enrolled: %v", err)
		} else {
			log.Fatalf("Failed to enroll agent: %v", err)
		}
	}

	// Start background tasks
	stopHeartbeat := make(chan struct{})
	service.ManagerHeartbeatMonitor(10*time.Second, stopHeartbeat)

	// Start the gRPC server
	err = agentgrpc.StartGRPCServer(50051)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
