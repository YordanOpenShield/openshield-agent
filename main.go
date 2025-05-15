package main

import (
	"log"
	grpcserver "openshield-agent/internal/grpc"
	"openshield-agent/internal/utils"
	"time"
)

const configFile = "config/config.yml"

func main() {
	// Load the configuration file
	config, err := utils.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Start the heartbeat to the manager
	utils.StartHeartbeat(config.ManagerURL, 10*time.Second)

	// Start the gRPC server
	err = grpcserver.StartGRPCServer(50051)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
