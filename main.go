package main

import (
	"log"
	grpcserver "openshield-agent/internal/grpc"
	"openshield-agent/internal/utils"
	"os/exec"
	"time"
)

const configFile = "config/config.yml"

func main() {
	// Check if osqueryi is installed
	_, err := exec.LookPath("osqueryi")
	if err != nil {
		log.Printf("osqueryi is not installed or not in PATH. Please install osquery using your system's package manager.")
	}

	// Load the configuration file
	config, err := utils.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Start the heartbeat to the manager
	utils.StartHeartbeat(config.ManagerURL, 1*time.Minute)

	// Start the gRPC server
	err = grpcserver.StartGRPCServer(50051)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
