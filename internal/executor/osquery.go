package executor

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

var osqueryPath string

// SetOsqueryPath sets the path to the osquery binary
func SetOsqueryPath(path string) {
	osqueryPath = path
}

func checkOsqueryAvailable() error {
	cmd := exec.Command(osqueryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("osquery not working: %v, output: %s", err, string(output))
	}
	return nil
}

func RunOSQuery(query string) ([]map[string]interface{}, error) {
	if err := checkOsqueryAvailable(); err != nil {
		return nil, err
	}
	cmd := exec.Command(osqueryPath, "--json", query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run osquery: %v\nOutput: %s", err, string(output))
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse osquery output: %v\nOutput: %s", err, string(output))
	}
	return result, nil
}
