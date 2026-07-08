package utils

import (
	"fmt"
	"log"
	"net"
	"openshield-agent/internal/models"
	"openshield-agent/internal/osquery"
	"os/exec"
	"runtime"
	"strings"

	"github.com/denisbrodbeck/machineid"
)

// DetectOS returns the name of the operating system (e.g., "windows", "linux", "darwin").
func GetDeviceOS() string {
	return runtime.GOOS
}

// GetDeviceID retrieves the unique machine/device ID using the machineid package.
func GetDeviceID() (string, error) {
	id, err := machineid.ID()
	if err != nil {
		return "", fmt.Errorf("failed to get device ID: %w", err)
	}
	return id, nil
}

// GetAllLocalAddresses returns all non-loopback, non-APIPA IPv4 addresses.
func GetAllLocalAddresses() ([]string, error) {
	var addresses []string
	addressMap := make(map[string]bool) // deduplication

	// Method 1: Try using osquery first
	osqueryAddrs, _ := osquery.GetAllLocalAddresses()
	for _, addr := range osqueryAddrs {
		if !addressMap[addr] {
			addressMap[addr] = true
			addresses = append(addresses, addr)
		}
	}

	// Method 2: Get all interface addresses including VPN/tunnel interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}
	for _, iface := range ifaces {
		// Skip loopback interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		// Include all interfaces (even down ones, to capture VPN IPs)
		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				ip := ipnet.IP.To4()
				// Skip APIPA addresses (169.254.x.x)
				if ip[0] == 169 && ip[1] == 254 {
					continue
				}
				ipStr := ip.String()
				if !addressMap[ipStr] {
					addressMap[ipStr] = true
					addresses = append(addresses, ipStr)
					log.Printf("[AGENT] Found IP: %s on interface %s", ipStr, iface.Name)
				}
			}
		}
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no non-loopback addresses found")
	}
	return addresses, nil
}

// GetAllServices returns a list of all services on the machine and their states.
func GetAllServices() ([]models.Service, error) {
	var cmd *exec.Cmd
	osType := runtime.GOOS

	switch osType {
	case "windows":
		// Use 'Get-Service' to get all services and their states
		cmd = exec.Command("powershell", "-Command", "Get-Service | Select-Object Status,Name")
	case "linux":
		// Try systemctl first
		cmd = exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--no-legend")
	case "darwin":
		// Use 'launchctl list' to get all services (macOS)
		cmd = exec.Command("launchctl", "list")
	default:
		return nil, fmt.Errorf("unsupported OS: %s", osType)
	}

	output, err := cmd.Output()
	if err != nil && osType == "linux" {
		// Fallback: Try 'service --status-all' if systemctl fails
		cmd = exec.Command("service", "--status-all")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to list services with both systemctl and service: %w", err)
		}
	}

	var services []models.Service

	switch osType {
	case "windows":
		// Parse output for Windows
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			// Skip header lines (first 2 lines)
			if i < 3 {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				state := fields[0]
				name := fields[1]
				services = append(services, models.Service{Name: name, State: strings.ToUpper(state)})
			}
		}
	case "linux":
		// Parse output for Linux
		if strings.Contains(cmd.String(), "systemctl") {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) >= 4 {
					name := fields[0]
					state := fields[3]
					services = append(services, models.Service{Name: name, State: state})
				}
			}
		} else {
			// Fallback: Parse 'service --status-all' output
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if len(line) > 0 {
					// Format: [ + ]  service_name
					parts := strings.Fields(line)
					if len(parts) >= 4 {
						state := parts[1]
						name := parts[3]
						services = append(services, models.Service{Name: name, State: state})
					}
				}
			}
		}
	case "darwin":
		// Parse output for macOS
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				name := fields[2]
				state := "Unknown"
				if fields[0] == "-" {
					state = "Stopped"
				} else if fields[0] == "0" {
					state = "Running"
				}
				services = append(services, models.Service{Name: name, State: state})
			}
		}
	}

	return services, nil
}
