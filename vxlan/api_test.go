//go:build freebsd
// +build freebsd

package vxlan

import (
	"errors"
	"net"
	"os"
	"testing"

	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Test endpoints. The local address is the loopback address so the suite never
// binds to, or sends through, a real network interface, and the ports are well
// away from the IANA VXLAN port so a running deployment is not disturbed.
const (
	testLocal      = "127.0.0.1"
	testRemote     = "127.0.0.2"
	testPeerRemote = "127.0.0.3"
	testPort       = 14789
	testVNI        = 4242
)

func skipIfNotRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}
}

func skipIfNotE2E(t *testing.T) {
	t.Helper()
	if os.Getenv("IFCLIB_E2E") != "1" {
		t.Skip("E2E tests disabled. Set IFCLIB_E2E=1 to enable")
	}
}

// newVxlan creates a VXLAN interface and registers its destruction, so nothing
// survives the test on any path.
func newVxlan(t *testing.T) string {
	t.Helper()
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := Create()
	if err != nil {
		if errors.Is(err, isyscall.ErrInvalidArgument) || errors.Is(err, isyscall.ErrNotSupported) {
			t.Skipf("vxlan(4) is unavailable, run kldload if_vxlan: %v", err)
		}
		t.Fatalf("Create() failed: %v", err)
	}
	if name == "" {
		t.Fatal("Create() returned an empty name")
	}

	t.Cleanup(func() {
		if err := Destroy(name); err != nil {
			t.Errorf("cleanup: Destroy(%s) failed: %v", name, err)
		}
	})
	return name
}

// configureTestTunnel applies a complete, valid configuration.
func configureTestTunnel(t *testing.T, name string) {
	t.Helper()

	err := Configure(name, Params{
		VNI:        testVNI,
		Local:      net.ParseIP(testLocal),
		Remote:     net.ParseIP(testRemote),
		LocalPort:  testPort,
		RemotePort: testPort,
	})
	if err != nil {
		t.Fatalf("Configure(%s) failed: %v", name, err)
	}
}

// TestCreateDestroy checks a VXLAN interface can be cloned and removed.
func TestCreateDestroy(t *testing.T) {
	name := newVxlan(t)

	if _, err := Get(name); err != nil {
		t.Errorf("Get(%s) on a fresh interface failed: %v", name, err)
	}
}

// TestFreshInterfaceHasNoVNI checks the kernel's "unconfigured" marker.
// vxlan_set_default_config() sets vxl_vni to VXLAN_VNI_MAX, which
// vxlan_check_vni() then rejects, so a fresh interface cannot be brought up.
func TestFreshInterfaceHasNoVNI(t *testing.T) {
	name := newVxlan(t)

	cfg, err := Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if cfg.VNI != VNIMax {
		t.Errorf("fresh interface VNI = %d, want %d (the kernel's not-set marker)", cfg.VNI, VNIMax)
	}
	if cfg.RemotePort != DefaultPort || cfg.LocalPort != DefaultPort {
		t.Errorf("fresh interface ports = %d/%d, want %d", cfg.LocalPort, cfg.RemotePort, DefaultPort)
	}
}

// TestConfigureGet round-trips a full configuration through the kernel.
func TestConfigureGet(t *testing.T) {
	name := newVxlan(t)

	learn := false
	err := Configure(name, Params{
		VNI:           testVNI,
		Local:         net.ParseIP(testLocal),
		Remote:        net.ParseIP(testRemote),
		LocalPort:     testPort,
		RemotePort:    testPort + 1,
		PortRangeMin:  40000,
		PortRangeMax:  42000,
		TTL:           17,
		Learn:         &learn,
		FtableTimeout: 900,
		FtableMax:     512,
	})
	if err != nil {
		t.Fatalf("Configure(%s) failed: %v", name, err)
	}

	cfg, err := Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}

	if cfg.Name != name {
		t.Errorf("Name = %q, want %q", cfg.Name, name)
	}
	if cfg.VNI != testVNI {
		t.Errorf("VNI = %d, want %d", cfg.VNI, testVNI)
	}
	if !cfg.Local.Equal(net.ParseIP(testLocal)) {
		t.Errorf("Local = %s, want %s", cfg.Local, testLocal)
	}
	if !cfg.Remote.Equal(net.ParseIP(testRemote)) {
		t.Errorf("Remote = %s, want %s", cfg.Remote, testRemote)
	}
	if cfg.LocalPort != testPort {
		t.Errorf("LocalPort = %d, want %d", cfg.LocalPort, testPort)
	}
	if cfg.RemotePort != testPort+1 {
		t.Errorf("RemotePort = %d, want %d", cfg.RemotePort, testPort+1)
	}
	if cfg.PortRangeMin != 40000 || cfg.PortRangeMax != 42000 {
		t.Errorf("port range = %d-%d, want 40000-42000", cfg.PortRangeMin, cfg.PortRangeMax)
	}
	if cfg.TTL != 17 {
		t.Errorf("TTL = %d, want 17", cfg.TTL)
	}
	if cfg.Learn {
		t.Error("Learn = true, want false")
	}
	if cfg.PeerTimeout != 900 {
		t.Errorf("PeerTimeout = %d, want 900", cfg.PeerTimeout)
	}
	if cfg.PeerMax != 512 {
		t.Errorf("PeerMax = %d, want 512", cfg.PeerMax)
	}
	if cfg.Multicast {
		t.Error("Multicast = true for a unicast remote")
	}
	if cfg.Up {
		t.Error("Up = true before Up() was called")
	}
}

// TestUpDown brings a configured tunnel up and back down.
func TestUpDown(t *testing.T) {
	name := newVxlan(t)
	configureTestTunnel(t, name)

	if err := Up(name, true); err != nil {
		t.Fatalf("Up(%s, true) failed: %v", name, err)
	}

	cfg, err := Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if !cfg.Up {
		t.Error("Up = false after Up(name, true)")
	}
	if !cfg.Running {
		t.Error("Running = false after bringing up a fully configured tunnel")
	}

	if err := Up(name, false); err != nil {
		t.Fatalf("Up(%s, false) failed: %v", name, err)
	}
	cfg, err = Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if cfg.Up {
		t.Error("Up = true after Up(name, false)")
	}
}

// TestUpWithoutVNIDoesNotRun checks how the kernel refuses an invalid
// configuration: vxlan_init() is void, so setting IFF_UP succeeds and the
// interface stays UP but never becomes RUNNING.
func TestUpWithoutVNIDoesNotRun(t *testing.T) {
	name := newVxlan(t)

	if err := Up(name, true); err != nil {
		t.Fatalf("Up(%s, true) failed: %v", name, err)
	}
	t.Cleanup(func() { _ = Up(name, false) })

	cfg, err := Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if !cfg.Up {
		t.Error("Up = false after Up(name, true)")
	}
	if cfg.Running {
		t.Error("Running = true for an interface with no VNI or addresses")
	}
}

// TestEncapOverhead checks the encapsulation overhead the kernel accounts for.
//
// vxlan_setup_interface_hdrlen() sets if_hdrlen to ETHER_HDR_LEN plus the UDP
// and VXLAN headers plus the outer IP header, and then derives if_mtu as
// ETHERMTU minus that. Over IPv4 those are 50 and 1450. Overhead reads
// if_data.ifi_hdrlen back from the live interface, so this checks the kernel's
// own number rather than arithmetic done in the library.
func TestEncapOverhead(t *testing.T) {
	name := newVxlan(t)
	configureTestTunnel(t, name)

	overhead, err := Overhead(name)
	if err != nil {
		t.Fatalf("Overhead(%s) failed: %v", name, err)
	}
	if overhead != OverheadIPv4 {
		t.Errorf("Overhead(%s) = %d, want %d", name, overhead, OverheadIPv4)
	}

	cfg, err := Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if cfg.MTU != 1500-OverheadIPv4 {
		t.Errorf("MTU = %d, want %d (ETHERMTU - overhead)", cfg.MTU, 1500-OverheadIPv4)
	}
	if cfg.MTU+overhead != 1500 {
		t.Errorf("MTU %d + overhead %d = %d, want 1500", cfg.MTU, overhead, cfg.MTU+overhead)
	}
}

// TestConfigureWhileRunningIsBusy checks that the kernel refuses configuration
// changes on a running interface, as vxlan_can_change_config() requires.
func TestConfigureWhileRunningIsBusy(t *testing.T) {
	name := newVxlan(t)
	configureTestTunnel(t, name)

	if err := Up(name, true); err != nil {
		t.Fatalf("Up(%s, true) failed: %v", name, err)
	}
	t.Cleanup(func() { _ = Up(name, false) })

	err := Configure(name, Params{VNI: testVNI + 1})
	if err == nil {
		t.Fatal("Configure() on a running interface should fail")
	}
	if !errors.Is(err, isyscall.ErrBusy) {
		t.Errorf("Configure() on a running interface: got %v, want a busy error", err)
	}

	// The VNI must be unchanged.
	cfg, err := Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if cfg.VNI != testVNI {
		t.Errorf("VNI = %d after a rejected change, want %d", cfg.VNI, testVNI)
	}
}

// TestPeerAddDelList exercises static forwarding table entries, which is how
// FreeBSD expresses VXLAN peers (VXLAN_CMD_FTABLE_ENTRY_ADD / _REM).
func TestPeerAddDelList(t *testing.T) {
	name := newVxlan(t)
	configureTestTunnel(t, name)

	mac, err := net.ParseMAC("02:11:22:33:44:55")
	if err != nil {
		t.Fatalf("ParseMAC: %v", err)
	}
	remote := net.ParseIP(testPeerRemote)

	if err := PeerAdd(name, mac, remote, 0); err != nil {
		t.Fatalf("PeerAdd(%s, %s, %s) failed: %v", name, mac, remote, err)
	}
	t.Cleanup(func() { _ = FlushPeers(name, true) })

	cfg, err := Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if cfg.PeerCount != 1 {
		t.Errorf("PeerCount = %d after one PeerAdd, want 1", cfg.PeerCount)
	}

	peers, err := Peers(name)
	if err != nil {
		t.Fatalf("Peers(%s) failed: %v", name, err)
	}
	if len(peers) != 1 {
		t.Fatalf("Peers() returned %d entries, want 1: %+v", len(peers), peers)
	}
	if peers[0].MAC.String() != mac.String() {
		t.Errorf("peer MAC = %s, want %s", peers[0].MAC, mac)
	}
	if !peers[0].Remote.Equal(remote) {
		t.Errorf("peer remote = %s, want %s", peers[0].Remote, remote)
	}
	if !peers[0].Static {
		t.Error("peer added with PeerAdd should be static")
	}

	if err := PeerDel(name, mac); err != nil {
		t.Fatalf("PeerDel(%s, %s) failed: %v", name, mac, err)
	}

	peers, err = Peers(name)
	if err != nil {
		t.Fatalf("Peers(%s) after delete failed: %v", name, err)
	}
	if len(peers) != 0 {
		t.Errorf("Peers() returned %d entries after PeerDel, want 0: %+v", len(peers), peers)
	}

	cfg, err = Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if cfg.PeerCount != 0 {
		t.Errorf("PeerCount = %d after PeerDel, want 0", cfg.PeerCount)
	}

	// Removing an entry that is not there is reported, not swallowed.
	if err := PeerDel(name, mac); err == nil {
		t.Error("PeerDel() on an absent entry should report an error")
	} else if !errors.Is(err, isyscall.ErrNotFound) {
		t.Errorf("PeerDel() on an absent entry: got %v, want a not-found error", err)
	}
}

// TestFlushPeers checks that FlushPeers drops static entries when asked.
func TestFlushPeers(t *testing.T) {
	name := newVxlan(t)
	configureTestTunnel(t, name)

	macs := []string{"02:11:22:33:44:66", "02:11:22:33:44:77"}
	for _, m := range macs {
		mac, err := net.ParseMAC(m)
		if err != nil {
			t.Fatalf("ParseMAC(%s): %v", m, err)
		}
		if err := PeerAdd(name, mac, net.ParseIP(testPeerRemote), 0); err != nil {
			t.Fatalf("PeerAdd(%s) failed: %v", m, err)
		}
	}

	cfg, err := Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if cfg.PeerCount != uint32(len(macs)) {
		t.Fatalf("PeerCount = %d, want %d", cfg.PeerCount, len(macs))
	}

	// Without includeStatic the static entries must survive.
	if err := FlushPeers(name, false); err != nil {
		t.Fatalf("FlushPeers(%s, false) failed: %v", name, err)
	}
	cfg, err = Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if cfg.PeerCount != uint32(len(macs)) {
		t.Errorf("PeerCount = %d after a dynamic-only flush, want %d", cfg.PeerCount, len(macs))
	}

	if err := FlushPeers(name, true); err != nil {
		t.Fatalf("FlushPeers(%s, true) failed: %v", name, err)
	}
	cfg, err = Get(name)
	if err != nil {
		t.Fatalf("Get(%s) failed: %v", name, err)
	}
	if cfg.PeerCount != 0 {
		t.Errorf("PeerCount = %d after a full flush, want 0", cfg.PeerCount)
	}
}

// TestPeerAddFamilyMismatch checks the kernel's EAFNOSUPPORT guard: a static
// entry must use the same address family as the interface's remote address.
func TestPeerAddFamilyMismatch(t *testing.T) {
	name := newVxlan(t)
	configureTestTunnel(t, name)

	mac, err := net.ParseMAC("02:11:22:33:44:88")
	if err != nil {
		t.Fatalf("ParseMAC: %v", err)
	}

	if err := PeerAdd(name, mac, net.ParseIP("2001:db8::1"), 0); err == nil {
		_ = PeerDel(name, mac)
		t.Error("PeerAdd() with an IPv6 peer on an IPv4 tunnel should fail")
	}
}

// TestInvalidVNI checks the VNI range guard.
func TestInvalidVNI(t *testing.T) {
	name := newVxlan(t)

	err := Configure(name, Params{VNI: VNIMax})
	if err == nil {
		t.Errorf("Configure() with VNI %d should fail", uint32(VNIMax))
	}
}

// TestInvalidMAC checks the peer MAC length guard.
func TestInvalidMAC(t *testing.T) {
	name := newVxlan(t)
	configureTestTunnel(t, name)

	if err := PeerAdd(name, net.HardwareAddr{1, 2, 3}, net.ParseIP(testPeerRemote), 0); err == nil {
		t.Error("PeerAdd() with a 3-byte MAC should fail")
	}
	if err := PeerDel(name, net.HardwareAddr{1, 2, 3}); err == nil {
		t.Error("PeerDel() with a 3-byte MAC should fail")
	}
}

// TestGetNonExistent checks that Get reports a missing interface.
func TestGetNonExistent(t *testing.T) {
	if _, err := Get("vxlan999999"); err == nil {
		t.Error("Get() should fail for a non-existent interface")
	}
}

// TestPeersNeedsKernelName documents that the forwarding table dump is keyed by
// the clone unit, so it only works under the kernel-assigned name.
func TestPeersNeedsKernelName(t *testing.T) {
	if _, err := Peers("myoverlay0"); err == nil {
		t.Error("Peers() should reject a name that is not vxlanN")
	}
}
