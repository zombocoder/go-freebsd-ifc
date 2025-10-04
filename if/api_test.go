//go:build freebsd
// +build freebsd

package ifc

import (
	"testing"
)

// TestList tests the List function
func TestList(t *testing.T) {
	ifaces, err := List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(ifaces) == 0 {
		t.Error("List() returned no interfaces, expected at least lo0")
	}

	// Check that lo0 exists (should always be present)
	foundLoopback := false
	for _, iface := range ifaces {
		if iface.Name == "lo0" {
			foundLoopback = true
			if !iface.Flags.IsLoopback() {
				t.Error("lo0 does not have LOOPBACK flag")
			}
			break
		}
	}

	if !foundLoopback {
		t.Error("lo0 interface not found")
	}
}

// TestGet tests the Get function
func TestGet(t *testing.T) {
	// Test getting lo0 (should always exist)
	iface, err := Get("lo0")
	if err != nil {
		t.Fatalf("Get(lo0) failed: %v", err)
	}

	if iface.Name != "lo0" {
		t.Errorf("Expected name 'lo0', got '%s'", iface.Name)
	}

	if !iface.Flags.IsLoopback() {
		t.Error("lo0 does not have LOOPBACK flag")
	}

	// Test getting non-existent interface
	_, err = Get("nonexistent999")
	if err == nil {
		t.Error("Get(nonexistent999) should return error")
	}
}

// TestFlagsIsUp tests the InterfaceFlags.IsUp method
func TestFlagsIsUp(t *testing.T) {
	flags := InterfaceFlags(FlagUp)
	if !flags.IsUp() {
		t.Error("Flags with FlagUp should return IsUp() = true")
	}

	flags = InterfaceFlags(0)
	if flags.IsUp() {
		t.Error("Flags without FlagUp should return IsUp() = false")
	}
}

// TestFlagsIsLoopback tests the InterfaceFlags.IsLoopback method
func TestFlagsIsLoopback(t *testing.T) {
	flags := InterfaceFlags(FlagLoopback)
	if !flags.IsLoopback() {
		t.Error("Flags with FlagLoopback should return IsLoopback() = true")
	}

	flags = InterfaceFlags(0)
	if flags.IsLoopback() {
		t.Error("Flags without FlagLoopback should return IsLoopback() = false")
	}
}

// TestFlagsIsRunning tests the InterfaceFlags.IsRunning method
func TestFlagsIsRunning(t *testing.T) {
	flags := InterfaceFlags(FlagRunning)
	if !flags.IsRunning() {
		t.Error("Flags with FlagRunning should return IsRunning() = true")
	}

	flags = InterfaceFlags(0)
	if flags.IsRunning() {
		t.Error("Flags without FlagRunning should return IsRunning() = false")
	}
}
