package agentgrpc

import (
	"openshield-agent/internal/config"
	"openshield-agent/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentClient wraps the gRPC client and connection.
type ManagerClient struct {
	conn   *grpc.ClientConn
	client proto.ManagerServiceClient
}

// NewAgentClient initializes and returns a new AgentClient.
func NewManagerClient(managerAddress string) (*ManagerClient, error) {
	config := config.GlobalConfig

	conn, err := grpc.NewClient(
		managerAddress+":"+config.MANAGER_GRPC_PORT,
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

// Close terminates the connection to the agent.
func (a *ManagerClient) Close() {
	a.conn.Close()
}
