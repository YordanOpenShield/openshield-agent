package tools

import "openshield-agent/internal/executor"

type ClamAVTool struct {
	*ScriptTool
}

// Action Exec functions for ClamAV
func install(opts []string) (string, error) {
	output, err := executor.ExecuteScript(ClamAV.script, []string{"install"})
	if err != nil {
		return "", err
	}
	return output, nil
}

func scan(opts []string) (string, error) {
	output, err := executor.ExecuteScript(ClamAV.script, []string{"scan"})
	if err != nil {
		return "", err
	}
	return output, nil
}

func uninstall(opts []string) (string, error) {
	output, err := executor.ExecuteScript(ClamAV.script, []string{"uninstall"})
	if err != nil {
		return "", err
	}
	return output, nil
}

var ClamAV = &ClamAVTool{}

func init() {
	ClamAV.ScriptTool = &ScriptTool{
		Tool: &Tool{
			Name: "clamav",
			Actions: []Action{
				{
					Name: "install",
					Opts: []string{},
					Exec: install,
				},
				{
					Name: "scan",
					Opts: []string{},
					Exec: scan,
				},
				{
					Name: "uninstall",
					Opts: []string{},
					Exec: uninstall,
				},
			},
			OS: []string{"linux", "darwin"},
		},
		script: "clamav.sh",
	}

	RegisterTool(*ClamAV.Tool) // Register the ClamAV tool
}
