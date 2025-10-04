//go:build freebsd
// +build freebsd

package ifc

import (
	"testing"
)

// TestGetStats tests getting interface statistics
func TestGetStats(t *testing.T) {
	// Test with loopback interface (always exists)
	stats, err := GetStats("lo0")
	if err != nil {
		t.Fatalf("GetStats(lo0) failed: %v", err)
	}

	// Stats should be non-nil
	if stats == nil {
		t.Fatal("GetStats returned nil stats")
	}

	// All counters should be >= 0 (uint64 so always non-negative)
	// Just verify the structure is populated correctly
	t.Logf("lo0 stats: InPackets=%d, InBytes=%d, OutPackets=%d, OutBytes=%d",
		stats.InPackets, stats.InBytes, stats.OutPackets, stats.OutBytes)
}

// TestGetStatsNonExistent tests error handling for non-existent interface
func TestGetStatsNonExistent(t *testing.T) {
	_, err := GetStats("nonexistent999")
	if err != ErrNotFound {
		t.Errorf("GetStats(nonexistent) should return ErrNotFound, got: %v", err)
	}
}

// TestGetStatsRealInterface tests with a real network interface if available
func TestGetStatsRealInterface(t *testing.T) {
	// Try to find a real interface (not lo0)
	ifaces, err := List()
	if err != nil {
		t.Skipf("Cannot list interfaces: %v", err)
	}

	var realIface string
	for _, iface := range ifaces {
		if iface.Name != "lo0" && !iface.Flags.IsLoopback() {
			realIface = iface.Name
			break
		}
	}

	if realIface == "" {
		t.Skip("No real network interface found")
	}

	stats, err := GetStats(realIface)
	if err != nil {
		t.Fatalf("GetStats(%s) failed: %v", realIface, err)
	}

	t.Logf("%s stats:", realIface)
	t.Logf("  RX: %d packets, %d bytes, %d errors, %d dropped",
		stats.InPackets, stats.InBytes, stats.InErrors, stats.InDropped)
	t.Logf("  TX: %d packets, %d bytes, %d errors",
		stats.OutPackets, stats.OutBytes, stats.OutErrors)
	t.Logf("  Multicast: %d, Collisions: %d",
		stats.InMulticast, stats.Collisions)
}
