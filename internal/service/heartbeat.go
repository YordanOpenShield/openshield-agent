package service

import (
	"context"
	"log"
	"time"

	"openshield-agent/internal/config"
	agentgrpc "openshield-agent/internal/grpc"
	"strings"
)

// StartHeartbeatGRPC starts a goroutine that sends heartbeats to the manager over gRPC at the given interval.
// managerAddr is the address of the manager's gRPC server (e.g., "localhost:50051").
// agentID is the unique identifier for this agent.
func ManagerHeartbeatMonitor(interval time.Duration, stopCh <-chan struct{}) {
	config := config.GlobalConfig

	go func() {
		client, err := agentgrpc.NewManagerClient(config.MANAGER_ADDRESS)
		if err != nil {
			log.Printf("[HEARTBEAT SYNC] Could not create client for manager: %v", err)
			EnrollAgent()
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				log.Print("[HEARTBEAT] Stopping heartbeat monitor")
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, err = client.Heartbeat(ctx)
				cancel()
				if err != nil {
					if strings.Contains(err.Error(), "record not found") {
						log.Printf("[HEARTBEAT] Registering agent")
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						_, err := client.RegisterAgent(ctx)
						cancel()
						if err != nil {
							log.Printf("[HEARTBEAT] Agent registration failed: %v", err)
						}
					}
					log.Printf("[HEARTBEAT] Heartbeat failed: %v", err)
				} else {
					log.Printf("[HEARTBEAT] Heartbeat sent to manager")
				}
			}
		}
	}()
}
