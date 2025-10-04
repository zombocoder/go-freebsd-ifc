//go:build freebsd
// +build freebsd

/*
Package ifc provides FreeBSD network interface management.

This package allows listing, querying, and configuring network interfaces
using FreeBSD's native ioctl system calls.

# Basic Usage

List all network interfaces:

	ifaces, err := ifc.List()
	if err != nil {
		log.Fatal(err)
	}
	for _, iface := range ifaces {
		fmt.Printf("%s: %v\n", iface.Name, iface.Addrs)
	}

Get a specific interface:

	iface, err := ifc.Get("em0")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("MTU: %d, Up: %v\n", iface.MTU, iface.Flags.IsUp())

Configure an interface:

	// Bring interface up
	if err := ifc.SetUp("em0", true); err != nil {
		log.Fatal(err)
	}

	// Set MTU
	if err := ifc.SetMTU("em0", 9000); err != nil {
		log.Fatal(err)
	}

	// Rename interface
	if err := ifc.Rename("em0", "wan0"); err != nil {
		log.Fatal(err)
	}

# Permissions

Read operations (List, Get) work without special privileges.
Mutation operations (SetUp, SetMTU, Rename) require root privileges.

# Error Handling

The package returns typed errors for common conditions:
  - ErrPermission: Operation requires root privileges
  - ErrNotFound: Interface not found
  - ErrExists: Interface already exists
  - ErrSyscall: Generic syscall error (wrapped with details)
*/
package ifc
