package agentgrpc

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"openshield-agent/internal/models"
	"openshield-agent/internal/tools"
	"openshield-agent/proto"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

// In-memory tool action state for now
var (
	toolActionStatus   proto.TaskStatus = proto.TaskStatus_PENDING
	toolActionResult   string           = ""
	toolActionStatusMu sync.Mutex
)

// GetTools handles the GetTools RPC.
func (s *AgentServer) GetTools(ctx context.Context, req *emptypb.Empty) (*proto.GetToolsResponse, error) {
	tools := tools.GetTools()

	var toolList []*proto.Tool
	for _, tool := range tools {
		toolProto := &proto.Tool{
			Name:    tool.Name,
			Actions: make([]*proto.ToolAction, len(tool.Actions)),
			Os:      tool.OS,
		}
		for i, action := range tool.Actions {
			toolProto.Actions[i] = &proto.ToolAction{
				Name:    action.Name,
				Options: action.Opts,
			}
		}
		toolList = append(toolList, toolProto)
	}

	return &proto.GetToolsResponse{
		Tools: toolList,
	}, nil
}

// ExecuteTool handles the ExecuteTool RPC.
func (s *AgentServer) ExecuteTool(ctx context.Context, req *proto.ExecuteToolRequest) (*proto.ExecuteToolResponse, error) {
	// TODO: Implement logic to execute the requested tool with given action and options
	log.Printf("[AGENT] Received tool execution: %s (%s) [%v]", req.Name, req.Action, req.Options)

	// Check if another action is currently executing
	toolActionStatusMu.Lock()
	if toolActionStatus == proto.TaskStatus_RUNNING {
		toolActionStatusMu.Unlock()
		return nil, fmt.Errorf("another tool action is currently executing")
	}
	toolActionStatusMu.Unlock()

	// Set the status to running
	toolActionStatusMu.Lock()
	toolActionStatus = proto.TaskStatus_RUNNING
	toolActionStatusMu.Unlock()

	// Check if the tool exists
	tool, exists := tools.GetTool(req.Name)
	if !exists {
		return nil, fmt.Errorf("tool %s not found", req.Name)
	}
	// Check if the action is supported
	if !tool.isActionSupported(req.Action) {
		return nil, fmt.Errorf("action %s not supported by tool %s", req.Action, req.Name)
	}
	// Check if the OS is supported
	if !tool.isOSSupported(models.GetCurrentOS()) {
		return nil, fmt.Errorf("tool %s is not supported on this OS", req.Name)
	}

	go func() {
		var (
			result string
			err    error
		)

		// Execute the action
		output, err := tool.ExecAction(req.Action, req.Options)
		if err != nil {
			log.Printf("[TOOL] Tool action execution failed: %v", err)
			toolActionStatusMu.Lock()
			toolActionStatus = proto.TaskStatus_FAILED
			toolActionResult = err.Error()
			toolActionStatusMu.Unlock()
			log.Printf("[AGENT] Tool %s action %s failed", req.Name, req.Action)
			return
		}

		// Set the result and status
		toolActionStatusMu.Lock()
		toolActionStatus = proto.TaskStatus_COMPLETED
		toolActionResult = result
		toolActionStatusMu.Unlock()

		log.Printf("[AGENT] Tool %s action %s completed", req.Name, req.Action)
	}()

	// Start a goroutine to report task status every second
	go func() {
		for {
			toolActionStatusMu.Lock()
			if toolActionStatus == proto.TaskStatus_COMPLETED || toolActionStatus == proto.TaskStatus_FAILED {
				toolActionStatusMu.Unlock()
				break
			}
			log.Printf("[AGENT] Tool %s action %s status: %v", req.Name, req.Action, toolActionStatus)
			toolActionStatusMu.Unlock()
			time.Sleep(5 * time.Second)
		}
	}()

	return &proto.ExecuteToolResponse{
		Name:     req.Name,
		Action:   req.Action,
		Accepted: true,
		Message:  "Tool execution started",
	}, nil
}

// ReportToolExecutionStatus handles the ReportToolExecutionStatus RPC.
func (s *AgentServer) ReportToolExecutionStatus(ctx context.Context, req *proto.ToolExecutionStatusRequest) (*proto.ToolExecutionStatusResponse, error) {
	// TODO: Implement logic to report the status of a tool execution
	toolActionStatusMu.Lock()
	defer toolActionStatusMu.Unlock()

	log.Printf("[AGENT] Reporting status for tool %s action %s", req.Name, req.Action)

	// Encode the result to base64 to ensure safe transmission
	encodedResult := base64.StdEncoding.EncodeToString([]byte(toolActionResult))

	return &proto.ToolExecutionStatusResponse{
		Name:   req.Name,
		Action: req.Action,
		Status: toolActionStatus,
		Result: encodedResult,
	}, nil
}
