//go:build freebsd
// +build freebsd

/*
Package ip provides FreeBSD IP address management for network interfaces.

This package allows adding and removing IPv4 and IPv6 addresses from interfaces.

# Basic Usage

IPv4 addresses:

	// Add IPv4 address
	ip4 := net.ParseIP("192.168.1.10")
	mask := net.CIDRMask(24, 32)
	if err := ip.Add4("em0", ip4, mask); err != nil {
		log.Fatal(err)
	}

	// Remove IPv4 address
	if err := ip.Del4("em0", ip4, mask); err != nil {
		log.Fatal(err)
	}

IPv6 addresses:

	// Add IPv6 address
	ip6 := net.ParseIP("fd00::1")
	if err := ip.Add6("em0", ip6, 64); err != nil {
		log.Fatal(err)
	}

	// Remove IPv6 address
	if err := ip.Del6("em0", ip6, 64); err != nil {
		log.Fatal(err)
	}

# Permissions

All operations require root privileges.

# Idempotency

Operations are idempotent:
  - Add: Returns nil if address already exists
  - Del: Returns nil if address doesn't exist
*/
package ip
