package agentgrpc

import (
	"context"
	"encoding/base64"
	"log"
	"strings"
	"sync"
	"time"

	"openshield-agent/internal/executor"
	"openshield-agent/internal/models"
	"openshield-agent/proto"
)

// In-memory task state for now
var (
	currentStatus proto.TaskStatus = proto.TaskStatus_PENDING
	currentResult string           = ""
	statusMu      sync.Mutex
)

// AssignTask handles a new task assignment from the manager
func (s *AgentServer) AssignTask(ctx context.Context, req *proto.AssignTaskRequest) (*proto.AssignTaskResponse, error) {
	log.Printf("[AGENT] Received task: %s (%s)", req.Task.Id, req.Job.Name)

	statusMu.Lock()
	currentStatus = proto.TaskStatus_RUNNING
	statusMu.Unlock()

	// Start a goroutine to execute the task and update the status
	go func() {
		// Run the command
		command := models.Command{
			Command:  strings.Split(req.Job.Command, " ")[0],
			Args:     strings.Split(req.Job.Command, " ")[1:],
			TargetOS: models.GetCurrentOS(),
		}
		result, err := executor.ExecuteCommand(command)
		if err != nil {
			log.Printf("[AGENT] Error executing command: %v", err)
		}
		if err != nil {
			log.Printf("[AGENT] Command execution failed: %v", err)
			statusMu.Lock()
			currentStatus = proto.TaskStatus_FAILED
			currentResult = err.Error()
			statusMu.Unlock()
			log.Printf("[AGENT] Task %s failed", req.Task.Id)
			return
		}

		log.Printf("[AGENT] Command output: %s", result)

		// Set the result and status
		statusMu.Lock()
		currentStatus = proto.TaskStatus_COMPLETED
		currentResult = result
		statusMu.Unlock()

		log.Printf("[AGENT] Task %s completed", req.Task.Id)
	}()

	// Start a goroutine to report task status every second
	go func() {
		for {
			statusMu.Lock()
			if currentStatus == proto.TaskStatus_COMPLETED || currentStatus == proto.TaskStatus_FAILED {
				statusMu.Unlock()
				break
			}
			log.Printf("[AGENT] Task %s status: %v", req.Task.Id, currentStatus)
			statusMu.Unlock()
			time.Sleep(5 * time.Second)
		}
	}()

	return &proto.AssignTaskResponse{
		Accepted: true,
		Message:  "Task accepted",
	}, nil
}

// ReportTaskStatus returns the current status and result of a task
func (s *AgentServer) ReportTaskStatus(ctx context.Context, req *proto.JobStatusRequest) (*proto.JobStatusResponse, error) {
	statusMu.Lock()
	defer statusMu.Unlock()

	log.Printf("[AGENT] Reporting status for job %s", req.JobId)

	// Encode the result to base64 to ensure safe transmission
	encodedResult := base64.StdEncoding.EncodeToString([]byte(currentResult))

	return &proto.JobStatusResponse{
		JobId:  req.JobId,
		Status: currentStatus,
		Result: encodedResult,
	}, nil
}
