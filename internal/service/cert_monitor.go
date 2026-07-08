package service

import (
	"log"
	"openshield-agent/internal/config"
	"openshield-agent/internal/utils"
	"os"
	"sort"
	"time"
)

var knownAddresses []string

// StartCertMonitor starts a background goroutine that monitors for network interface changes
// and requests certificate updates when new IPs are detected
func StartCertMonitor(interval time.Duration, stopCh <-chan struct{}) {
	// Get initial addresses
	addrs, err := utils.GetAllLocalAddresses()
	if err != nil {
		log.Printf("[CERT MONITOR] Failed to get initial addresses: %v", err)
		return
	}
	knownAddresses = sortAddresses(addrs)
	log.Printf("[CERT MONITOR] Initial addresses: %v", knownAddresses)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				log.Print("[CERT MONITOR] Stopping certificate monitor")
				return
			case <-ticker.C:
				checkForAddressChanges()
			}
		}
	}()
}

func checkForAddressChanges() {
	currentAddrs, err := utils.GetAllLocalAddresses()
	if err != nil {
		log.Printf("[CERT MONITOR] Failed to get current addresses: %v", err)
		return
	}

	sortedCurrent := sortAddresses(currentAddrs)

	// Check if addresses have changed
	if addressesChanged(knownAddresses, sortedCurrent) {
		log.Printf("[CERT MONITOR] Address change detected!")
		log.Printf("[CERT MONITOR] Old: %v", knownAddresses)
		log.Printf("[CERT MONITOR] New: %v", sortedCurrent)

		// Find new addresses
		newAddrs := findNewAddresses(knownAddresses, sortedCurrent)
		if len(newAddrs) > 0 {
			log.Printf("[CERT MONITOR] New addresses found: %v", newAddrs)
			requestCertUpdate(sortedCurrent)
		}

		knownAddresses = sortedCurrent
	}
}

func sortAddresses(addrs []string) []string {
	sorted := make([]string, len(addrs))
	copy(sorted, addrs)
	sort.Strings(sorted)
	return sorted
}

func addressesChanged(old, new []string) bool {
	if len(old) != len(new) {
		return true
	}
	for i := range old {
		if old[i] != new[i] {
			return true
		}
	}
	return false
}

func findNewAddresses(old, current []string) []string {
	oldMap := make(map[string]bool)
	for _, addr := range old {
		oldMap[addr] = true
	}

	var newAddrs []string
	for _, addr := range current {
		if !oldMap[addr] {
			newAddrs = append(newAddrs, addr)
		}
	}
	return newAddrs
}

func requestCertUpdate(newAddresses []string) {
	log.Printf("[CERT MONITOR] Requesting certificate update with new addresses")

	// Generate new CSR with all addresses
	creds, err := utils.GetAgentCredentials()
	if err != nil {
		log.Printf("[CERT MONITOR] Failed to get credentials: %v", err)
		return
	}

	// Generate new CSR
	err = utils.GenerateCSR(creds.AgentID)
	if err != nil {
		log.Printf("[CERT MONITOR] Failed to generate CSR: %v", err)
		return
	}

	// Read CSR
	csrPath := config.CertsPath + "/agent.csr"
	csrData, err := os.ReadFile(csrPath)
	if err != nil {
		log.Printf("[CERT MONITOR] Failed to read CSR: %v", err)
		return
	}

	// Request new certificate
	certsResp, err := utils.RequestCSRSigning(csrData)
	if err != nil {
		log.Printf("[CERT MONITOR] Failed to request certificate: %v", err)
		return
	}

	// Save new certificates
	err = utils.SaveCertificates(certsResp)
	if err != nil {
		log.Printf("[CERT MONITOR] Failed to save certificates: %v", err)
		return
	}

	log.Printf("[CERT MONITOR] Certificate updated successfully with new addresses")
}
