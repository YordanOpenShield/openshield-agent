package agentgrpc

import (
	"fmt"
	"log"
	"net"

	"openshield-agent/proto"

	"google.golang.org/grpc"
)

// AgentServer implements proto.AgentServiceServer
type AgentServer struct {
	proto.UnimplementedAgentServiceServer
}

func StartGRPCServer(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Register the gRPC server
	grpcServer := grpc.NewServer()
	proto.RegisterAgentServiceServer(grpcServer, &AgentServer{})

	log.Printf("[AGENT] gRPC server listening on port %d", port)
	return grpcServer.Serve(lis)
}
