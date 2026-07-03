package agentgrpc

import (
	"fmt"
	"log"
	"net"

	"openshield-agent/internal/utils"
	"openshield-agent/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// AgentServer implements proto.AgentServiceServer
type AgentServer struct {
	proto.UnimplementedAgentServiceServer
	managerClient *ManagerClient
}

// SetManagerClient sets the manager client for reporting results
func (s *AgentServer) SetManagerClient(client *ManagerClient) {
	s.managerClient = client
}

func StartGRPCServer(port int, server *AgentServer) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Load TLS credentials for secure communication
	tlsConfig, err := utils.LoadServerTLSCredentials()
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials: %w", err)
	}

	// Register the gRPC server
	grpcServer := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)
	proto.RegisterAgentServiceServer(grpcServer, server)

	log.Printf("[AGENT] gRPC server listening on port %d", port)
	return grpcServer.Serve(lis)
}
