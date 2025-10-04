//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/zombocoder/go-freebsd-ifc/bridge"
	"github.com/zombocoder/go-freebsd-ifc/epair"
	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/ip"
	"github.com/zombocoder/go-freebsd-ifc/lagg"
	"github.com/zombocoder/go-freebsd-ifc/route"
	"github.com/zombocoder/go-freebsd-ifc/tap"
	"github.com/zombocoder/go-freebsd-ifc/tun"
	"github.com/zombocoder/go-freebsd-ifc/vlan"
)

func main() {
	if os.Geteuid() != 0 {
		log.Fatal("This demo requires root privileges. Run with: doas go run examples/comprehensive-demo/main.go")
	}

	fmt.Println("==========================================================")
	fmt.Println("  FreeBSD Network Interface Library - Comprehensive Demo")
	fmt.Println("==========================================================")
	fmt.Println()

	// Demo 1: TAP Interface
	fmt.Println("--- Demo 1: TAP Interface (Layer 2) ---")
	tapName, err := tap.Create()
	if err != nil {
		log.Fatalf("Failed to create TAP: %v", err)
	}
	fmt.Printf("✓ Created TAP interface: %s\n", tapName)
	defer tap.Destroy(tapName)

	if err := tap.Up(tapName, true); err != nil {
		log.Printf("Warning: Failed to bring TAP up: %v", err)
	} else {
		fmt.Printf("✓ TAP interface is UP\n")
	}
	fmt.Println()

	// Demo 2: TUN Interface
	fmt.Println("--- Demo 2: TUN Interface (Layer 3) ---")
	tunName, err := tun.Create()
	if err != nil {
		log.Fatalf("Failed to create TUN: %v", err)
	}
	fmt.Printf("✓ Created TUN interface: %s\n", tunName)
	defer tun.Destroy(tunName)

	if err := tun.Up(tunName, true); err != nil {
		log.Printf("Warning: Failed to bring TUN up: %v", err)
	} else {
		fmt.Printf("✓ TUN interface is UP\n")
	}
	fmt.Println()

	// Demo 3: VLAN Interface
	fmt.Println("--- Demo 3: VLAN Interface (802.1Q) ---")
	vlanName, err := vlan.Create()
	if err != nil {
		log.Fatalf("Failed to create VLAN: %v", err)
	}
	fmt.Printf("✓ Created VLAN interface: %s\n", vlanName)
	defer vlan.Destroy(vlanName)

	// Configure VLAN on loopback for demo purposes
	if err := vlan.Configure(vlanName, 100, "lo0"); err != nil {
		log.Printf("Warning: Failed to configure VLAN: %v", err)
	} else {
		fmt.Printf("✓ Configured VLAN: tag=100, parent=lo0\n")
	}

	vlanInfo, _ := vlan.Get(vlanName)
	fmt.Printf("✓ VLAN Info: tag=%d, parent=%s\n", vlanInfo.Tag, vlanInfo.Parent)
	fmt.Println()

	// Demo 4: Bridge with Epair
	fmt.Println("--- Demo 4: Bridge with Epair ---")
	brName, err := bridge.Create()
	if err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}
	fmt.Printf("✓ Created bridge: %s\n", brName)
	defer bridge.Destroy(brName)

	pair, err := epair.Create()
	if err != nil {
		log.Fatalf("Failed to create epair: %v", err)
	}
	fmt.Printf("✓ Created epair: %s <-> %s\n", pair.A, pair.B)
	defer epair.Destroy(pair.A)

	if err := bridge.AddMember(brName, pair.B); err != nil {
		log.Printf("Warning: Failed to add member: %v", err)
	} else {
		fmt.Printf("✓ Added %s to bridge %s\n", pair.B, brName)
	}

	if err := bridge.Up(brName, true); err != nil {
		log.Printf("Warning: Failed to bring bridge up: %v", err)
	} else {
		fmt.Printf("✓ Bridge is UP\n")
	}

	brInfo, _ := bridge.Get(brName)
	fmt.Printf("✓ Bridge has %d member(s)\n", len(brInfo.Members))
	fmt.Println()

	// Demo 5: LAGG (Link Aggregation)
	fmt.Println("--- Demo 5: LAGG (Link Aggregation) ---")
	laggName, err := lagg.Create()
	if err != nil {
		log.Fatalf("Failed to create LAGG: %v", err)
	}
	fmt.Printf("✓ Created LAGG interface: %s\n", laggName)
	defer lagg.Destroy(laggName)

	if err := lagg.SetProto(laggName, lagg.ProtoFailover); err != nil {
		log.Printf("Warning: Failed to set protocol: %v", err)
	} else {
		fmt.Printf("✓ Set LAGG protocol to failover\n")
	}

	laggInfo, _ := lagg.Get(laggName)
	fmt.Printf("✓ LAGG Info: protocol=%s, ports=%d\n", laggInfo.Proto.String(), len(laggInfo.Ports))
	fmt.Println()

	// Demo 6: IP Address Management (IPv4)
	fmt.Println("--- Demo 6: IPv4 Address Management ---")
	loIface, err := ifc.Get("lo0")
	if err != nil {
		log.Printf("Warning: Failed to get lo0: %v", err)
	} else {
		fmt.Printf("✓ Interface lo0: %d existing addresses\n", len(loIface.Addrs))
	}

	testIP := net.ParseIP("192.0.2.100")
	testMask := net.IPv4Mask(255, 255, 255, 0)

	if err := ip.Add4("lo0", testIP, testMask); err != nil {
		log.Printf("Warning: Failed to add IPv4: %v", err)
	} else {
		fmt.Printf("✓ Added IPv4 address 192.0.2.100/24 to lo0\n")
		defer ip.Del4("lo0", testIP, testMask)
	}
	fmt.Println()

	// Demo 7: IP Address Management (IPv6)
	fmt.Println("--- Demo 7: IPv6 Address Management ---")
	testIP6 := net.ParseIP("2001:db8::100")

	if err := ip.Add6("lo0", testIP6, 64); err != nil {
		log.Printf("Warning: Failed to add IPv6: %v", err)
	} else {
		fmt.Printf("✓ Added IPv6 address 2001:db8::100/64 to lo0\n")
		defer ip.Del6("lo0", testIP6, 64)
	}
	fmt.Println()

	// Demo 8: IPv4 Routing
	fmt.Println("--- Demo 8: IPv4 Routing ---")
	_, testNet, _ := net.ParseCIDR("198.51.100.0/24")
	gwIP := net.ParseIP("127.0.0.1")

	if err := route.AddRoute4(testNet, gwIP, "lo0"); err != nil {
		log.Printf("Warning: Failed to add IPv4 route: %v", err)
	} else {
		fmt.Printf("✓ Added IPv4 route to 198.51.100.0/24 via 127.0.0.1\n")
		defer route.DelRoute4(testNet, gwIP, "lo0")
	}
	fmt.Println()

	// Demo 9: IPv6 Routing
	fmt.Println("--- Demo 9: IPv6 Routing ---")
	_, testNet6, _ := net.ParseCIDR("2001:db8:1::/48")
	gwIP6 := net.ParseIP("::1")

	if err := route.AddRoute6(testNet6, gwIP6, "lo0"); err != nil {
		log.Printf("Warning: Failed to add IPv6 route: %v", err)
	} else {
		fmt.Printf("✓ Added IPv6 route to 2001:db8:1::/48 via ::1\n")
		defer route.DelRoute6(testNet6, gwIP6, "lo0")
	}
	fmt.Println()

	// Summary
	fmt.Println("==========================================================")
	fmt.Println("  Demo Complete!")
	fmt.Println("==========================================================")
	fmt.Println()
	fmt.Println("Summary of created interfaces:")
	fmt.Printf("  - TAP:    %s (Layer 2 virtual)\n", tapName)
	fmt.Printf("  - TUN:    %s (Layer 3 virtual)\n", tunName)
	fmt.Printf("  - VLAN:   %s (802.1Q tag 100 on lo0)\n", vlanName)
	fmt.Printf("  - Bridge: %s (with member %s)\n", brName, pair.B)
	fmt.Printf("  - Epair:  %s <-> %s\n", pair.A, pair.B)
	fmt.Printf("  - LAGG:   %s (failover protocol)\n", laggName)
	fmt.Println()
	fmt.Println("All interfaces will be cleaned up automatically...")
}
