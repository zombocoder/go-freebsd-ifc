//go:build freebsd
// +build freebsd

/*
Package epair provides FreeBSD epair(4) paired virtual interface management.

An epair creates a pair of connected virtual Ethernet interfaces. Packets
sent on one side appear on the other side, making them ideal for connecting
jails, VMs, or network namespaces.

# Basic Usage

	// Create an epair
	pair, err := epair.Create()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created: %s <-> %s\n", pair.A, pair.B)
	defer epair.Destroy(pair.A)

	// Typically, one side goes into a jail/VM, the other stays on host
	// For example, add pair.B to a bridge:
	//   bridge.AddMember("bridge0", pair.B)

# Permissions

All operations require root privileges.

# Notes

- Destroying either side destroys both interfaces
- The "A" side is typically used on the host
- The "B" side is typically moved into a jail or VM
*/
package epair
