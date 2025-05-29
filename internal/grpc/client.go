package agentgrpc

import (
	"context"
	"log"
	"openshield-agent/internal/config"
	"openshield-agent/internal/utils"
	"openshield-agent/proto"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentClient wraps the gRPC client and connection.
type ManagerClient struct {
	conn   *grpc.ClientConn
	client proto.ManagerServiceClient
}

func NewRegistrationClient(managerAddress string) (*ManagerClient, error) {
	conn, err := grpc.NewClient(
		managerAddress+":"+config.GlobalConfig.MANAGER_REGISTER_PORT,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Use TLS in production
	)
	if err != nil {
		return nil, err
	}

	client := proto.NewManagerServiceClient(conn)

	return &ManagerClient{
		conn:   conn,
		client: client,
	}, nil
}

// NewAgentClient initializes and returns a new AgentClient.
func NewManagerClient(managerAddress string) (*ManagerClient, error) {
	// Load TLS credentials
	tlsConfig, err := utils.LoadClientTLSCredentials()
	if err != nil {
		log.Printf("[AGENT] Failed to load TLS credentials: %v", err)
		return nil, err
	}

	creds, err := utils.GetAgentCredentials()
	if err != nil {
		// Create a client without credentials only for registration
		conn, err := grpc.NewClient(
			managerAddress+":"+config.GlobalConfig.MANAGER_GRPC_PORT,
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), // Use TLS in production
		)
		if err != nil {
			return nil, err
		}

		client := proto.NewManagerServiceClient(conn)

		managerClient := &ManagerClient{
			conn:   conn,
			client: client,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Attempt to register the agent
		resp, err := managerClient.RegisterAgent(ctx)
		if err != nil {
			return nil, err
		}
		log.Printf("[AGENT] Agent registered: %v", resp)
	}

	conn, err := grpc.NewClient(
		managerAddress+":"+config.GlobalConfig.MANAGER_GRPC_PORT,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), // Use TLS in production
		WithAgentToken(creds.AgentToken),                             // Inject the agent token
	)
	if err != nil {
		return nil, err
	}

	client := proto.NewManagerServiceClient(conn)

	return &ManagerClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close terminates the connection to the agent.
func (a *ManagerClient) Close() {
	a.conn.Close()
}
