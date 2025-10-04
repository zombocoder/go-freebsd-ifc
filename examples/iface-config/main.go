//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "show":
		if len(os.Args) < 3 {
			fmt.Println("Usage: iface-config show <interface>")
			os.Exit(1)
		}
		showInterface(os.Args[2])
	case "mtu":
		if len(os.Args) < 4 {
			fmt.Println("Usage: iface-config mtu <interface> <mtu>")
			os.Exit(1)
		}
		setMTU(os.Args[2], os.Args[3])
	case "up":
		if len(os.Args) < 3 {
			fmt.Println("Usage: iface-config up <interface>")
			os.Exit(1)
		}
		setUp(os.Args[2], true)
	case "down":
		if len(os.Args) < 3 {
			fmt.Println("Usage: iface-config down <interface>")
			os.Exit(1)
		}
		setUp(os.Args[2], false)
	case "promisc":
		if len(os.Args) < 4 {
			fmt.Println("Usage: iface-config promisc <interface> <on|off>")
			os.Exit(1)
		}
		setPromisc(os.Args[2], os.Args[3])
	case "rename":
		if len(os.Args) < 4 {
			fmt.Println("Usage: iface-config rename <old-name> <new-name>")
			os.Exit(1)
		}
		renameInterface(os.Args[2], os.Args[3])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Interface Configuration Management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  iface-config show <interface>             # Show interface details")
	fmt.Println("  iface-config mtu <interface> <mtu>        # Set MTU (requires root)")
	fmt.Println("  iface-config up <interface>               # Bring interface up (requires root)")
	fmt.Println("  iface-config down <interface>             # Bring interface down (requires root)")
	fmt.Println("  iface-config promisc <interface> <on|off> # Set promiscuous mode (requires root)")
	fmt.Println("  iface-config rename <old> <new>           # Rename interface (requires root)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Show interface details")
	fmt.Println("  go run examples/iface-config/main.go show em0")
	fmt.Println()
	fmt.Println("  # Set MTU to 9000 (jumbo frames)")
	fmt.Println("  doas go run examples/iface-config/main.go mtu em0 9000")
	fmt.Println()
	fmt.Println("  # Bring interface up/down")
	fmt.Println("  doas go run examples/iface-config/main.go up em0")
	fmt.Println("  doas go run examples/iface-config/main.go down em0")
	fmt.Println()
	fmt.Println("  # Enable promiscuous mode for packet capture")
	fmt.Println("  doas go run examples/iface-config/main.go promisc em0 on")
	fmt.Println()
	fmt.Println("  # Rename interface")
	fmt.Println("  doas go run examples/iface-config/main.go rename em0 wan0")
}

func showInterface(name string) {
	iface, err := ifc.Get(name)
	if err == ifc.ErrNotFound {
		log.Fatalf("Interface not found: %s", name)
	} else if err != nil {
		log.Fatalf("Failed to get interface: %v", err)
	}

	fmt.Printf("Interface: %s\n", iface.Name)
	fmt.Println("==================")
	fmt.Printf("  Index:      %d\n", iface.Index)
	fmt.Printf("  MTU:        %d\n", iface.MTU)
	fmt.Printf("  State:      %s\n", getState(iface.Flags))
	fmt.Printf("  Flags:      %s\n", getFlags(iface.Flags))
	fmt.Println()

	if len(iface.Addrs) > 0 {
		fmt.Println("  Addresses:")
		for _, addr := range iface.Addrs {
			fmt.Printf("    %s\n", addr.String())
		}
	} else {
		fmt.Println("  Addresses:  (none)")
	}
}

func getState(flags ifc.InterfaceFlags) string {
	if flags.IsUp() && flags.IsRunning() {
		return "UP,RUNNING"
	} else if flags.IsUp() {
		return "UP"
	}
	return "DOWN"
}

func getFlags(flags ifc.InterfaceFlags) string {
	var flagList []string

	if flags.IsUp() {
		flagList = append(flagList, "UP")
	}
	if flags.IsRunning() {
		flagList = append(flagList, "RUNNING")
	}
	if flags.IsLoopback() {
		flagList = append(flagList, "LOOPBACK")
	}
	if flags&ifc.FlagBroadcast != 0 {
		flagList = append(flagList, "BROADCAST")
	}
	if flags&ifc.FlagMulticast != 0 {
		flagList = append(flagList, "MULTICAST")
	}
	if flags&ifc.FlagPromisc != 0 {
		flagList = append(flagList, "PROMISC")
	}
	if flags&ifc.FlagPointToPoint != 0 {
		flagList = append(flagList, "POINTOPOINT")
	}
	if flags&ifc.FlagNoARP != 0 {
		flagList = append(flagList, "NOARP")
	}

	if len(flagList) == 0 {
		return "NONE"
	}

	result := flagList[0]
	for i := 1; i < len(flagList); i++ {
		result += "," + flagList[i]
	}
	return result
}

func setMTU(name, mtuStr string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	mtu, err := strconv.Atoi(mtuStr)
	if err != nil {
		log.Fatalf("Invalid MTU value: %s", mtuStr)
	}

	if mtu < 68 || mtu > 65535 {
		log.Fatal("MTU must be between 68 and 65535")
	}

	fmt.Printf("Setting MTU of %s to %d...\n", name, mtu)

	if err := ifc.SetMTU(name, mtu); err != nil {
		log.Fatalf("Failed to set MTU: %v", err)
	}

	fmt.Printf("✓ MTU set to %d\n", mtu)
}

func setUp(name string, up bool) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	action := "up"
	if !up {
		action = "down"
	}

	fmt.Printf("Bringing interface %s %s...\n", name, action)

	if err := ifc.SetUp(name, up); err != nil {
		log.Fatalf("Failed to bring interface %s: %v", action, err)
	}

	fmt.Printf("✓ Interface is %s\n", action)
}

func setPromisc(name, mode string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	var enable bool
	switch mode {
	case "on", "true", "1", "yes":
		enable = true
	case "off", "false", "0", "no":
		enable = false
	default:
		log.Fatalf("Invalid mode: %s (use 'on' or 'off')", mode)
	}

	action := "Enabling"
	if !enable {
		action = "Disabling"
	}

	fmt.Printf("%s promiscuous mode on %s...\n", action, name)

	if err := ifc.SetPromisc(name, enable); err != nil {
		log.Fatalf("Failed to set promiscuous mode: %v", err)
	}

	fmt.Printf("✓ Promiscuous mode %s\n", map[bool]string{true: "enabled", false: "disabled"}[enable])

	// Show current status
	promisc, _ := ifc.IsPromisc(name)
	if promisc {
		fmt.Println("  ⚠ Warning: Promiscuous mode allows capturing all network traffic")
	}
}

func renameInterface(oldName, newName string) {
	if os.Geteuid() != 0 {
		log.Fatal("This operation requires root privileges")
	}

	fmt.Printf("Renaming interface %s to %s...\n", oldName, newName)

	if err := ifc.Rename(oldName, newName); err != nil {
		log.Fatalf("Failed to rename interface: %v", err)
	}

	fmt.Printf("✓ Interface renamed from %s to %s\n", oldName, newName)
}
