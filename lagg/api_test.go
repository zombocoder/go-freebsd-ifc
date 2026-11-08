//go:build freebsd
// +build freebsd

package lagg

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/zombocoder/go-freebsd-ifc/epair"
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
			continue
		}
		// ifconfig is the independent source of truth for what the
		// kernel actually applied.
		if got := ifconfigLaggProto(t, name); got != proto.String() {
			t.Errorf("after SetProto(%s), ifconfig reports laggproto %q", proto, got)
		}
		info, err := Get(name)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}
		if info.Proto != proto {
			t.Errorf("after SetProto(%s), Get().Proto = %s", proto, info.Proto)
		}
	}
}

// ifconfigLaggProto returns the protocol ifconfig(8) reports for a lagg.
func ifconfigLaggProto(t *testing.T, name string) string {
	t.Helper()
	out, err := exec.Command("ifconfig", name).CombinedOutput()
	if err != nil {
		t.Fatalf("ifconfig %s: %v: %s", name, err, out)
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "laggproto" {
			return fields[1]
		}
	}
	t.Fatalf("no laggproto in ifconfig %s output:\n%s", name, out)
	return ""
}

// TestPorts adds and removes lagg ports and checks them with Get and ifconfig.
func TestPorts(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(name)

	var ports []string
	for i := 0; i < 2; i++ {
		pair, err := epair.Create()
		if err != nil {
			t.Fatalf("epair.Create: %v", err)
		}
		defer epair.Destroy(pair.A)
		ports = append(ports, pair.A)
	}

	for _, p := range ports {
		if err := AddPort(name, p); err != nil {
			t.Fatalf("AddPort(%s, %s) failed: %v", name, p, err)
		}
	}

	info, err := Get(name)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if len(info.Ports) != len(ports) {
		t.Fatalf("Get().Ports = %v, want %v", info.Ports, ports)
	}
	for _, p := range ports {
		out, _ := exec.Command("ifconfig", name).CombinedOutput()
		if !strings.Contains(string(out), "laggport: "+p) {
			t.Errorf("ifconfig does not list %s as laggport:\n%s", p, out)
		}
	}

	if err := DelPort(name, ports[0]); err != nil {
		t.Fatalf("DelPort(%s, %s) failed: %v", name, ports[0], err)
	}
	info, err = Get(name)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if len(info.Ports) != 1 || info.Ports[0] != ports[1] {
		t.Errorf("after DelPort, Get().Ports = %v, want [%s]", info.Ports, ports[1])
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
