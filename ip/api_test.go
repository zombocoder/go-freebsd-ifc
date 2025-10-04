//go:build freebsd
// +build freebsd

package ip

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

// TestAdd4Del4 tests IPv4 address addition and deletion
func TestAdd4Del4(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Use lo0 for testing (always exists)
	iface := "lo0"
	ip := net.ParseIP("127.0.0.2")
	mask := net.CIDRMask(32, 32)

	// Add address
	if err := Add4(iface, ip, mask); err != nil {
		t.Fatalf("Add4() failed: %v", err)
	}

	// Add again (should be idempotent)
	if err := Add4(iface, ip, mask); err != nil {
		t.Errorf("Add4() should be idempotent, got error: %v", err)
	}

	// Delete address
	if err := Del4(iface, ip, mask); err != nil {
		t.Errorf("Del4() failed: %v", err)
	}

	// Delete again (should be idempotent)
	if err := Del4(iface, ip, mask); err != nil {
		t.Errorf("Del4() should be idempotent, got error: %v", err)
	}
}

// TestAdd6Del6 tests IPv6 address addition and deletion
func TestAdd6Del6(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	iface := "lo0"
	ip := net.ParseIP("::2")
	prefixLen := 128

	// Add address
	if err := Add6(iface, ip, prefixLen); err != nil {
		t.Fatalf("Add6() failed: %v", err)
	}

	// Add again (should be idempotent)
	if err := Add6(iface, ip, prefixLen); err != nil {
		t.Errorf("Add6() should be idempotent, got error: %v", err)
	}

	// Delete address
	if err := Del6(iface, ip, prefixLen); err != nil {
		t.Errorf("Del6() failed: %v", err)
	}

	// Delete again (should be idempotent)
	if err := Del6(iface, ip, prefixLen); err != nil {
		t.Errorf("Del6() should be idempotent, got error: %v", err)
	}
}

// TestInvalidIPv4 tests error handling for invalid IPv4 addresses
func TestInvalidIPv4(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Test with IPv6 address (should fail)
	ip := net.ParseIP("::1")
	mask := net.CIDRMask(24, 32)

	err := Add4("lo0", ip, mask)
	if err == nil {
		t.Error("Add4() should fail with IPv6 address")
	}
}

// TestInvalidIPv6 tests error handling for invalid IPv6 addresses
func TestInvalidIPv6(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Test with IPv4 address (should fail)
	ip := net.ParseIP("127.0.0.1")

	err := Add6("lo0", ip, 64)
	if err == nil {
		t.Error("Add6() should fail with IPv4 address")
	}
}

// TestInvalidInterface tests operations on non-existent interface
func TestInvalidInterface(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	ip := net.ParseIP("192.168.1.1")
	mask := net.CIDRMask(24, 32)

	err := Add4("nonexistent999", ip, mask)
	if err == nil {
		t.Error("Add4() should fail with non-existent interface")
	}
}

// TestInvalidMask tests error handling for invalid network masks
func TestInvalidMask(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	ip := net.ParseIP("192.168.1.1")

	// IPv6 mask for IPv4 address (wrong size)
	mask := net.CIDRMask(64, 128)

	err := Add4("lo0", ip, mask)
	if err == nil {
		t.Error("Add4() should fail with incorrect mask size")
	}
}

// TestInvalidPrefixLength tests error handling for invalid prefix lengths
func TestInvalidPrefixLength(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	ip := net.ParseIP("::1")

	// Invalid prefix length
	err := Add6("lo0", ip, 200)
	if err == nil {
		t.Error("Add6() should fail with invalid prefix length")
	}

	err = Add6("lo0", ip, -1)
	if err == nil {
		t.Error("Add6() should fail with negative prefix length")
	}
}
