//go:build freebsd
// +build freebsd

package vxlan

import (
	"fmt"
	"net"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	"github.com/zombocoder/go-freebsd-ifc/internal/ifops"
	"github.com/zombocoder/go-freebsd-ifc/internal/vxlanops"
)

// Protocol constants from <net/if_vxlan.h>.
const (
	// VNIMax is one past the largest usable VNI. vxlan_check_vni() in
	// sys/net/if_vxlan.c rejects any value >= VNIMax, so the usable range
	// is 0..VNIMax-1.
	VNIMax = constants.VXLAN_VNI_MAX

	// DefaultPort is the UDP port assigned to VXLAN by IANA, and the port
	// vxlan_set_default_config() uses unless net.link.vxlan.legacy_port is
	// set.
	DefaultPort = constants.VXLAN_PORT

	// LegacyPort is the UDP port early Linux implementations used.
	LegacyPort = constants.VXLAN_LEGACY_PORT
)

// Encapsulation overhead, as the kernel accounts for it in
// vxlan_setup_interface_hdrlen() (sys/net/if_vxlan.c):
//
//	if_hdrlen = ETHER_HDR_LEN + sizeof(struct vxlanudphdr) + sizeof(struct ip)
//
// where struct vxlanudphdr is struct udphdr followed by struct vxlan_header. A
// C probe against the FreeBSD 14.3 headers gives ETHER_HDR_LEN 14,
// sizeof(struct udphdr) 8, sizeof(struct vxlan_header) 8, sizeof(struct ip) 20
// and sizeof(struct ip6_hdr) 40, hence 50 and 70 bytes. The kernel then sets
// if_mtu to ETHERMTU minus that, so a vxlan over IPv4 defaults to MTU 1450.
//
// Overhead() reads the live number back out of the kernel; prefer it over
// these constants whenever an interface actually exists.
const (
	// OverheadIPv4 is the per-frame overhead of a VXLAN tunnel whose outer
	// header is IPv4.
	OverheadIPv4 = 50

	// OverheadIPv6 is the per-frame overhead of a VXLAN tunnel whose outer
	// header is IPv6.
	OverheadIPv6 = 70
)

// Params describes the configuration to apply to a VXLAN interface.
//
// Zero-valued fields are left alone, so Params can be used to change one
// setting at a time. The kernel only accepts configuration changes while the
// interface is down (vxlan_can_change_config() returns 0 once IFF_DRV_RUNNING
// is set), so configure before calling Up.
type Params struct {
	// VNI is the virtual network identifier. Set VNISet to apply a VNI of
	// 0, which is otherwise indistinguishable from "leave alone".
	VNI    uint32
	VNISet bool

	// Local is the source address of the tunnel endpoint. It must not be a
	// multicast address.
	Local net.IP

	// Remote is the destination of the tunnel. A unicast address makes a
	// point-to-point tunnel; a multicast address makes the interface join
	// that group, which is what ifconfig calls "vxlangroup".
	Remote net.IP

	// Dev is the interface used to send multicast traffic, for use with a
	// multicast Remote (ifconfig's "vxlandev").
	Dev string

	// LocalPort and RemotePort are the UDP ports. Zero leaves the kernel
	// default (DefaultPort) in place.
	LocalPort  uint16
	RemotePort uint16

	// PortRangeMin and PortRangeMax bound the source port range used for
	// per-flow entropy. Both must be set together.
	PortRangeMin uint16
	PortRangeMax uint16

	// TTL is the TTL or hop limit of the outer header. Zero leaves the
	// kernel default.
	TTL uint8

	// Learn, when non-nil, enables or disables learning of remote MAC
	// addresses from received traffic. Learning is on by default.
	Learn *bool

	// FtableTimeout is the lifetime in seconds of a learned forwarding
	// table entry. Zero leaves the kernel default.
	FtableTimeout uint32

	// FtableMax is the maximum number of forwarding table entries. Zero
	// leaves the kernel default.
	FtableMax uint32
}

// Config is the kernel's view of a VXLAN interface.
type Config struct {
	Name string

	// VNI is the virtual network identifier. A freshly cloned interface
	// reports VNIMax, which the kernel uses to mean "not configured".
	VNI uint32

	// Local and LocalPort are the source endpoint.
	Local     net.IP
	LocalPort uint16

	// Remote and RemotePort are the destination endpoint. Remote is a
	// multicast group when Multicast is true.
	Remote     net.IP
	RemotePort uint16

	// Multicast reports whether Remote is a multicast group.
	Multicast bool

	// Dev is the interface used for multicast traffic, resolved from the
	// index the kernel reports. It is empty when no multicast interface is
	// configured.
	Dev string

	// PortRangeMin and PortRangeMax are the source port range.
	PortRangeMin uint16
	PortRangeMax uint16

	// TTL is the TTL or hop limit of the outer header.
	TTL uint8

	// Learn reports whether MAC learning is enabled.
	Learn bool

	// PeerCount is the number of forwarding table entries, static and
	// learned. It comes straight from ifvxlancfg.vxlc_ftable_cnt.
	PeerCount uint32

	// PeerMax and PeerTimeout are the forwarding table limits.
	PeerMax     uint32
	PeerTimeout uint32

	// MTU and Up come from the interface itself.
	MTU int
	Up  bool

	// Running reports whether the tunnel actually started, i.e. whether
	// the kernel set IFF_DRV_RUNNING. Up alone is not enough: vxlan_init()
	// in sys/net/if_vxlan.c returns void, so a configuration that
	// vxlan_valid_init_config() rejects leaves the interface UP but never
	// RUNNING, and the ioctl that set IFF_UP still succeeds.
	Running bool
}

// Peer is a static or learned forwarding table entry: the remote VTEP that
// frames for MAC are sent to.
type Peer struct {
	// MAC is the remote MAC address reached through Remote.
	MAC net.HardwareAddr

	// Remote is the address of the remote VTEP.
	Remote net.IP

	// Static reports whether the entry was installed with PeerAdd rather
	// than learned from received traffic.
	Static bool

	// Expire is the kernel's raw expiry stamp for the entry, in seconds
	// of system uptime. vxlan_ftable_entry_init sets it for every entry,
	// but only learned entries are ever pruned, so it is meaningful only
	// when Static is false.
	Expire int64
}

// Create creates a new VXLAN interface.
//
// The kernel assigns the name (e.g. "vxlan0"). The interface comes up
// unconfigured: vxlan_set_default_config() gives it VNI VNIMax, which is the
// kernel's "not set" marker, so Configure must supply a VNI and a local and
// remote address before Up will succeed.
//
// Requires root privileges.
func Create() (string, error) {
	name, err := cloneops.Create("vxlan")
	if err != nil {
		return "", fmt.Errorf("create vxlan: %w", err)
	}
	return name, nil
}

// Destroy destroys a VXLAN interface.
//
// Requires root privileges.
func Destroy(name string) error {
	if err := cloneops.Destroy(name); err != nil {
		return fmt.Errorf("destroy vxlan %s: %w", name, err)
	}
	return nil
}

// Configure applies p to a VXLAN interface.
//
// Each field maps to one VXLAN_CMD_SET_* driver command, issued through
// SIOCSDRVSPEC. The interface must be down: the kernel answers EBUSY for every
// one of these commands once IFF_DRV_RUNNING is set.
//
// Requires root privileges.
//
// Example:
//
//	name, _ := vxlan.Create()
//	err := vxlan.Configure(name, vxlan.Params{
//		VNI:    100,
//		Local:  net.ParseIP("192.0.2.1"),
//		Remote: net.ParseIP("192.0.2.2"),
//	})
func Configure(name string, p Params) error {
	if p.VNISet || p.VNI != 0 {
		if err := vxlanops.SetVNI(name, p.VNI); err != nil {
			return fmt.Errorf("configure vxlan %s (vni=%d): %w", name, p.VNI, err)
		}
	}
	if p.Local != nil {
		if err := vxlanops.SetLocalAddr(name, p.Local); err != nil {
			return fmt.Errorf("configure vxlan %s (local=%s): %w", name, p.Local, err)
		}
	}
	if p.Remote != nil {
		if err := vxlanops.SetRemoteAddr(name, p.Remote); err != nil {
			return fmt.Errorf("configure vxlan %s (remote=%s): %w", name, p.Remote, err)
		}
	}
	if p.Dev != "" {
		if err := vxlanops.SetMulticastIf(name, p.Dev); err != nil {
			return fmt.Errorf("configure vxlan %s (dev=%s): %w", name, p.Dev, err)
		}
	}
	if p.LocalPort != 0 {
		if err := vxlanops.SetLocalPort(name, p.LocalPort); err != nil {
			return fmt.Errorf("configure vxlan %s (local port=%d): %w", name, p.LocalPort, err)
		}
	}
	if p.RemotePort != 0 {
		if err := vxlanops.SetRemotePort(name, p.RemotePort); err != nil {
			return fmt.Errorf("configure vxlan %s (remote port=%d): %w", name, p.RemotePort, err)
		}
	}
	if p.PortRangeMin != 0 || p.PortRangeMax != 0 {
		if err := vxlanops.SetPortRange(name, p.PortRangeMin, p.PortRangeMax); err != nil {
			return fmt.Errorf("configure vxlan %s (port range=%d-%d): %w",
				name, p.PortRangeMin, p.PortRangeMax, err)
		}
	}
	if p.TTL != 0 {
		if err := vxlanops.SetTTL(name, p.TTL); err != nil {
			return fmt.Errorf("configure vxlan %s (ttl=%d): %w", name, p.TTL, err)
		}
	}
	if p.Learn != nil {
		if err := vxlanops.SetLearn(name, *p.Learn); err != nil {
			return fmt.Errorf("configure vxlan %s (learn=%v): %w", name, *p.Learn, err)
		}
	}
	if p.FtableTimeout != 0 {
		if err := vxlanops.SetFtableTimeout(name, p.FtableTimeout); err != nil {
			return fmt.Errorf("configure vxlan %s (ftable timeout=%d): %w",
				name, p.FtableTimeout, err)
		}
	}
	if p.FtableMax != 0 {
		if err := vxlanops.SetFtableMax(name, p.FtableMax); err != nil {
			return fmt.Errorf("configure vxlan %s (ftable max=%d): %w", name, p.FtableMax, err)
		}
	}
	return nil
}

// Get returns the configuration of a VXLAN interface.
//
// The tunnel settings come from VXLAN_CMD_GET_CONFIG, the only VXLAN driver
// command the kernel marks COPYOUT, so no privileges are required. MTU and Up
// come from the interface itself.
func Get(name string) (Config, error) {
	cfg, err := vxlanops.GetConfig(name)
	if err != nil {
		return Config{}, fmt.Errorf("get vxlan %s config: %w", name, err)
	}

	iface, err := ifc.Get(name)
	if err != nil {
		return Config{}, fmt.Errorf("get vxlan %s interface: %w", name, err)
	}

	out := Config{
		Name:         name,
		VNI:          cfg.VNI,
		Local:        cfg.Local,
		LocalPort:    cfg.LocalPort,
		Remote:       cfg.Remote,
		RemotePort:   cfg.RemotePort,
		Multicast:    cfg.Remote != nil && cfg.Remote.IsMulticast(),
		PortRangeMin: cfg.PortMin,
		PortRangeMax: cfg.PortMax,
		TTL:          cfg.TTL,
		Learn:        cfg.Learn,
		PeerCount:    cfg.FtableCount,
		PeerMax:      cfg.FtableMax,
		PeerTimeout:  cfg.FtableTimeout,
		MTU:          iface.MTU,
		Up:           iface.Flags.IsUp(),
		Running:      iface.Flags.IsRunning(),
	}

	if cfg.MulticastIndex > 0 {
		if dev, err := net.InterfaceByIndex(cfg.MulticastIndex); err == nil {
			out.Dev = dev.Name
		}
	}

	return out, nil
}

// Up brings the VXLAN interface up or down.
//
// Bringing the interface up is what commits the configuration: the kernel
// validates it in vxlan_valid_init_config() and only then starts the tunnel.
//
// Note that a rejected configuration is not reported here. vxlan_init() in
// sys/net/if_vxlan.c returns void and logs the reason to the console, so
// setting IFF_UP succeeds even when the VNI or the local and remote addresses
// are missing or unusable; the interface simply never becomes RUNNING. Check
// Get().Running to confirm the tunnel actually started.
//
// Requires root privileges.
func Up(name string, up bool) error {
	if err := ifops.SetFlags(name, uint32(ifc.FlagUp), up); err != nil {
		if up {
			return fmt.Errorf("bring vxlan %s up: %w", name, err)
		}
		return fmt.Errorf("bring vxlan %s down: %w", name, err)
	}
	return nil
}

// PeerAdd installs a static forwarding table entry, so frames addressed to mac
// are encapsulated towards the VTEP at remote.
//
// This is how FreeBSD expresses a VXLAN "peer": there is no peer list in the
// driver, only the per-interface forwarding table that maps a remote MAC
// address to the remote VTEP that owns it. The kernel command is
// VXLAN_CMD_FTABLE_ENTRY_ADD (sys/net/if_vxlan.c,
// vxlan_ctrl_ftable_entry_add), which requires:
//
//   - remote to be a unicast address in the same family as the interface's
//     configured remote address, otherwise EAFNOSUPPORT;
//   - remote not to be the unspecified address or a multicast group.
//
// port may be 0, in which case the kernel uses the interface's remote port.
//
// Requires root privileges.
func PeerAdd(name string, mac net.HardwareAddr, remote net.IP, port uint16) error {
	if err := vxlanops.FtableAdd(name, mac, remote, port); err != nil {
		return fmt.Errorf("add vxlan %s peer %s via %s: %w", name, mac, remote, err)
	}
	return nil
}

// PeerDel removes the static forwarding table entry for mac.
//
// The kernel looks entries up by MAC address alone
// (vxlan_ctrl_ftable_entry_rem) and answers ENOENT when there is none, which
// PeerDel reports as an error rather than swallowing: unlike an address or a
// route, removing an entry that is not there is usually a sign the caller's
// idea of the table is stale. Use errors.Is(err, ...) against the not-found
// error if idempotent removal is wanted.
//
// Requires root privileges.
func PeerDel(name string, mac net.HardwareAddr) error {
	if err := vxlanops.FtableDel(name, mac); err != nil {
		return fmt.Errorf("delete vxlan %s peer %s: %w", name, mac, err)
	}
	return nil
}

// Peers returns the forwarding table of a VXLAN interface, static entries and
// learned ones alike.
//
// The driver exposes no ioctl for reading the table back; the only kernel
// interface is the per-unit sysctl net.link.vxlan.<unit>.ftable.dump. That has
// two consequences, both imposed by the kernel:
//
//   - the sysctl node is keyed by the clone unit, so Peers only works while
//     the interface still carries its kernel-assigned "vxlanN" name;
//   - the dump is capped at one page. sys/net/if_vxlan.c says as much: "This
//     is mostly intended for debugging during development. It is not practical
//     to dump an entire large table this way."
//
// Get().PeerCount is an exact count with neither limitation.
func Peers(name string) ([]Peer, error) {
	entries, err := vxlanops.FtableList(name)
	if err != nil {
		return nil, fmt.Errorf("list vxlan %s peers: %w", name, err)
	}

	peers := make([]Peer, len(entries))
	for i, e := range entries {
		peers[i] = Peer{
			MAC:    e.MAC,
			Remote: e.Remote,
			Static: e.Static,
			Expire: e.Expire,
		}
	}
	return peers, nil
}

// FlushPeers drops forwarding table entries. When includeStatic is false only
// learned entries go; when true, entries added with PeerAdd go too.
//
// Requires root privileges.
func FlushPeers(name string, includeStatic bool) error {
	if err := vxlanops.Flush(name, includeStatic); err != nil {
		return fmt.Errorf("flush vxlan %s peers: %w", name, err)
	}
	return nil
}

// Overhead returns the number of bytes of encapsulation the kernel accounts
// for on every frame leaving this interface.
//
// The number is if_data.ifi_hdrlen, which for a vxlan interface is exactly
// what vxlan_setup_interface_hdrlen() computed: ETHER_HDR_LEN plus the UDP and
// VXLAN headers plus the outer IP header, i.e. OverheadIPv4 (50) for an IPv4
// remote and OverheadIPv6 (70) for an IPv6 one. The kernel then derives the
// default MTU as ETHERMTU minus this value, giving 1450 over IPv4.
//
// It is read from the live interface rather than computed here, so it stays
// correct if the kernel's accounting ever changes.
func Overhead(name string) (int, error) {
	n, err := vxlanops.HeaderLen(name)
	if err != nil {
		return 0, fmt.Errorf("get vxlan %s header length: %w", name, err)
	}
	return n, nil
}
