package tools

import (
	"fmt"
	"os"
	"os/exec"
)

type Fail2BanTool struct {
	*Tool
}

var Fail2Ban = &Fail2BanTool{
	Tool: &Tool{
		Name: "fail2ban",
		Actions: []Action{
			{
				Name: "install",
				Opts: []string{},
				Exec: func(opts []string) (string, error) {
					// Detect distro by reading /etc/os-release
					osRelease, err := os.ReadFile("/etc/os-release")
					if err != nil {
						return "", fmt.Errorf("could not read /etc/os-release: %w", err)
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
						return "", fmt.Errorf("unsupported Linux distribution: %s", id)
					}

					for _, cmd := range cmds {
						output, err := cmd.CombinedOutput()
						if err != nil {
							return "", fmt.Errorf("command %v failed: %w\nOutput: %s", cmd.Args, err, string(output))
						}
					}
					return "", nil
				},
			},
			{
				Name: "configure",
				Opts: []string{"ssh", "web", "mail"},
				Exec: func(opts []string) (string, error) {
					var output string
					for _, opt := range opts {
						switch opt {
						case "ssh":
							cmd := exec.Command("sudo", "bash", "-c", `echo -e "[sshd]\nenabled = true" | sudo tee /etc/fail2ban/jail.d/sshd.conf`)
							output, err := cmd.CombinedOutput()
							if err != nil {
								return "", fmt.Errorf("failed to configure ssh jail: %w\nOutput: %s", err, string(output))
							}
						case "web":
							webJails := `[nginx-http-auth]
enabled = true
[nginx-botsearch]
enabled = true
[nginx-limit-req]
enabled = true
[nginx-req-limit]
enabled = true
[nginx-noscript]
enabled = true
[nginx-nohome]
enabled = true
[nginx-badbots]
enabled = true
[apache-auth]
enabled = true
[apache-badbots]
enabled = true
[apache-noscript]
enabled = true
[apache-overflows]
enabled = true
[apache-nohome]
enabled = true
[apache-shellshock]
enabled = true
`
							cmd := exec.Command("sudo", "bash", "-c", fmt.Sprintf("echo -e '%s' | sudo tee /etc/fail2ban/jail.d/web.conf", webJails))
							output, err := cmd.CombinedOutput()
							if err != nil {
								return "", fmt.Errorf("failed to configure web jail: %w\nOutput: %s", err, string(output))
							}
						case "mail":
							cmd := exec.Command("sudo", "bash", "-c", `echo -e "[dovecot]\nenabled = true\n[postfix]\nenabled = true" | sudo tee /etc/fail2ban/jail.d/mail.conf`)
							output, err := cmd.CombinedOutput()
							if err != nil {
								return "", fmt.Errorf("failed to configure mail jail: %w\nOutput: %s", err, string(output))
							}
						default:
							return "", fmt.Errorf("unsupported module: %s", opt)
						}
					}
					return string(output), nil
				},
			},
			{
				Name: "start",
				Opts: []string{},
				Exec: func(opts []string) (string, error) {
					checkCmd := exec.Command("sudo", "systemctl", "is-active", "fail2ban")
					output, err := checkCmd.CombinedOutput()
					if err != nil {
						return string(output), fmt.Errorf("failed to check fail2ban service status: %w", err)
					}

					if output != nil && string(output) == "active\n" {
						// Service is running, restart it
						restartCmd := exec.Command("sudo", "systemctl", "restart", "fail2ban")
						output, err := restartCmd.CombinedOutput()
						if err != nil {
							return string(output), fmt.Errorf("failed to restart fail2ban service: %w", err)
						}
						return string(output), nil
					}

					// Service is not running, start it
					startCmd := exec.Command("sudo", "systemctl", "start", "fail2ban")
					output, err = startCmd.CombinedOutput()
					if err != nil {
						return string(output), fmt.Errorf("failed to start fail2ban service: %w", err)
					}
					return string(output), nil
				},
			},
			{
				Name: "stop",
				Opts: []string{},
				Exec: func(opts []string) (string, error) {
					checkCmd := exec.Command("sudo", "systemctl", "is-active", "fail2ban")
					output, err := checkCmd.CombinedOutput()
					if err != nil {
						return string(output), fmt.Errorf("failed to check fail2ban service status: %w", err)
					}

					if output == nil || string(output) != "active\n" {
						// Service is not running, nothing to stop
						return string(output), nil
					}

					stopCmd := exec.Command("sudo", "systemctl", "stop", "fail2ban")
					output, err = stopCmd.CombinedOutput()
					if err != nil {
						return string(output), fmt.Errorf("failed to stop fail2ban service: %w", err)
					}
					return string(output), nil
				},
			},
			{
				Name: "uninstall",
				Opts: []string{},
				Exec: func(opts []string) (string, error) {
					var output string

					osRelease, err := os.ReadFile("/etc/os-release")
					if err != nil {
						return "", fmt.Errorf("could not read /etc/os-release: %w", err)
					}
					id := parseOSReleaseID(osRelease)

					var cmds []*exec.Cmd
					switch id {
					case "ubuntu", "debian", "kali":
						cmds = []*exec.Cmd{
							exec.Command("sudo", "apt-get", "remove", "-y", "fail2ban"),
							exec.Command("sudo", "apt-get", "autoremove", "-y"),
						}
					case "centos", "rhel", "fedora":
						cmds = []*exec.Cmd{
							exec.Command("sudo", "yum", "remove", "-y", "fail2ban"),
						}
					default:
						return "", fmt.Errorf("unsupported Linux distribution: %s", id)
					}

					for _, cmd := range cmds {
						output, err := cmd.CombinedOutput()
						if err != nil {
							return string(output), fmt.Errorf("command %v failed: %w", cmd.Args, err)
						}
					}
					return string(output), nil
				},
			},
		},
		OS: []string{"linux"},
	},
}

func init() {
	RegisterTool(*Fail2Ban.Tool)
}
