//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/lagg"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		listLAGGs()
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: lagg-demo create <protocol>")
			fmt.Println("Protocols: failover, loadbalance, lacp, roundrobin, broadcast")
			os.Exit(1)
		}
		createLAGG(os.Args[2])
	case "add-port":
		if len(os.Args) < 4 {
			fmt.Println("Usage: lagg-demo add-port <lagg-name> <port-name>")
			os.Exit(1)
		}
		addPort(os.Args[2], os.Args[3])
	case "del-port":
		if len(os.Args) < 4 {
			fmt.Println("Usage: lagg-demo del-port <lagg-name> <port-name>")
			os.Exit(1)
		}
		delPort(os.Args[2], os.Args[3])
	case "show":
		if len(os.Args) < 3 {
			fmt.Println("Usage: lagg-demo show <lagg-name>")
			os.Exit(1)
		}
		showLAGG(os.Args[2])
	case "destroy":
		if len(os.Args) < 3 {
			fmt.Println("Usage: lagg-demo destroy <lagg-name>")
			os.Exit(1)
		}
		destroyLAGG(os.Args[2])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("LAGG (Link Aggregation) Management Demo")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  lagg-demo list                          # List all LAGG interfaces")
	fmt.Println("  lagg-demo create <protocol>             # Create LAGG (requires root)")
	fmt.Println("  lagg-demo add-port <lagg> <port>        # Add port to LAGG (requires root)")
	fmt.Println("  lagg-demo del-port <lagg> <port>        # Remove port from LAGG (requires root)")
	fmt.Println("  lagg-demo show <lagg>                   # Show LAGG details")
	fmt.Println("  lagg-demo destroy <lagg>                # Destroy LAGG (requires root)")
	fmt.Println()
	fmt.Println("Protocols:")
	fmt.Println("  failover     - Sends traffic through primary port, fails over to backup")
	fmt.Println("  loadbalance  - Balances traffic across ports using hash")
	fmt.Println("  lacp         - IEEE 802.3ad Link Aggregation Control Protocol")
	fmt.Println("  roundrobin   - Distributes traffic in round-robin fashion")
	fmt.Println("  broadcast    - Sends traffic on all ports simultaneously")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List all LAGG interfaces")
	fmt.Println("  go run examples/lagg-demo/main.go list")
	fmt.Println()
	fmt.Println("  # Create LAGG with LACP protocol")
	fmt.Println("  doas go run examples/lagg-demo/main.go create lacp")
	fmt.Println()
	fmt.Println("  # Add ports to LAGG")
	fmt.Println("  doas go run examples/lagg-demo/main.go add-port lagg0 em0")
	fmt.Println("  doas go run examples/lagg-demo/main.go add-port lagg0 em1")
	fmt.Println()
	fmt.Println("  # Show LAGG details")
	fmt.Println("  go run examples/lagg-demo/main.go show lagg0")
	fmt.Println()
	fmt.Println("  # Remove port from LAGG")
	fmt.Println("  doas go run examples/lagg-demo/main.go del-port lagg0 em1")
	fmt.Println()
	fmt.Println("  # Destroy LAGG")
	fmt.Println("  doas go run examples/lagg-demo/main.go destroy lagg0")
}

func parseProtocol(protoStr string) (lagg.Proto, error) {
	switch strings.ToLower(protoStr) {
	case "failover":
		return lagg.ProtoFailover, nil
	case "loadbalance":
		return lagg.ProtoLoadBalance, nil
	case "lacp":
		return lagg.ProtoLACP, nil
	case "roundrobin":
		return lagg.ProtoRoundRobin, nil
	case "broadcast":
		return lagg.ProtoBroadcast, nil
	default:
		return 0, fmt.Errorf("unknown protocol: %s", protoStr)
	}
}

func createLAGG(protoStr string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	proto, err := parseProtocol(protoStr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Creating LAGG interface with protocol %s...\n", proto.String())

	name, err := lagg.Create()
	if err != nil {
		log.Fatalf("Failed to create LAGG: %v", err)
	}
	fmt.Printf("✓ Created LAGG interface: %s\n", name)

	fmt.Printf("Setting protocol to %s...\n", proto.String())
	if err := lagg.SetProto(name, proto); err != nil {
		lagg.Destroy(name)
		log.Fatalf("Failed to set protocol: %v", err)
	}
	fmt.Println("✓ Protocol set successfully")

	fmt.Println("Bringing LAGG interface up...")
	if err := lagg.Up(name, true); err != nil {
		log.Printf("Warning: Failed to bring LAGG up: %v", err)
	} else {
		fmt.Println("✓ LAGG interface is up")
	}

	fmt.Println()
	showLAGG(name)
}

func addPort(laggName, portName string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	fmt.Printf("Adding port %s to LAGG %s...\n", portName, laggName)

	if err := lagg.AddPort(laggName, portName); err != nil {
		log.Fatalf("Failed to add port: %v", err)
	}

	fmt.Printf("✓ Port %s added to LAGG %s\n", portName, laggName)
	fmt.Println()
	showLAGG(laggName)
}

func delPort(laggName, portName string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	fmt.Printf("Removing port %s from LAGG %s...\n", portName, laggName)

	if err := lagg.DelPort(laggName, portName); err != nil {
		log.Fatalf("Failed to remove port: %v", err)
	}

	fmt.Printf("✓ Port %s removed from LAGG %s\n", portName, laggName)
	fmt.Println()
	showLAGG(laggName)
}

func destroyLAGG(name string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	fmt.Printf("Destroying LAGG %s...\n", name)

	if err := lagg.Destroy(name); err != nil {
		log.Fatalf("Failed to destroy LAGG: %v", err)
	}

	fmt.Printf("✓ LAGG %s destroyed\n", name)
}

func showLAGG(name string) {
	info, err := lagg.Get(name)
	if err != nil {
		log.Fatalf("Failed to get LAGG info: %v", err)
	}

	state := "DOWN"
	if info.Up {
		state = "UP"
	}

	fmt.Printf("LAGG Interface: %s\n", name)
	fmt.Println("===================")
	fmt.Printf("  Protocol:   %s\n", info.Proto.String())
	fmt.Printf("  State:      %s\n", state)
	fmt.Printf("  MTU:        %d\n", info.MTU)
	fmt.Printf("  Ports (%d):\n", len(info.Ports))
	if len(info.Ports) > 0 {
		for _, port := range info.Ports {
			fmt.Printf("    - %s\n", port)
		}
	} else {
		fmt.Println("    (no ports)")
	}
}

func listLAGGs() {
	ifaces, err := ifc.List()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}

	fmt.Println("LAGG Interfaces:")
	fmt.Println("================")
	fmt.Println()

	found := false
	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, "lagg") {
			info, err := lagg.Get(iface.Name)
			if err != nil {
				continue
			}

			found = true
			state := "DOWN"
			if iface.Flags.IsUp() {
				state = "UP"
			}

			fmt.Printf("%s:\n", iface.Name)
			fmt.Printf("  Protocol: %s\n", info.Proto.String())
			fmt.Printf("  State:    %s\n", state)
			fmt.Printf("  MTU:      %d\n", iface.MTU)
			fmt.Printf("  Ports:    %d", len(info.Ports))
			if len(info.Ports) > 0 {
				fmt.Printf(" [%s]", strings.Join(info.Ports, ", "))
			}
			fmt.Println()
			fmt.Println()
		}
	}

	if !found {
		fmt.Println("No LAGG interfaces found")
	}
}
