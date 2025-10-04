//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"strings"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/vlan"
)

func main() {
	// List all interfaces
	ifaces, err := ifc.List()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}

	// Filter VLAN interfaces and get their configuration
	var vlans []struct {
		Name   string
		Tag    uint16
		Parent string
		Up     bool
		MTU    int
	}

	for _, iface := range ifaces {
		// VLAN interfaces typically start with "vlan"
		if strings.HasPrefix(iface.Name, "vlan") {
			// Get VLAN configuration
			cfg, err := vlan.Get(iface.Name)
			if err != nil {
				// Skip if not a VLAN or error getting config
				continue
			}

			vlans = append(vlans, struct {
				Name   string
				Tag    uint16
				Parent string
				Up     bool
				MTU    int
			}{
				Name:   iface.Name,
				Tag:    cfg.Tag,
				Parent: cfg.Parent,
				Up:     iface.Flags.IsUp(),
				MTU:    iface.MTU,
			})
		}
	}

	// Print results
	if len(vlans) == 0 {
		fmt.Println("No VLAN interfaces found")
		return
	}

	fmt.Printf("Found %d VLAN interface(s):\n\n", len(vlans))

	// Print header
	fmt.Printf("%-10s %-6s %-10s %-6s %-6s\n", "INTERFACE", "TAG", "PARENT", "STATE", "MTU")
	fmt.Println(strings.Repeat("-", 50))

	// Print each VLAN
	for _, v := range vlans {
		state := "DOWN"
		if v.Up {
			state = "UP"
		}

		fmt.Printf("%-10s %-6d %-10s %-6s %-6d\n",
			v.Name,
			v.Tag,
			v.Parent,
			state,
			v.MTU,
		)
	}

	fmt.Println()
}
