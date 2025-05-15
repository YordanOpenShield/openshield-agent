package grpcserver

import (
	"context"
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

// RegisterAgent registers the agent with the manager and returns the agent ID and token.
func (s *AgentServer) Heartbeat(ctx context.Context, req *proto.HeartbeatRequest) (*proto.HeartbeatResponse, error) {
	log.Printf("[AGENT] Received heartbeat from manager for agent %s", req.AgentId)
	return &proto.HeartbeatResponse{Ok: true, Message: "Alive"}, nil
}
