//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"os"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/tap"
	"github.com/zombocoder/go-freebsd-ifc/tun"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create-tap":
		createTAP()
	case "create-tun":
		createTUN()
	case "destroy-tap":
		if len(os.Args) < 3 {
			fmt.Println("Usage: tap-tun-demo destroy-tap <name>")
			os.Exit(1)
		}
		destroyTAP(os.Args[2])
	case "destroy-tun":
		if len(os.Args) < 3 {
			fmt.Println("Usage: tap-tun-demo destroy-tun <name>")
			os.Exit(1)
		}
		destroyTUN(os.Args[2])
	case "list":
		listInterfaces()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("TAP/TUN Interface Management Demo")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tap-tun-demo create-tap           # Create TAP interface (requires root)")
	fmt.Println("  tap-tun-demo create-tun           # Create TUN interface (requires root)")
	fmt.Println("  tap-tun-demo destroy-tap <name>   # Destroy TAP interface (requires root)")
	fmt.Println("  tap-tun-demo destroy-tun <name>   # Destroy TUN interface (requires root)")
	fmt.Println("  tap-tun-demo list                 # List all tap/tun interfaces")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  doas go run examples/tap-tun-demo/main.go create-tap")
	fmt.Println("  doas go run examples/tap-tun-demo/main.go create-tun")
	fmt.Println("  go run examples/tap-tun-demo/main.go list")
	fmt.Println("  doas go run examples/tap-tun-demo/main.go destroy-tap tap0")
}

func createTAP() {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	fmt.Println("Creating TAP interface...")
	name, err := tap.Create()
	if err != nil {
		log.Fatalf("Failed to create TAP: %v", err)
	}
	fmt.Printf("✓ Created TAP interface: %s\n", name)

	fmt.Println("Bringing TAP interface up...")
	if err := tap.Up(name, true); err != nil {
		log.Printf("Warning: Failed to bring TAP up: %v", err)
	} else {
		fmt.Println("✓ TAP interface is up")
	}

	showInterface(name)
}

func createTUN() {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	fmt.Println("Creating TUN interface...")
	name, err := tun.Create()
	if err != nil {
		log.Fatalf("Failed to create TUN: %v", err)
	}
	fmt.Printf("✓ Created TUN interface: %s\n", name)

	fmt.Println("Bringing TUN interface up...")
	if err := tun.Up(name, true); err != nil {
		log.Printf("Warning: Failed to bring TUN up: %v", err)
	} else {
		fmt.Println("✓ TUN interface is up")
	}

	showInterface(name)
}

func destroyTAP(name string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	fmt.Printf("Destroying TAP interface %s...\n", name)
	if err := tap.Destroy(name); err != nil {
		log.Fatalf("Failed to destroy TAP: %v", err)
	}
	fmt.Printf("✓ TAP interface %s destroyed\n", name)
}

func destroyTUN(name string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	fmt.Printf("Destroying TUN interface %s...\n", name)
	if err := tun.Destroy(name); err != nil {
		log.Fatalf("Failed to destroy TUN: %v", err)
	}
	fmt.Printf("✓ TUN interface %s destroyed\n", name)
}

func listInterfaces() {
	ifaces, err := ifc.List()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}

	tapFound := false
	tunFound := false

	fmt.Println("TAP Interfaces:")
	fmt.Println("===============")
	for _, iface := range ifaces {
		if len(iface.Name) >= 3 && iface.Name[:3] == "tap" {
			tapFound = true
			state := "DOWN"
			if iface.Flags.IsUp() {
				state = "UP"
			}
			fmt.Printf("  %s - %s (MTU: %d)\n", iface.Name, state, iface.MTU)
		}
	}
	if !tapFound {
		fmt.Println("  (none)")
	}

	fmt.Println()
	fmt.Println("TUN Interfaces:")
	fmt.Println("===============")
	for _, iface := range ifaces {
		if len(iface.Name) >= 3 && iface.Name[:3] == "tun" {
			tunFound = true
			state := "DOWN"
			if iface.Flags.IsUp() {
				state = "UP"
			}
			fmt.Printf("  %s - %s (MTU: %d)\n", iface.Name, state, iface.MTU)
		}
	}
	if !tunFound {
		fmt.Println("  (none)")
	}
}

func showInterface(name string) {
	iface, err := ifc.Get(name)
	if err != nil {
		log.Printf("Failed to get interface details: %v", err)
		return
	}

	state := "DOWN"
	if iface.Flags.IsUp() {
		state = "UP"
	}

	fmt.Println()
	fmt.Printf("Interface Details: %s\n", name)
	fmt.Println("===================")
	fmt.Printf("  State:  %s\n", state)
	fmt.Printf("  MTU:    %d\n", iface.MTU)
	fmt.Printf("  Index:  %d\n", iface.Index)
	if len(iface.Addrs) > 0 {
		fmt.Printf("  Addresses:\n")
		for _, addr := range iface.Addrs {
			fmt.Printf("    %s\n", addr)
		}
	} else {
		fmt.Printf("  Addresses: (none)\n")
	}
}
