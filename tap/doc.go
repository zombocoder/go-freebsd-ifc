//go:build freebsd
// +build freebsd

/*
Package tap provides FreeBSD TAP (Layer 2 virtual network) interface management.

TAP interfaces are virtual Ethernet devices that operate at Layer 2 (data link layer).
They are commonly used for VPNs, network virtualization, and bhyve/jail networking.

# Basic Usage

Create and configure a TAP interface:

	// Create TAP interface
	name, err := tap.Create()
	if err != nil {
		log.Fatal(err)
	}
	defer tap.Destroy(name)

	// Bring interface up
	if err := tap.Up(name, true); err != nil {
		log.Fatal(err)
	}

# Permissions

Creating and destroying TAP interfaces requires root privileges.

# Common Use Cases

  - VPN tunnels (OpenVPN, WireGuard)
  - Virtual machine networking (bhyve)
  - Container/jail networking
  - Network simulation and testing

# FreeBSD Device Notes

TAP interfaces are created via /dev/tap device cloning. The kernel automatically
assigns interface names (tap0, tap1, etc.). Each TAP interface has a corresponding
/dev/tapN character device for userspace I/O operations.
*/
package tap
