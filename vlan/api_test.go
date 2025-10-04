//go:build freebsd
// +build freebsd

package vlan

import (
	"os"
	"testing"
)

func skipIfNotRoot(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}
}

func skipIfNotE2E(t *testing.T) {
	if os.Getenv("IFCLIB_E2E") != "1" {
		t.Skip("E2E tests disabled. Set IFCLIB_E2E=1 to enable")
	}
}

// TestCreateDestroy tests VLAN creation and destruction
func TestCreateDestroy(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Create VLAN
	vl, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if vl == "" {
		t.Error("Create() returned empty VLAN name")
	}

	// Destroy VLAN
	if err := Destroy(vl); err != nil {
		t.Errorf("Destroy(%s) failed: %v", vl, err)
	}
}

// TestConfigure tests VLAN configuration
func TestConfigure(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	vl, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(vl)

	// Configure VLAN tag 100 on lo0 (using lo0 since it always exists)
	if err := Configure(vl, 100, "lo0"); err != nil {
		t.Errorf("Configure(%s, 100, lo0) failed: %v", vl, err)
	}

	// Get configuration
	cfg, err := Get(vl)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", vl, err)
	}

	if cfg.Tag != 100 {
		t.Errorf("Expected tag 100, got %d", cfg.Tag)
	}

	if cfg.Parent != "lo0" {
		t.Errorf("Expected parent 'lo0', got '%s'", cfg.Parent)
	}
}

// TestUpDown tests bringing VLAN up and down
func TestUpDown(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	vl, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(vl)

	// Configure first (required before bringing up)
	if err := Configure(vl, 200, "lo0"); err != nil {
		t.Fatalf("Configure() failed: %v", err)
	}

	// Bring up
	if err := Up(vl, true); err != nil {
		t.Errorf("Up(%s, true) failed: %v", vl, err)
	}

	// Bring down
	if err := Up(vl, false); err != nil {
		t.Errorf("Up(%s, false) failed: %v", vl, err)
	}
}

// TestGetNonExistent tests getting non-existent VLAN
func TestGetNonExistent(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	_, err := Get("vlan999999")
	if err == nil {
		t.Error("Get() should fail for non-existent VLAN")
	}
}

// TestInvalidTag tests invalid VLAN tags
func TestInvalidTag(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	vl, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(vl)

	// Tag 0 is typically invalid
	err = Configure(vl, 0, "lo0")
	if err == nil {
		t.Error("Configure() with tag 0 should fail")
	}

	// Tag > 4094 is invalid for 802.1Q
	err = Configure(vl, 5000, "lo0")
	if err == nil {
		t.Error("Configure() with tag 5000 should fail")
	}
}
