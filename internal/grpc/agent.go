package agentgrpc

import (
	"context"
	"encoding/json"
	"openshield-agent/internal/utils"
	"openshield-agent/proto"
)

// Heartbeat sends a heartbeat signal to the agent and checks if it's alive.
func (c *ManagerClient) Heartbeat(ctx context.Context, agentID string) (bool, error) {
	// Fetch credentials
	creds, err := utils.GetAgentCredentials()
	if err != nil {
		return false, err
	}

	// Collect data
	addresses, err := utils.GetAllLocalAddresses()
	if err != nil {
		return false, err
	}

	// Prepare a JSON message
	msg := map[string]interface{}{
		"id":        creds.AgentID,
		"addresses": addresses,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}

	// Send the request
	req := &proto.HeartbeatRequest{AgentId: agentID, Message: string(jsonMsg)}
	resp, err := c.client.Heartbeat(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.Ok, nil
}
