//go:build freebsd
// +build freebsd

package tun

import (
	"os"
	"strings"
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

// TestCreateDestroy tests TUN interface creation and destruction
func TestCreateDestroy(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Create TUN interface
	name, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if !strings.HasPrefix(name, "tun") {
		t.Errorf("Expected interface name to start with 'tun', got: %s", name)
	}

	// Destroy TUN interface
	if err := Destroy(name); err != nil {
		t.Errorf("Destroy() failed: %v", err)
	}

	// Destroy again (should be idempotent)
	if err := Destroy(name); err != nil {
		t.Logf("Destroy() second call returned: %v (expected)", err)
	}
}

// TestUp tests bringing TUN interface up/down
func TestUp(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(name)

	// Bring up
	if err := Up(name, true); err != nil {
		t.Errorf("Up(true) failed: %v", err)
	}

	// Bring down
	if err := Up(name, false); err != nil {
		t.Errorf("Up(false) failed: %v", err)
	}
}

// TestUpNonExistent tests Up on non-existent interface
func TestUpNonExistent(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	err := Up("tun999", true)
	if err == nil {
		t.Error("Up() should fail for non-existent interface")
	}
}
