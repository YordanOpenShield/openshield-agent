package service

import (
	"context"
	"log"
	"time"

	"openshield-agent/internal/api"
	"openshield-agent/internal/config"
	agentgrpc "openshield-agent/internal/grpc"
	"openshield-agent/internal/utils"
)

// StartHeartbeatGRPC starts a goroutine that sends heartbeats to the manager over gRPC at the given interval.
// managerAddr is the address of the manager's gRPC server (e.g., "localhost:50051").
// agentID is the unique identifier for this agent.
func ManagerHeartbeatMonitor(interval time.Duration, stopCh <-chan struct{}) {
	config := config.GlobalConfig

	go func() {
		client, err := agentgrpc.NewManagerClient(config.MANAGER_ADDRESS)
		if err != nil {
			log.Print("[HEARTBEAT SYNC] Could not create client for manager")
			return
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				log.Print("[HEARTBEAT] Stopping heartbeat monitor")
				return
			case <-ticker.C:
				creds, err := utils.GetAgentCredentials()

				// If we can't get the credentials, we need to register the agent
				if err != nil {
					log.Printf("[HEARTBEAT] Failed to get agent credentials: %v. Attempting to register agent...", err)
					err = api.RegisterAgent()
					if err != nil {
						log.Fatalf("[HEARTBEAT] Agent registration failed: %v", err)
					}
					creds, err = utils.GetAgentCredentials()
					if err != nil {
						log.Fatalf("[HEARTBEAT] Failed to get agent credentials after registration: %v", err)
					}
				}
				agentID := creds.AgentID

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, err = client.Heartbeat(ctx, agentID)
				cancel()
				if err != nil {
					log.Printf("[HEARTBEAT] Heartbeat failed: %v", err)
				} else {
					log.Printf("[HEARTBEAT] Heartbeat sent to manager")
				}
			}
		}
	}()
}
