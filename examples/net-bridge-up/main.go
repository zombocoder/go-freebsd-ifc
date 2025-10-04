//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zombocoder/go-freebsd-ifc/bridge"
	"github.com/zombocoder/go-freebsd-ifc/epair"
)

func main() {
	if os.Geteuid() != 0 {
		log.Fatal("This program must be run as root")
	}

	// Create bridge
	fmt.Println("Creating bridge...")
	br, err := bridge.Create()
	if err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}
	fmt.Printf("Created bridge: %s\n", br)
	defer func() {
		fmt.Printf("Cleaning up: destroying bridge %s\n", br)
		bridge.Destroy(br)
	}()

	// Create epair
	fmt.Println("\nCreating epair...")
	pair, err := epair.Create()
	if err != nil {
		log.Fatalf("Failed to create epair: %v", err)
	}
	fmt.Printf("Created epair: %s <-> %s\n", pair.A, pair.B)
	defer func() {
		fmt.Printf("Cleaning up: destroying epair %s\n", pair.A)
		epair.Destroy(pair.A)
	}()

	// Add epair B side to bridge
	fmt.Printf("\nAdding %s to bridge %s...\n", pair.B, br)
	if err := bridge.AddMember(br, pair.B); err != nil {
		log.Fatalf("Failed to add member to bridge: %v", err)
	}

	// Bring bridge up
	fmt.Printf("\nBringing bridge %s up...\n", br)
	if err := bridge.Up(br, true); err != nil {
		log.Fatalf("Failed to bring bridge up: %v", err)
	}

	// Get bridge info
	fmt.Println("\nBridge information:")
	info, err := bridge.Get(br)
	if err != nil {
		log.Fatalf("Failed to get bridge info: %v", err)
	}

	fmt.Printf("  Name: %s\n", info.Name)
	fmt.Printf("  Up: %v\n", info.Up)
	fmt.Printf("  MTU: %d\n", info.MTU)
	fmt.Printf("  Members: %v\n", info.Members)

	fmt.Println("\nSuccess! Bridge setup complete.")
}
