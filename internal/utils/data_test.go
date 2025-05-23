package utils

import (
	"testing"
)

func TestDetectOS(t *testing.T) {
	os := GetDeviceOS()
	if os == "" {
		t.Error("DetectOS returned an empty string")
	}
}

func TestGetDeviceID(t *testing.T) {
	id, err := GetDeviceID()
	if err != nil {
		t.Errorf("GetDeviceID returned error: %v", err)
	}
	if id == "" {
		t.Error("GetDeviceID returned an empty ID")
	}
}

func TestGetAllLocalAddresses(t *testing.T) {
	addrs, err := GetAllLocalAddresses()
	if err != nil {
		t.Errorf("GetAllLocalAddresses returned error: %v", err)
	}
	if len(addrs) == 0 {
		t.Error("GetAllLocalAddresses returned no addresses")
	}
}
