package main

import (
	"log"
	"openshield-agent/internal/api"
	"openshield-agent/internal/config"
	grpcserver "openshield-agent/internal/grpc"
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
	err = config.LoadAndSetConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Start the heartbeat to the manager
	api.StartHeartbeat(1 * time.Minute)

	// Start the gRPC server
	err = grpcserver.StartGRPCServer(50051)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
