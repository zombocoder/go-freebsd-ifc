/*
Package vxlan provides FreeBSD vxlan(4) overlay interface management.

A VXLAN interface tunnels Ethernet frames inside UDP, tagging each with a
24-bit virtual network identifier (VNI) so that many Layer 2 segments can share
one underlay network. FreeBSD implements it in sys/net/if_vxlan.c.

# Basic Usage

	// Create the interface; the kernel picks the name.
	name, err := vxlan.Create()
	if err != nil {
		log.Fatal(err)
	}
	defer vxlan.Destroy(name)

	// Configure it while it is still down.
	err = vxlan.Configure(name, vxlan.Params{
		VNI:    100,
		Local:  net.ParseIP("192.0.2.1"),
		Remote: net.ParseIP("192.0.2.2"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Commit the configuration.
	if err := vxlan.Up(name, true); err != nil {
		log.Fatal(err)
	}

	cfg, err := vxlan.Get(name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("VNI %d, %s -> %s, MTU %d\n", cfg.VNI, cfg.Local, cfg.Remote, cfg.MTU)

# Configuration Ordering

The kernel accepts configuration changes only while the interface is down:
vxlan_can_change_config() returns 0 once IFF_DRV_RUNNING is set, and every
VXLAN_CMD_SET_* command then answers EBUSY. Configure first, then Up.

Bringing the interface up is what validates the configuration. A VNI and a
usable local and remote address are required. A freshly cloned interface
reports VNI VNIMax, which is the kernel's marker for "not configured".

A rejected configuration is not reported through the ioctl. vxlan_init() in
sys/net/if_vxlan.c returns void: it logs the reason to the console and leaves
the interface UP but not RUNNING, and setting IFF_UP still succeeds. Check
Get().Running to confirm the tunnel actually started.

# Peers

FreeBSD has no VXLAN peer list. What it has is a per-interface forwarding
table mapping a remote MAC address to the VTEP that owns it, the same structure
that MAC learning populates. A static entry is added with PeerAdd and removed
with PeerDel, which issue VXLAN_CMD_FTABLE_ENTRY_ADD and
VXLAN_CMD_FTABLE_ENTRY_REM.

Note that ifconfig(8) exposes no command for this: sbin/ifconfig/ifvxlan.c
implements vxlanflush and vxlanflushall but nothing that adds an entry, so
these two commands are reachable only through the ioctl.

A static entry needs the interface's remote address to be configured first: the
kernel rejects an entry whose address family differs from the interface's
remote address with EAFNOSUPPORT.

Peers reads the table back, but only through the kernel's debugging sysctl; see
its documentation for the limits that imposes. Get().PeerCount is an exact
count without them.

# MTU

VXLAN adds an outer Ethernet, IP, UDP and VXLAN header to every frame. The
kernel computes that in vxlan_setup_interface_hdrlen() and stores it in
if_hdrlen: 50 bytes over IPv4 and 70 over IPv6. Unless the MTU has been set
explicitly the kernel then derives if_mtu as ETHERMTU minus that, so a VXLAN
over IPv4 defaults to an MTU of 1450.

Overhead returns that number read back from the live interface, and the
OverheadIPv4 and OverheadIPv6 constants record the values the kernel's
accounting produces.

# Permissions

Create, Destroy, Configure, Up, PeerAdd, PeerDel and FlushPeers require root.
Get, Peers and Overhead do not: VXLAN_CMD_GET_CONFIG is the one driver command
the kernel does not gate behind PRIV_NET_VXLAN, and the forwarding table dump
is a read-only sysctl.
*/
package vxlan
