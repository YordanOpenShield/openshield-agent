package tools

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type Fail2BanTool struct {
	*Tool

	// Name is the name of the tool.
	// Name string `yaml:"name"`
}

var Fail2Ban = &Fail2BanTool{
	Tool: &Tool{
		Name: "fail2ban",
		Actions: []Action{
			{
				Name: "install",
				Opts: []string{},
			},
		},
		OS: []string{"linux"},
	},
}

func (f2b *Fail2BanTool) Install() error {
	if !f2b.isOSSupported(runtime.GOOS) {
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Detect distro by reading /etc/os-release
	osRelease, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return fmt.Errorf("could not read /etc/os-release: %w", err)
	}
	id := parseOSReleaseID(osRelease)

	var cmds []*exec.Cmd
	switch id {
	case "ubuntu", "debian", "kali":
		cmds = []*exec.Cmd{
			exec.Command("sudo", "apt-get", "update"),
			exec.Command("sudo", "apt-get", "install", "-y", "fail2ban"),
		}
	case "centos", "rhel", "fedora":
		cmds = []*exec.Cmd{
			exec.Command("sudo", "yum", "install", "-y", "epel-release"),
			exec.Command("sudo", "yum", "install", "-y", "fail2ban"),
		}
	default:
		return fmt.Errorf("unsupported Linux distribution: %s", id)
	}

	for _, cmd := range cmds {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command %v failed: %w\nOutput: %s", cmd.Args, err, string(output))
		}
	}
	return nil
}
