//go:build freebsd
// +build freebsd

package lagg

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

// TestCreateDestroy tests LAGG interface creation and destruction
func TestCreateDestroy(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Create LAGG interface
	name, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if !strings.HasPrefix(name, "lagg") {
		t.Errorf("Expected interface name to start with 'lagg', got: %s", name)
	}

	// Destroy LAGG interface
	if err := Destroy(name); err != nil {
		t.Errorf("Destroy() failed: %v", err)
	}
}

// TestSetProto tests setting LAGG protocol
func TestSetProto(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(name)

	// Test different protocols
	protocols := []Proto{
		ProtoFailover,
		ProtoLoadBalance,
		ProtoLACP,
		ProtoRoundRobin,
		ProtoBroadcast,
	}

	for _, proto := range protocols {
		if err := SetProto(name, proto); err != nil {
			t.Errorf("SetProto(%s) failed: %v", proto.String(), err)
		}
	}
}

// TestProtoString tests protocol string representation
func TestProtoString(t *testing.T) {
	tests := []struct {
		proto Proto
		want  string
	}{
		{ProtoFailover, "failover"},
		{ProtoLoadBalance, "loadbalance"},
		{ProtoLACP, "lacp"},
		{ProtoRoundRobin, "roundrobin"},
		{ProtoBroadcast, "broadcast"},
		{Proto(999), "unknown(999)"},
	}

	for _, tt := range tests {
		got := tt.proto.String()
		if got != tt.want {
			t.Errorf("Proto(%d).String() = %q, want %q", tt.proto, got, tt.want)
		}
	}
}

// TestUp tests bringing LAGG interface up/down
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

// TestGet tests getting LAGG configuration
func TestGet(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(name)

	// Set protocol
	if err := SetProto(name, ProtoFailover); err != nil {
		t.Fatalf("SetProto() failed: %v", err)
	}

	// Get configuration
	info, err := Get(name)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if info.Name != name {
		t.Errorf("Get().Name = %q, want %q", info.Name, name)
	}

	if info.Proto != ProtoFailover {
		t.Errorf("Get().Proto = %v, want %v", info.Proto, ProtoFailover)
	}
}
