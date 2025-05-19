package utils

import (
	"fmt"
	"net"
	"openshield-agent/internal/osquery"
	"runtime"

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

	// Try using osquery first
	addresses, err := osquery.GetAllLocalAddresses()

	// If osquery fails or returns no addresses, fallback to net.InterfaceAddrs
	if len(addresses) == 0 || err != nil {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return nil, fmt.Errorf("failed to get network interfaces: %w", err)
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ip := ipnet.IP.To4()
				// Skip APIPA addresses (169.254.x.x)
				if ip[0] == 169 && ip[1] == 254 {
					continue
				}
				addresses = append(addresses, ip.String())
			}
		}
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no non-loopback addresses found")
	}
	return addresses, nil
}
