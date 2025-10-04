//go:build freebsd
// +build freebsd

/*
Package vlan provides FreeBSD vlan(4) VLAN interface management.

VLANs allow you to create multiple virtual networks on a single physical interface
using 802.1Q tagging.

# Basic Usage

	// Create VLAN interface
	vlan, err := vlan.Create()
	if err != nil {
		log.Fatal(err)
	}
	defer vlan.Destroy(vlan)

	// Configure: tag 100 on parent interface em0
	if err := vlan.Configure(vlan, 100, "em0"); err != nil {
		log.Fatal(err)
	}

	// Bring it up
	if err := vlan.Up(vlan, true); err != nil {
		log.Fatal(err)
	}

	// Get configuration
	cfg, err := vlan.Get(vlan)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("VLAN %d on %s\n", cfg.Tag, cfg.Parent)

# Permissions

All operations require root privileges.
*/
package vlan
