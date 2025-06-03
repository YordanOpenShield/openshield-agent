package agentgrpc

import (
	"context"
	"log"
	"openshield-agent/internal/tools"
	"openshield-agent/proto"
	"sync"

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

	// toolActionStatusMu.Lock()
	// toolActionStatus = proto.TaskStatus_RUNNING
	// toolActionStatusMu.Unlock()

	// tool, exists := tools.GetTool(req.Name)
	// if !exists {
	// 	return nil, fmt.Errorf("tool %s not found", req.Name)
	// }

	// go func() {
	// 	var (
	// 		result string
	// 		err    error
	// 	)

	// 	// Execute the command directly
	// 	parts := strings.Fields(req.Job.Target)
	// 	command := models.Command{
	// 		Command:  parts[0],
	// 		Args:     parts[1:],
	// 		TargetOS: models.GetCurrentOS(),
	// 	}
	// 	result, err = executor.ExecuteCommand(command)

	// 	if err != nil {
	// 		log.Printf("[AGENT] Command/script execution failed: %v", err)
	// 		statusMu.Lock()
	// 		currentStatus = proto.TaskStatus_FAILED
	// 		currentResult = err.Error()
	// 		statusMu.Unlock()
	// 		log.Printf("[AGENT] Task %s failed", req.Task.Id)
	// 		return
	// 	}

	// 	// Set the result and status
	// 	statusMu.Lock()
	// 	currentStatus = proto.TaskStatus_COMPLETED
	// 	currentResult = result
	// 	statusMu.Unlock()

	// 	log.Printf("[AGENT] Task %s completed", req.Task.Id)
	// }()

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
	return &proto.ToolExecutionStatusResponse{
		Name:   req.Name,
		Action: req.Action,
		Status: proto.TaskStatus_COMPLETED,
		Result: "Execution result here",
	}, nil
}
