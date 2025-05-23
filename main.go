package main

import (
	"flag"
	"log"
	"openshield-agent/internal/config"
	agentgrpc "openshield-agent/internal/grpc"
	"openshield-agent/internal/utils"
	"os/exec"
	"time"

	"openshield-agent/internal/service"
)

const defaultConfigFile = "config/config.yml"

func main() {
	// Parse command-line arguments
	managerAddr := flag.String("manager", "", "Manager address (hostname or IP)")
	flag.Parse()

	// Create the config directory if it doesn't exist
	utils.CreateConfig(defaultConfigFile, *managerAddr)

	// Check if osqueryi is installed
	_, err := exec.LookPath("osqueryi")
	if err != nil {
		log.Printf("osqueryi is not installed or not in PATH. Please install osquery using your system's package manager.")
	}

	// Load the configuration file
	err = config.LoadAndSetConfig(defaultConfigFile)
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
