//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
)

func main() {
	ifaces, err := ifc.List()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}

	fmt.Printf("Found %d interfaces:\n\n", len(ifaces))

	for _, iface := range ifaces {
		fmt.Printf("Interface: %s\n", iface.Name)
		fmt.Printf("  Index: %d\n", iface.Index)
		fmt.Printf("  MTU: %d\n", iface.MTU)
		fmt.Printf("  Flags: 0x%x", iface.Flags)
		if iface.Flags.IsUp() {
			fmt.Print(" UP")
		}
		if iface.Flags.IsRunning() {
			fmt.Print(" RUNNING")
		}
		if iface.Flags.IsLoopback() {
			fmt.Print(" LOOPBACK")
		}
		fmt.Println()

		if len(iface.Addrs) > 0 {
			fmt.Printf("  Addresses:\n")
			for _, addr := range iface.Addrs {
				fmt.Printf("    %s\n", addr.String())
			}
		}
		fmt.Println()
	}
}
