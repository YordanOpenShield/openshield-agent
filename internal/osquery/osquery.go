package osquery

import (
	"fmt"
	"openshield-agent/internal/executor"
)

// GetAllLocalAddresses returns all non-loopback, non-APIPA IPv4 addresses.
func GetAllLocalAddresses() ([]string, error) {
	var addresses []string

	query := "SELECT address FROM interface_addresses WHERE address NOT LIKE '127.%' AND address NOT LIKE '169.254.%' AND address LIKE '%.%';"
	results, err := executor.RunOSQuery(query)
	if err == nil && len(results) > 0 {
		for _, row := range results {
			if addr, ok := row["address"]; ok {
				if addrStr, ok := addr.(string); ok && addrStr != "" {
					addresses = append(addresses, addrStr)
				}
			}
		}
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no non-loopback addresses found")
	}
	return addresses, nil
}

func GetAllServicesStates() (map[string]string, error) {
	services := make(map[string]string)

	query := "SELECT name, state FROM services;"
	results, err := executor.RunOSQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}

	for _, row := range results {
		name, nameOk := row["name"].(string)
		state, stateOk := row["state"].(string)
		if nameOk && stateOk {
			services[name] = state
		}
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no services found")
	}
	return services, nil
}
