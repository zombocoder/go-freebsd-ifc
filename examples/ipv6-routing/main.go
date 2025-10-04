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
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "add-default":
		if len(os.Args) < 4 {
			fmt.Println("Usage: ipv6-routing add-default <interface> <gateway>")
			os.Exit(1)
		}
		addDefault(os.Args[2], os.Args[3])
	case "del-default":
		if len(os.Args) < 4 {
			fmt.Println("Usage: ipv6-routing del-default <interface> <gateway>")
			os.Exit(1)
		}
		delDefault(os.Args[2], os.Args[3])
	case "add-route":
		if len(os.Args) < 5 {
			fmt.Println("Usage: ipv6-routing add-route <destination> <gateway> <interface>")
			os.Exit(1)
		}
		addRoute(os.Args[2], os.Args[3], os.Args[4])
	case "del-route":
		if len(os.Args) < 5 {
			fmt.Println("Usage: ipv6-routing del-route <destination> <gateway> <interface>")
			os.Exit(1)
		}
		delRoute(os.Args[2], os.Args[3], os.Args[4])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("IPv6 Routing Management Demo")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ipv6-routing add-default <iface> <gw>     # Add default route (requires root)")
	fmt.Println("  ipv6-routing del-default <iface> <gw>     # Delete default route (requires root)")
	fmt.Println("  ipv6-routing add-route <dst> <gw> <iface> # Add route (requires root)")
	fmt.Println("  ipv6-routing del-route <dst> <gw> <iface> # Delete route (requires root)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Add default route via link-local gateway")
	fmt.Println("  doas go run examples/ipv6-routing/main.go add-default em0 fe80::1")
	fmt.Println()
	fmt.Println("  # Add route to specific network")
	fmt.Println("  doas go run examples/ipv6-routing/main.go add-route 2001:db8::/32 fe80::1 em0")
	fmt.Println()
	fmt.Println("  # Delete default route")
	fmt.Println("  doas go run examples/ipv6-routing/main.go del-default em0 fe80::1")
	fmt.Println()
	fmt.Println("  # Delete specific route")
	fmt.Println("  doas go run examples/ipv6-routing/main.go del-route 2001:db8::/32 fe80::1 em0")
}

func addDefault(iface, gwStr string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	gw := net.ParseIP(gwStr)
	if gw == nil {
		log.Fatalf("Invalid IPv6 address: %s", gwStr)
	}

	if gw.To4() != nil {
		log.Fatal("Error: Not an IPv6 address. Use IPv4 commands for IPv4 routing.")
	}

	fmt.Printf("Adding IPv6 default route via %s on %s...\n", gwStr, iface)

	if err := route.AddDefault6(iface, gw); err != nil {
		log.Fatalf("Failed to add default route: %v", err)
	}

	fmt.Println("✓ IPv6 default route added successfully")
}

func delDefault(iface, gwStr string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	gw := net.ParseIP(gwStr)
	if gw == nil {
		log.Fatalf("Invalid IPv6 address: %s", gwStr)
	}

	if gw.To4() != nil {
		log.Fatal("Error: Not an IPv6 address. Use IPv4 commands for IPv4 routing.")
	}

	fmt.Printf("Deleting IPv6 default route via %s on %s...\n", gwStr, iface)

	if err := route.DelDefault6(iface, gw); err != nil {
		log.Fatalf("Failed to delete default route: %v", err)
	}

	fmt.Println("✓ IPv6 default route deleted successfully")
}

func addRoute(dstStr, gwStr, iface string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	_, dst, err := net.ParseCIDR(dstStr)
	if err != nil {
		log.Fatalf("Invalid destination network: %v", err)
	}

	gw := net.ParseIP(gwStr)
	if gw == nil {
		log.Fatalf("Invalid IPv6 address: %s", gwStr)
	}

	if dst.IP.To4() != nil {
		log.Fatal("Error: Not an IPv6 network. Use IPv4 commands for IPv4 routing.")
	}

	if gw.To4() != nil {
		log.Fatal("Error: Not an IPv6 address. Use IPv4 commands for IPv4 routing.")
	}

	fmt.Printf("Adding route to %s via %s on %s...\n", dstStr, gwStr, iface)

	if err := route.AddRoute6(dst, gw, iface); err != nil {
		log.Fatalf("Failed to add route: %v", err)
	}

	fmt.Println("✓ IPv6 route added successfully")
}

func delRoute(dstStr, gwStr, iface string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	_, dst, err := net.ParseCIDR(dstStr)
	if err != nil {
		log.Fatalf("Invalid destination network: %v", err)
	}

	gw := net.ParseIP(gwStr)
	if gw == nil {
		log.Fatalf("Invalid IPv6 address: %s", gwStr)
	}

	if dst.IP.To4() != nil {
		log.Fatal("Error: Not an IPv6 network. Use IPv4 commands for IPv4 routing.")
	}

	if gw.To4() != nil {
		log.Fatal("Error: Not an IPv6 address. Use IPv4 commands for IPv4 routing.")
	}

	fmt.Printf("Deleting route to %s via %s on %s...\n", dstStr, gwStr, iface)

	if err := route.DelRoute6(dst, gw, iface); err != nil {
		log.Fatalf("Failed to delete route: %v", err)
	}

	fmt.Println("✓ IPv6 route deleted successfully")
}
