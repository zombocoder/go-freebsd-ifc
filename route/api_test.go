//go:build freebsd
// +build freebsd

package route

import (
	"net"
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

// TestAddDelRoute4 tests IPv4 route addition and deletion
func TestAddDelRoute4(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Use lo0 for testing
	iface := "lo0"
	gw := net.ParseIP("127.0.0.1")
	_, dst, _ := net.ParseCIDR("198.51.100.0/24")

	// Add route
	if err := AddRoute4(dst, gw, iface); err != nil {
		t.Fatalf("AddRoute4() failed: %v", err)
	}

	// Add again (should be idempotent)
	if err := AddRoute4(dst, gw, iface); err != nil {
		t.Errorf("AddRoute4() should be idempotent, got error: %v", err)
	}

	// Delete route
	if err := DelRoute4(dst, gw, iface); err != nil {
		t.Errorf("DelRoute4() failed: %v", err)
	}

	// Delete again (should be idempotent)
	if err := DelRoute4(dst, gw, iface); err != nil {
		t.Errorf("DelRoute4() should be idempotent, got error: %v", err)
	}
}

// TestInvalidGateway tests error handling for invalid gateway
func TestInvalidGateway(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	_, dst, _ := net.ParseCIDR("192.0.2.0/24")

	// IPv6 gateway for IPv4 route (should fail)
	gw := net.ParseIP("::1")

	err := AddRoute4(dst, gw, "lo0")
	if err == nil {
		t.Error("AddRoute4() should fail with IPv6 gateway")
	}
}

// TestInvalidDestination tests error handling for invalid destination
func TestInvalidDestination(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	gw := net.ParseIP("127.0.0.1")

	// Nil destination
	err := AddRoute4(nil, gw, "lo0")
	if err == nil {
		t.Error("AddRoute4() should fail with nil destination")
	}
}

// TestInvalidInterface tests operations on non-existent interface
func TestInvalidInterface(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	gw := net.ParseIP("127.0.0.1")
	_, dst, _ := net.ParseCIDR("203.0.113.0/24")

	err := AddRoute4(dst, gw, "nonexistent999")
	if err == nil {
		t.Error("AddRoute4() should fail with non-existent interface")
	}
}

// TestAddDelRoute6 tests IPv6 route addition and deletion
func TestAddDelRoute6(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Use lo0 for testing
	iface := "lo0"
	gw := net.ParseIP("::1")
	_, dst, _ := net.ParseCIDR("2001:db8::/32")

	// Add route
	if err := AddRoute6(dst, gw, iface); err != nil {
		t.Fatalf("AddRoute6() failed: %v", err)
	}

	// Add again (should be idempotent)
	if err := AddRoute6(dst, gw, iface); err != nil {
		t.Errorf("AddRoute6() should be idempotent, got error: %v", err)
	}

	// Delete route
	if err := DelRoute6(dst, gw, iface); err != nil {
		t.Errorf("DelRoute6() failed: %v", err)
	}

	// Delete again (should be idempotent)
	if err := DelRoute6(dst, gw, iface); err != nil {
		t.Errorf("DelRoute6() should be idempotent, got error: %v", err)
	}
}

// TestInvalidGateway6 tests error handling for invalid IPv6 gateway
func TestInvalidGateway6(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	_, dst, _ := net.ParseCIDR("2001:db8:1::/48")

	// IPv4 gateway for IPv6 route (should fail)
	gw := net.ParseIP("127.0.0.1")

	err := AddRoute6(dst, gw, "lo0")
	if err == nil {
		t.Error("AddRoute6() should fail with IPv4 gateway")
	}
}

// TestInvalidDestination6 tests error handling for invalid IPv6 destination
func TestInvalidDestination6(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	gw := net.ParseIP("::1")

	// IPv4 destination for IPv6 route (should fail)
	_, dst, _ := net.ParseCIDR("192.0.2.0/24")

	err := AddRoute6(dst, gw, "lo0")
	if err == nil {
		t.Error("AddRoute6() should fail with IPv4 destination")
	}
}
