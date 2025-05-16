package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func getOsqueryPath() string {
	baseDir := filepath.Dir(os.Args[0])
	if runtime.GOOS == "windows" {
		return filepath.Join(baseDir, "bin", "osqueryi.exe")
	}
	return filepath.Join(baseDir, "bin", "osqueryi")
}

func checkOsqueryAvailable() error {
	osqueryPath := getOsqueryPath()
	cmd := exec.Command(osqueryPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("osquery not working: %v, output: %s", err, string(output))
	}
	return nil
}

func runOsquery(query string) (string, error) {
	osqueryPath := getOsqueryPath()
	cmd := exec.Command(osqueryPath, "--json", query)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run osquery: %v\nOutput: %s", err, string(output))
	}
	return string(output), nil
}
