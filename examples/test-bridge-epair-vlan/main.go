//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"

	"github.com/zombocoder/go-freebsd-ifc/bridge"
	"github.com/zombocoder/go-freebsd-ifc/epair"
	"github.com/zombocoder/go-freebsd-ifc/vlan"
)

func main() {
	// Create bridge
	fmt.Println("Creating bridge...")
	br, err := bridge.Create()
	if err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}
	fmt.Printf("✓ Created bridge: %s\n", br)
	defer func() {
		fmt.Printf("Cleaning up bridge %s...\n", br)
		bridge.Destroy(br)
	}()

	// Create epair
	fmt.Println("\nCreating epair...")
	pair, err := epair.Create()
	if err != nil {
		log.Fatalf("Failed to create epair: %v", err)
	}
	fmt.Printf("✓ Created epair: %s <-> %s\n", pair.A, pair.B)
	defer func() {
		fmt.Printf("Cleaning up epair %s...\n", pair.A)
		epair.Destroy(pair.A)
	}()

	// Add epair B to bridge
	fmt.Printf("\nAdding %s to bridge %s...\n", pair.B, br)
	if err := bridge.AddMember(br, pair.B); err != nil {
		log.Fatalf("Failed to add member to bridge: %v", err)
	}
	fmt.Printf("✓ Added %s to bridge\n", pair.B)

	// List bridge members
	members, err := bridge.Members(br)
	if err != nil {
		log.Fatalf("Failed to list bridge members: %v", err)
	}
	fmt.Printf("Bridge %s members: %v\n", br, members)

	// Create VLAN
	fmt.Println("\nCreating VLAN...")
	vl, err := vlan.Create()
	if err != nil {
		log.Fatalf("Failed to create vlan: %v", err)
	}
	fmt.Printf("✓ Created VLAN interface: %s\n", vl)
	defer func() {
		fmt.Printf("Cleaning up VLAN %s...\n", vl)
		vlan.Destroy(vl)
	}()

	// Configure VLAN (tag 100 on em0)
	fmt.Printf("\nConfiguring VLAN %s with tag 100 on em0...\n", vl)
	if err := vlan.Configure(vl, 100, "em0"); err != nil {
		log.Fatalf("Failed to configure vlan: %v", err)
	}
	fmt.Printf("✓ Configured VLAN\n")

	// Get VLAN config
	cfg, err := vlan.Get(vl)
	if err != nil {
		log.Fatalf("Failed to get vlan config: %v", err)
	}
	fmt.Printf("VLAN config: tag=%d, parent=%s\n", cfg.Tag, cfg.Parent)

	// Add VLAN to bridge
	fmt.Printf("\nAdding VLAN %s to bridge %s...\n", vl, br)
	if err := bridge.AddMember(br, vl); err != nil {
		log.Fatalf("Failed to add vlan to bridge: %v", err)
	}
	fmt.Printf("✓ Added VLAN to bridge\n")

	// Bring bridge up
	fmt.Printf("\nBringing bridge %s up...\n", br)
	if err := bridge.Up(br, true); err != nil {
		log.Fatalf("Failed to bring bridge up: %v", err)
	}
	fmt.Printf("✓ Bridge is up\n")

	// List final bridge members
	members, err = bridge.Members(br)
	if err != nil {
		log.Fatalf("Failed to list bridge members: %v", err)
	}
	fmt.Printf("\nFinal bridge %s members: %v\n", br, members)

	fmt.Println("\n✓ All operations completed successfully!")
}
