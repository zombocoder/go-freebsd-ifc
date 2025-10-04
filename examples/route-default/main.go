//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/zombocoder/go-freebsd-ifc/route"
)

func main() {
	if os.Geteuid() != 0 {
		log.Fatal("This program must be run as root")
	}

	// WARNING: This modifies your routing table!
	// Use with caution, preferably in a jail or test environment

	iface := "em0"                   // Change to your interface
	gw := net.ParseIP("192.168.1.1") // Change to your gateway

	fmt.Printf("WARNING: This will add a default route via %s on %s\n", gw, iface)
	fmt.Println("Press Enter to continue or Ctrl+C to abort...")
	fmt.Scanln()

	// Add default route
	fmt.Printf("\nAdding default route via %s on %s...\n", gw, iface)
	if err := route.AddDefault4(iface, gw); err != nil {
		log.Fatalf("Failed to add default route: %v", err)
	}
	fmt.Println("Default route added successfully")

	fmt.Println("\nVerify with: netstat -rn | grep default")
	fmt.Println("Press Enter to remove route and exit...")
	fmt.Scanln()

	// Delete default route
	fmt.Printf("\nRemoving default route via %s...\n", gw)
	if err := route.DelDefault4(iface, gw); err != nil {
		log.Fatalf("Failed to delete default route: %v", err)
	}
	fmt.Println("Default route removed successfully")

	fmt.Println("\nDone!")
}
