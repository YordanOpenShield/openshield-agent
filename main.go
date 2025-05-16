package main

import (
	"fmt"
	"log"
	grpcserver "openshield-agent/internal/grpc"
	"openshield-agent/internal/utils"
	"os"
	"time"
)

const configFile = "config/config.yml"

func main() {
	// Load the configuration file
	config, err := utils.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup osquery
	err = utils.SetupOsquery()
	if err != nil {
		fmt.Println("Error setting up osquery:", err)
		os.Exit(1)
	}
	fmt.Println("osquery setup complete.")

	// Start the heartbeat to the manager
	utils.StartHeartbeat(config.ManagerURL, 1*time.Minute)

	// Start the gRPC server
	err = grpcserver.StartGRPCServer(50051)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
