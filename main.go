package main

import (
	"flag"
	"log"
	"openshield-agent/internal/config"
	"openshield-agent/internal/executor"
	agentgrpc "openshield-agent/internal/grpc"
	"openshield-agent/internal/osquery"
	"openshield-agent/internal/utils"
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

	// Ensure osquery is available (system PATH, bundled, or auto-download)
	err := osquery.EnsureOsquery()
	if err != nil {
		log.Fatalf("Failed to ensure osquery availability: %v", err)
	}
	// Set the osquery path for the executor
	executor.SetOsqueryPath(osquery.OsqueryPath())

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

	// Create manager client for reporting results
	managerClient, err := agentgrpc.NewManagerClient(config.GlobalConfig.MANAGER_ADDRESS)
	if err != nil {
		log.Fatalf("Failed to create manager client: %v", err)
	}
	defer managerClient.Close()

	// Start background tasks
	stopHeartbeat := make(chan struct{})
	service.ManagerHeartbeatMonitor(10*time.Second, stopHeartbeat)

	// Create agent server with manager client
	agentServer := &agentgrpc.AgentServer{}
	agentServer.SetManagerClient(managerClient)

	// Start the gRPC server
	err = agentgrpc.StartGRPCServer(50051, agentServer)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
