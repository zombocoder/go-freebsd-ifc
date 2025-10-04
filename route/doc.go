//go:build freebsd
// +build freebsd

/*
Package route provides FreeBSD routing table management.

This package allows adding and removing IPv4 and IPv6 routes, including default routes.

# Basic Usage

IPv4 default route:

	// Add default route via gateway
	gw := net.ParseIP("192.168.1.1")
	if err := route.AddDefault4("em0", gw); err != nil {
		log.Fatal(err)
	}

	// Remove default route
	if err := route.DelDefault4("em0", gw); err != nil {
		log.Fatal(err)
	}

IPv6 default route:

	// Add IPv6 default route
	gw := net.ParseIP("fe80::1")
	if err := route.AddDefault6("em0", gw); err != nil {
		log.Fatal(err)
	}

	// Remove IPv6 default route
	if err := route.DelDefault6("em0", gw); err != nil {
		log.Fatal(err)
	}

Specific IPv4 routes:

	// Add route to 10.0.0.0/24
	_, dst, _ := net.ParseCIDR("10.0.0.0/24")
	gw := net.ParseIP("192.168.1.254")
	if err := route.AddRoute4(dst, gw, "em0"); err != nil {
		log.Fatal(err)
	}

	// Remove route
	if err := route.DelRoute4(dst, gw, "em0"); err != nil {
		log.Fatal(err)
	}

Specific IPv6 routes:

	// Add route to 2001:db8::/32
	_, dst, _ := net.ParseCIDR("2001:db8::/32")
	gw := net.ParseIP("fe80::1")
	if err := route.AddRoute6(dst, gw, "em0"); err != nil {
		log.Fatal(err)
	}

	// Remove route
	if err := route.DelRoute6(dst, gw, "em0"); err != nil {
		log.Fatal(err)
	}

# Permissions

All operations require root privileges.

# Idempotency

Operations are idempotent:
  - Add: Returns nil if route already exists
  - Del: Returns nil if route doesn't exist

# Safety Warning

Modifying routes can break network connectivity. Test in isolated environments
(jails, VMs) before using in production.
*/
package route
