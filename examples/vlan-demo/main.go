//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"os"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/vlan"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		listVLANs()
	case "create":
		if len(os.Args) < 4 {
			fmt.Println("Usage: vlan-demo create <tag> <parent-interface>")
			os.Exit(1)
		}
		var tag uint16
		fmt.Sscanf(os.Args[2], "%d", &tag)
		parent := os.Args[3]
		createVLAN(tag, parent)
	case "destroy":
		if len(os.Args) < 3 {
			fmt.Println("Usage: vlan-demo destroy <vlan-name>")
			os.Exit(1)
		}
		destroyVLAN(os.Args[2])
	case "show":
		if len(os.Args) < 3 {
			fmt.Println("Usage: vlan-demo show <vlan-name>")
			os.Exit(1)
		}
		showVLAN(os.Args[2])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("VLAN Management Demo")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  vlan-demo list                        # List all VLANs")
	fmt.Println("  vlan-demo create <tag> <parent>       # Create VLAN (requires root)")
	fmt.Println("  vlan-demo destroy <name>              # Destroy VLAN (requires root)")
	fmt.Println("  vlan-demo show <name>                 # Show VLAN details")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  vlan-demo list")
	fmt.Println("  doas vlan-demo create 100 em0")
	fmt.Println("  vlan-demo show vlan0")
	fmt.Println("  doas vlan-demo destroy vlan0")
}

func listVLANs() {
	// List all interfaces
	ifaces, err := ifc.List()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}

	fmt.Println("VLAN Interfaces:")
	fmt.Println("================")
	fmt.Println()

	found := false
	for _, iface := range ifaces {
		// Try to get VLAN config - if it succeeds, it's a VLAN
		cfg, err := vlan.Get(iface.Name)
		if err != nil {
			continue
		}

		found = true
		state := "DOWN"
		if iface.Flags.IsUp() {
			state = "UP"
		}

		fmt.Printf("%s:\n", iface.Name)
		fmt.Printf("  VLAN Tag: %d\n", cfg.Tag)
		fmt.Printf("  Parent:   %s\n", cfg.Parent)
		fmt.Printf("  State:    %s\n", state)
		fmt.Printf("  MTU:      %d\n", iface.MTU)
		fmt.Println()
	}

	if !found {
		fmt.Println("No VLAN interfaces found")
	}
}

func createVLAN(tag uint16, parent string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	// Create VLAN interface
	fmt.Printf("Creating VLAN with tag %d on %s...\n", tag, parent)

	vl, err := vlan.Create()
	if err != nil {
		log.Fatalf("Failed to create VLAN: %v", err)
	}
	fmt.Printf("✓ Created VLAN interface: %s\n", vl)

	// Configure VLAN
	fmt.Printf("Configuring VLAN tag %d on parent %s...\n", tag, parent)
	if err := vlan.Configure(vl, tag, parent); err != nil {
		vlan.Destroy(vl)
		log.Fatalf("Failed to configure VLAN: %v", err)
	}
	fmt.Printf("✓ Configured VLAN\n")

	// Bring VLAN up
	fmt.Println("Bringing VLAN up...")
	if err := vlan.Up(vl, true); err != nil {
		log.Printf("Warning: Failed to bring VLAN up: %v", err)
	} else {
		fmt.Printf("✓ VLAN is up\n")
	}

	fmt.Printf("\nVLAN %s created successfully!\n", vl)
	fmt.Println()
	showVLAN(vl)
}

func destroyVLAN(name string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	fmt.Printf("Destroying VLAN %s...\n", name)

	if err := vlan.Destroy(name); err != nil {
		log.Fatalf("Failed to destroy VLAN: %v", err)
	}

	fmt.Printf("✓ VLAN %s destroyed\n", name)
}

func showVLAN(name string) {
	// Get interface details
	iface, err := ifc.Get(name)
	if err != nil {
		log.Fatalf("Failed to get interface %s: %v", name, err)
	}

	// Get VLAN configuration
	cfg, err := vlan.Get(name)
	if err != nil {
		log.Fatalf("Failed to get VLAN config for %s: %v", name, err)
	}

	state := "DOWN"
	if iface.Flags.IsUp() {
		state = "UP"
	}

	fmt.Printf("VLAN Interface: %s\n", name)
	fmt.Println("===================")
	fmt.Printf("  VLAN Tag:   %d\n", cfg.Tag)
	fmt.Printf("  Parent:     %s\n", cfg.Parent)
	fmt.Printf("  State:      %s\n", state)
	fmt.Printf("  MTU:        %d\n", iface.MTU)
	fmt.Printf("  Index:      %d\n", iface.Index)

	if len(iface.Addrs) > 0 {
		fmt.Printf("  Addresses:\n")
		for _, addr := range iface.Addrs {
			fmt.Printf("    %s\n", addr)
		}
	} else {
		fmt.Printf("  Addresses:  (none)\n")
	}
}
