package agentgrpc

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"openshield-agent/internal/executor"
	"openshield-agent/proto"
)

// ExecuteQuery handles incoming queries from the manager
func (s *AgentServer) ExecuteQuery(ctx context.Context, req *proto.QueryExecutionRequest) (*proto.QueryExecutionResponse, error) {
	executionId := req.ExecutionId
	sql := req.Query.Sql
	log.Printf("[AGENT] Received query: %s - %s", executionId, sql)

	// Execute query immediately
	result, err := executeOsquery(sql)

	// Report results back to manager
	var resultJson string
	var success bool
	var errorMsg string

	if err != nil {
		log.Printf("[AGENT] Query %s failed: %v", executionId, err)
		resultJson = "[]"
		success = false
		errorMsg = err.Error()
	} else {
		log.Printf("[AGENT] Query %s completed successfully", executionId)
		resultJson = result
		success = true
		errorMsg = ""
	}

	// Report results to manager if client is available
	if s.managerClient != nil {
		reportCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, reportErr := s.managerClient.ReportQueryResult(reportCtx, &proto.QueryResultRequest{
			ExecutionId: executionId,
			ResultJson:  resultJson,
			Success:     success,
			Error:       errorMsg,
		})
		if reportErr != nil {
			log.Printf("[AGENT] Failed to report query result: %v", reportErr)
		} else {
			log.Printf("[AGENT] Reported query result for %s", executionId)
		}
	} else {
		log.Printf("[AGENT] Warning: No manager client available to report results")
	}

	return &proto.QueryExecutionResponse{
		Accepted: true,
		Message:  "Query executed",
	}, nil
}

// StreamQueryResults is called by manager to check query status (kept for compatibility)
func (s *AgentServer) StreamQueryResults(ctx context.Context, req *proto.QueryResultRequest) (*proto.QueryResultResponse, error) {
	log.Printf("[AGENT] StreamQueryResults called for %s", req.ExecutionId)
	// Results are now reported proactively via ReportQueryResult
	return &proto.QueryResultResponse{
		Received: true,
	}, nil
}

// executeOsquery runs an osquery SQL query
func executeOsquery(query string) (string, error) {
	results, err := executor.RunOSQuery(query)
	if err != nil {
		return "", err
	}

	jsonResult, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	return string(jsonResult), nil
}

// executeSQLite runs a SQLite query (placeholder for future implementation)
func executeSQLite(query string) (string, error) {
	// TODO: Implement SQLite execution
	return "", nil
}

// executeNative runs a native/system query (placeholder for future implementation)
func executeNative(query string) (string, error) {
	// TODO: Implement native query execution
	return "", nil
}
