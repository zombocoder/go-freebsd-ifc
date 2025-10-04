//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/zombocoder/go-freebsd-ifc/ip"
)

func main() {
	if os.Geteuid() != 0 {
		log.Fatal("This program must be run as root")
	}

	iface := "lo0"

	// Add IPv4 address
	ipv4 := net.ParseIP("127.0.1.1")
	mask := net.CIDRMask(8, 32)

	fmt.Printf("Adding IPv4 address %s/%d to %s...\n", ipv4, 8, iface)
	if err := ip.Add4(iface, ipv4, mask); err != nil {
		log.Fatalf("Failed to add IPv4 address: %v", err)
	}
	fmt.Println("IPv4 address added successfully")

	// Add IPv6 address
	ipv6 := net.ParseIP("::1:1")
	prefixLen := 128

	fmt.Printf("\nAdding IPv6 address %s/%d to %s...\n", ipv6, prefixLen, iface)
	if err := ip.Add6(iface, ipv6, prefixLen); err != nil {
		log.Fatalf("Failed to add IPv6 address: %v", err)
	}
	fmt.Println("IPv6 address added successfully")

	fmt.Println("\nPress Enter to remove addresses and exit...")
	fmt.Scanln()

	// Delete IPv4 address
	fmt.Printf("\nRemoving IPv4 address %s from %s...\n", ipv4, iface)
	if err := ip.Del4(iface, ipv4, mask); err != nil {
		log.Fatalf("Failed to delete IPv4 address: %v", err)
	}
	fmt.Println("IPv4 address removed successfully")

	// Delete IPv6 address
	fmt.Printf("\nRemoving IPv6 address %s from %s...\n", ipv6, iface)
	if err := ip.Del6(iface, ipv6, prefixLen); err != nil {
		log.Fatalf("Failed to delete IPv6 address: %v", err)
	}
	fmt.Println("IPv6 address removed successfully")

	fmt.Println("\nDone!")
}
