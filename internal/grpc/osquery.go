package agentgrpc

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"openshield-agent/internal/executor"
	"openshield-agent/proto"
)

// Query execution state management
var (
	queryResults = make(map[string]*queryExecutionState)
	queryMu      sync.RWMutex
)

type queryExecutionState struct {
	success    bool
	resultJSON string
	err        string
}

// ExecuteQuery handles incoming osquery queries from the manager
func (s *AgentServer) ExecuteQuery(ctx context.Context, req *proto.QueryExecutionRequest) (*proto.QueryExecutionResponse, error) {
	log.Printf("[AGENT] Received query: %s - %s", req.ExecutionId, req.Query.Sql)

	// Initialize query state
	queryMu.Lock()
	queryResults[req.ExecutionId] = &queryExecutionState{
		success:    false,
		resultJSON: "",
		err:        "",
	}
	queryMu.Unlock()

	// Execute query in background
	go func() {
		results, err := executor.RunOSQuery(req.Query.Sql)

		queryMu.Lock()
		defer queryMu.Unlock()

		if err != nil {
			log.Printf("[AGENT] Query %s failed: %v", req.ExecutionId, err)
			queryResults[req.ExecutionId] = &queryExecutionState{
				success:    false,
				resultJSON: "",
				err:        err.Error(),
			}
			return
		}

		// Serialize results to JSON
		jsonResult, err := json.Marshal(results)
		if err != nil {
			log.Printf("[AGENT] Failed to marshal query results: %v", err)
			queryResults[req.ExecutionId] = &queryExecutionState{
				success:    false,
				resultJSON: "",
				err:        err.Error(),
			}
			return
		}

		log.Printf("[AGENT] Query %s completed with %d rows", req.ExecutionId, len(results))
		queryResults[req.ExecutionId] = &queryExecutionState{
			success:    true,
			resultJSON: string(jsonResult),
			err:        "",
		}
	}()

	return &proto.QueryExecutionResponse{
		Accepted: true,
		Message:  "Query accepted and executing",
	}, nil
}

// StreamQueryResults returns the result of a query execution
func (s *AgentServer) StreamQueryResults(ctx context.Context, req *proto.QueryResultRequest) (*proto.QueryResultResponse, error) {
	log.Printf("[AGENT] Reporting query result for %s", req.ExecutionId)

	queryMu.RLock()
	state, exists := queryResults[req.ExecutionId]
	queryMu.RUnlock()

	if !exists {
		return &proto.QueryResultResponse{
			Received: false,
		}, nil
	}

	// Update the request with the results (agent reports back to manager)
	req.ResultJson = state.resultJSON
	req.Success = state.success
	req.Error = state.err

	return &proto.QueryResultResponse{
		Received: true,
	}, nil
}
