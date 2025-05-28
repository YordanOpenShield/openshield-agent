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

func main() {
	// Parse command-line arguments
	managerAddr := flag.String("manager", "", "Manager address (hostname or IP)")
	configPath := flag.String("config", config.ConfigPath, "Path to configuration file")
	scriptsPath := flag.String("scripts", config.ScriptsPath, "Path to scripts directory")
	flag.Parse()
	config.ConfigPath = *configPath
	config.ScriptsPath = *scriptsPath

	// Create the config directory if it doesn't exist
	utils.CreateConfig(config.ConfigPath, *managerAddr)
	// Create the scripts directory if it doesn't exist
	utils.CreateScriptsDir(config.ScriptsPath)

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

	// Start background tasks
	stopHeartbeat := make(chan struct{})
	service.ManagerHeartbeatMonitor(10*time.Second, stopHeartbeat)

	// Start the gRPC server
	err = agentgrpc.StartGRPCServer(50051)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
