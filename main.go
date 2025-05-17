package main

import (
	"log"
	"openshield-agent/internal/config"
	agentgrpc "openshield-agent/internal/grpc"
	"os/exec"
	"time"

	"openshield-agent/internal/service"
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

	// Start background tasks
	stopHeartbeat := make(chan struct{})
	service.ManagerHeartbeatMonitor(10*time.Second, stopHeartbeat)

	// Start the gRPC server
	err = agentgrpc.StartGRPCServer(50051)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
