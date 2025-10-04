//go:build freebsd
// +build freebsd

/*
Package tun provides FreeBSD TUN (Layer 3 virtual network) interface management.

TUN interfaces are virtual point-to-point network devices that operate at Layer 3 (network layer).
They handle IP packets directly without Ethernet framing, making them ideal for VPNs and tunnels.

# Basic Usage

Create and configure a TUN interface:

	// Create TUN interface
	name, err := tun.Create()
	if err != nil {
		log.Fatal(err)
	}
	defer tun.Destroy(name)

	// Bring interface up
	if err := tun.Up(name, true); err != nil {
		log.Fatal(err)
	}

# Permissions

Creating and destroying TUN interfaces requires root privileges.

# TAP vs TUN

  - TAP (Layer 2): Handles Ethernet frames, works with bridges, ARP, etc.
  - TUN (Layer 3): Handles IP packets only, no Ethernet overhead

# Common Use Cases

  - VPN tunnels (OpenVPN in tun mode, WireGuard)
  - IP tunneling protocols (GRE, IPsec)
  - Network routing and forwarding
  - Point-to-point links

# FreeBSD Device Notes

TUN interfaces are created via /dev/tun device cloning. The kernel automatically
assigns interface names (tun0, tun1, etc.). Each TUN interface has a corresponding
/dev/tunN character device for userspace I/O operations.
*/
package tun
