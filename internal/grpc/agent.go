package agentgrpc

import (
	"context"
	"encoding/json"
	"log"
	"openshield-agent/internal/config"
	"openshield-agent/internal/utils"
	"openshield-agent/proto"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ClearAgentKeyring clears the agent's credentials from the keyring.
func clearAgentKeyring() error {
	_ = keyring.Delete("openshield-agent", "agent_id")
	_ = keyring.Delete("openshield-agent", "agent_token")
	return nil
}

// DeleteAgentCredentialsFile deletes the agent_credentials.json file from the config directory.
func DeleteAgentCredentialsFile() error {
	configPath := filepath.Join("config", "agent_credentials.json")
	return os.Remove(configPath)
}

func DeleteAgentCredentials() error {
	if err := clearAgentKeyring(); err != nil {
		return err
	}
	if err := DeleteAgentCredentialsFile(); err != nil {
		return err
	}
	return nil
}

// UnregisterAgentAsk handles the UnregisterAgentAsk RPC and deletes agent credentials.
func (s *AgentServer) UnregisterAgentAsk(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	config := config.GlobalConfig

	// Unregister the agent
	client, err := NewManagerClient(config.MANAGER_ADDRESS)
	if err != nil {
		log.Print("[AGENT] Could not create client for manager")
		return nil, err
	}
	client.UnregisterAgent(ctx)

	log.Printf("[AGENT] Unregistering agent from manager")

	return &emptypb.Empty{}, nil
}

// TryAgentAddresses handles the TryAgentAddress RPC and logs all local addresses.
func (s *AgentServer) TryAgentAddress(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// RegisterAgent registers the agent with the manager.
func (c *ManagerClient) RegisterAgent(ctx context.Context) (*proto.RegisterAgentResponse, error) {
	// Get device's ID
	deviceID, err := utils.GetDeviceID()
	if err != nil {
		return nil, err

	}

	resp, err := c.client.RegisterAgent(ctx, &proto.RegisterAgentRequest{DeviceId: deviceID})
	if err != nil {
		return nil, err
	}

	// Save agent credentials
	creds := utils.AgentCredentials{AgentID: resp.Id, AgentToken: resp.Token}
	if err := utils.SaveAgentCredentials(creds); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *ManagerClient) UnregisterAgent(ctx context.Context) error {
	creds, err := utils.GetAgentCredentials()
	if err != nil {
		return err
	}

	_, err = c.client.UnregisterAgent(ctx, &proto.UnregisterAgentRequest{Id: creds.AgentID})
	if err != nil {
		return err
	}

	return DeleteAgentCredentials()
}

// Heartbeat sends a heartbeat signal to the agent and checks if it's alive.
func (c *ManagerClient) Heartbeat(ctx context.Context) (bool, error) {
	// Fetch credentials
	creds, err := utils.GetAgentCredentials()
	if err != nil {
		return false, err
	}

	// Collect all addresses
	addresses, err := utils.GetAllLocalAddresses()
	if err != nil {
		return false, err
	}

	// Collect all services
	services, err := utils.GetAllServices()
	if err != nil {
		return false, err
	}

	// Prepare a JSON message
	msg := map[string]interface{}{
		"id":        creds.AgentID,
		"addresses": addresses,
		"services":  services,
		"os":        utils.GetDeviceOS(),
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}

	// Send the request
	req := &proto.HeartbeatRequest{AgentId: creds.AgentID, Message: string(jsonMsg)}
	resp, err := c.client.Heartbeat(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.Ok, nil
}
