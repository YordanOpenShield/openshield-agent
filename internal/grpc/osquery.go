package agentgrpc

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"openshield-agent/internal/executor"
	"openshield-agent/proto"
)

// Osquery query state management
var (
	osqueryResults = make(map[string]*osqueryQueryState)
	osqueryMu      sync.RWMutex
)

type osqueryQueryState struct {
	status proto.TaskStatus
	result string
}

// RunOsquery handles incoming osquery queries from the manager
func (s *AgentServer) RunOsquery(ctx context.Context, req *proto.OsqueryQueryRequest) (*proto.OsqueryQueryResponse, error) {
	log.Printf("[AGENT] Received osquery: %s - %s", req.QueryId, req.Query)

	// Initialize query state
	osqueryMu.Lock()
	osqueryResults[req.QueryId] = &osqueryQueryState{
		status: proto.TaskStatus_RUNNING,
		result: "",
	}
	osqueryMu.Unlock()

	// Execute query in background
	go func() {
		results, err := executor.RunOSQuery(req.Query)

		osqueryMu.Lock()
		defer osqueryMu.Unlock()

		if err != nil {
			log.Printf("[AGENT] Osquery %s failed: %v", req.QueryId, err)
			osqueryResults[req.QueryId] = &osqueryQueryState{
				status: proto.TaskStatus_FAILED,
				result: err.Error(),
			}
			return
		}

		// Serialize results to JSON
		jsonResult, err := json.Marshal(results)
		if err != nil {
			log.Printf("[AGENT] Failed to marshal osquery results: %v", err)
			osqueryResults[req.QueryId] = &osqueryQueryState{
				status: proto.TaskStatus_FAILED,
				result: err.Error(),
			}
			return
		}

		log.Printf("[AGENT] Osquery %s completed with %d rows", req.QueryId, len(results))
		osqueryResults[req.QueryId] = &osqueryQueryState{
			status: proto.TaskStatus_COMPLETED,
			result: string(jsonResult),
		}
	}()

	return &proto.OsqueryQueryResponse{
		Accepted: true,
		QueryId:  req.QueryId,
		Message:  "Query accepted and executing",
	}, nil
}

// GetOsqueryResult returns the status and results of an osquery query
func (s *AgentServer) GetOsqueryResult(ctx context.Context, req *proto.OsqueryResultRequest) (*proto.OsqueryResultResponse, error) {
	log.Printf("[AGENT] Reporting osquery result for %s", req.QueryId)

	osqueryMu.RLock()
	state, exists := osqueryResults[req.QueryId]
	osqueryMu.RUnlock()

	if !exists {
		return &proto.OsqueryResultResponse{
			QueryId: req.QueryId,
			Status:  proto.TaskStatus_FAILED,
			Result:  "Query ID not found",
		}, nil
	}

	return &proto.OsqueryResultResponse{
		QueryId: req.QueryId,
		Status:  state.status,
		Result:  state.result,
	}, nil
}
