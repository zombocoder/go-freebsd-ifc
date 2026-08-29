//go:build freebsd
// +build freebsd

// vxlan-demo drives the vxlan package from the command line.
//
// It needs the vxlan(4) driver: doas kldload if_vxlan
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/vxlan"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		listVXLANs()

	case "create":
		if len(os.Args) < 5 {
			fmt.Println("Usage: vxlan-demo create <vni> <local-ip> <remote-ip> [port]")
			os.Exit(1)
		}
		port := uint16(vxlan.DefaultPort)
		if len(os.Args) >= 6 {
			port = parsePort(os.Args[5])
		}
		createVXLAN(parseVNI(os.Args[2]), parseIP(os.Args[3]), parseIP(os.Args[4]), port)

	case "destroy":
		if len(os.Args) < 3 {
			fmt.Println("Usage: vxlan-demo destroy <name>")
			os.Exit(1)
		}
		destroyVXLAN(os.Args[2])

	case "show":
		if len(os.Args) < 3 {
			fmt.Println("Usage: vxlan-demo show <name>")
			os.Exit(1)
		}
		showVXLAN(os.Args[2])

	case "peer-add":
		if len(os.Args) < 5 {
			fmt.Println("Usage: vxlan-demo peer-add <name> <mac> <remote-ip> [port]")
			os.Exit(1)
		}
		port := uint16(0)
		if len(os.Args) >= 6 {
			port = parsePort(os.Args[5])
		}
		peerAdd(os.Args[2], os.Args[3], parseIP(os.Args[4]), port)

	case "peer-del":
		if len(os.Args) < 4 {
			fmt.Println("Usage: vxlan-demo peer-del <name> <mac>")
			os.Exit(1)
		}
		peerDel(os.Args[2], os.Args[3])

	case "peers":
		if len(os.Args) < 3 {
			fmt.Println("Usage: vxlan-demo peers <name>")
			os.Exit(1)
		}
		listPeers(os.Args[2])

	case "flush":
		if len(os.Args) < 3 {
			fmt.Println("Usage: vxlan-demo flush <name> [all]")
			os.Exit(1)
		}
		flushPeers(os.Args[2], len(os.Args) >= 4 && os.Args[3] == "all")

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("VXLAN Overlay Management Demo")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  vxlan-demo list                                        # List VXLAN interfaces")
	fmt.Println("  vxlan-demo show <name>                                 # Show one interface")
	fmt.Println("  vxlan-demo peers <name>                                # Show the forwarding table")
	fmt.Println("  vxlan-demo create <vni> <local> <remote> [port]        # Create + configure + up (root)")
	fmt.Println("  vxlan-demo destroy <name>                              # Destroy (root)")
	fmt.Println("  vxlan-demo peer-add <name> <mac> <remote> [port]       # Add a static peer (root)")
	fmt.Println("  vxlan-demo peer-del <name> <mac>                       # Remove a static peer (root)")
	fmt.Println("  vxlan-demo flush <name> [all]                          # Flush learned [and static] (root)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  doas kldload if_vxlan")
	fmt.Println("  doas vxlan-demo create 100 192.0.2.1 192.0.2.2")
	fmt.Println("  vxlan-demo show vxlan0")
	fmt.Println("  doas vxlan-demo peer-add vxlan0 02:11:22:33:44:55 192.0.2.3")
	fmt.Println("  vxlan-demo peers vxlan0")
	fmt.Println("  doas vxlan-demo destroy vxlan0")
	fmt.Println()
	fmt.Println("Peers are FreeBSD forwarding table entries: a remote MAC address")
	fmt.Println("mapped to the VTEP that owns it. ifconfig(8) cannot add them.")
}

func listVXLANs() {
	ifaces, err := ifc.List()
	if err != nil {
		log.Fatalf("list interfaces: %v", err)
	}

	found := 0
	for _, iface := range ifaces {
		cfg, err := vxlan.Get(iface.Name)
		if err != nil {
			continue // not a VXLAN interface
		}
		found++
		fmt.Printf("%s: vni %s, %s -> %s, mtu %d%s\n",
			cfg.Name, vniString(cfg.VNI),
			endpoint(cfg.Local, cfg.LocalPort), endpoint(cfg.Remote, cfg.RemotePort),
			cfg.MTU, upState(cfg))
	}
	if found == 0 {
		fmt.Println("No VXLAN interfaces found.")
		fmt.Println("(If you expected some, the driver may not be loaded: doas kldload if_vxlan)")
	}
}

func createVXLAN(vni uint32, local, remote net.IP, port uint16) {
	name, err := vxlan.Create()
	if err != nil {
		log.Fatalf("create: %v", err)
	}
	fmt.Printf("Created %s\n", name)

	err = vxlan.Configure(name, vxlan.Params{
		VNI:        vni,
		Local:      local,
		Remote:     remote,
		LocalPort:  port,
		RemotePort: port,
	})
	if err != nil {
		// Leave nothing behind if the configuration is rejected.
		if derr := vxlan.Destroy(name); derr != nil {
			log.Printf("destroy %s after a failed configure: %v", name, derr)
		}
		log.Fatalf("configure: %v", err)
	}
	fmt.Printf("Configured vni %d, %s -> %s on port %d\n", vni, local, remote, port)

	if err := vxlan.Up(name, true); err != nil {
		log.Fatalf("up: %v", err)
	}

	// Setting IFF_UP always succeeds: vxlan_init() in the kernel is void, so
	// an invalid configuration leaves the interface UP but never RUNNING.
	cfg, err := vxlan.Get(name)
	if err != nil {
		log.Fatalf("get: %v", err)
	}
	if !cfg.Running {
		fmt.Printf("Warning: %s is up but not running; the kernel rejected the configuration\n", name)
		fmt.Println("         (check the console log; the local address must be an address of this host)")
	}
	showVXLAN(name)
}

func destroyVXLAN(name string) {
	if err := vxlan.Destroy(name); err != nil {
		log.Fatalf("destroy: %v", err)
	}
	fmt.Printf("Destroyed %s\n", name)
}

func showVXLAN(name string) {
	cfg, err := vxlan.Get(name)
	if err != nil {
		log.Fatalf("get: %v", err)
	}

	fmt.Printf("%s:\n", cfg.Name)
	fmt.Printf("  VNI:        %s\n", vniString(cfg.VNI))
	fmt.Printf("  Local:      %s\n", endpoint(cfg.Local, cfg.LocalPort))
	fmt.Printf("  Remote:     %s", endpoint(cfg.Remote, cfg.RemotePort))
	if cfg.Multicast {
		fmt.Printf(" (multicast group)")
	}
	fmt.Println()
	if cfg.Dev != "" {
		fmt.Printf("  Multicast device: %s\n", cfg.Dev)
	}
	fmt.Printf("  Port range: %d-%d\n", cfg.PortRangeMin, cfg.PortRangeMax)
	fmt.Printf("  TTL:        %d\n", cfg.TTL)
	fmt.Printf("  Learning:   %v\n", cfg.Learn)
	fmt.Printf("  Peers:      %d of %d, timeout %ds\n", cfg.PeerCount, cfg.PeerMax, cfg.PeerTimeout)
	fmt.Printf("  MTU:        %d\n", cfg.MTU)
	fmt.Printf("  State:      up=%v running=%v\n", cfg.Up, cfg.Running)

	if overhead, err := vxlan.Overhead(name); err == nil {
		fmt.Printf("  Encap overhead: %d bytes (kernel if_hdrlen; %d + %d = %d)\n",
			overhead, cfg.MTU, overhead, cfg.MTU+overhead)
	}
}

func peerAdd(name, macStr string, remote net.IP, port uint16) {
	mac := parseMAC(macStr)
	if err := vxlan.PeerAdd(name, mac, remote, port); err != nil {
		log.Fatalf("peer-add: %v", err)
	}
	fmt.Printf("Added peer %s via %s on %s\n", mac, remote, name)
}

func peerDel(name, macStr string) {
	mac := parseMAC(macStr)
	if err := vxlan.PeerDel(name, mac); err != nil {
		log.Fatalf("peer-del: %v", err)
	}
	fmt.Printf("Removed peer %s from %s\n", mac, name)
}

func listPeers(name string) {
	peers, err := vxlan.Peers(name)
	if err != nil {
		log.Fatalf("peers: %v", err)
	}
	if len(peers) == 0 {
		fmt.Printf("%s has no forwarding table entries.\n", name)
		return
	}
	fmt.Printf("%s forwarding table:\n", name)
	fmt.Printf("  %-18s %-40s %s\n", "MAC", "REMOTE", "KIND")
	for _, p := range peers {
		kind := "learned"
		if p.Static {
			kind = "static"
		}
		fmt.Printf("  %-18s %-40s %s\n", p.MAC, p.Remote, kind)
	}
}

func flushPeers(name string, all bool) {
	if err := vxlan.FlushPeers(name, all); err != nil {
		log.Fatalf("flush: %v", err)
	}
	if all {
		fmt.Printf("Flushed all forwarding table entries on %s\n", name)
	} else {
		fmt.Printf("Flushed learned forwarding table entries on %s\n", name)
	}
}

// vniString renders the VNI, spelling out the kernel's "not configured" value.
func vniString(vni uint32) string {
	if vni == uint32(vxlan.VNIMax) {
		return "unset"
	}
	return strconv.FormatUint(uint64(vni), 10)
}

func endpoint(ip net.IP, port uint16) string {
	if ip == nil {
		return fmt.Sprintf("(unset):%d", port)
	}
	return net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))
}

func upState(cfg vxlan.Config) string {
	switch {
	case cfg.Running:
		return " [running]"
	case cfg.Up:
		return " [up, not running]"
	default:
		return " [down]"
	}
}

func parseVNI(s string) uint32 {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		log.Fatalf("invalid VNI %q: %v", s, err)
	}
	if v >= uint64(vxlan.VNIMax) {
		log.Fatalf("invalid VNI %d: must be less than %d", v, uint32(vxlan.VNIMax))
	}
	return uint32(v)
}

func parsePort(s string) uint16 {
	p, err := strconv.ParseUint(s, 10, 16)
	if err != nil || p == 0 {
		log.Fatalf("invalid port %q", s)
	}
	return uint16(p)
}

func parseIP(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		log.Fatalf("invalid IP address %q", s)
	}
	return ip
}

func parseMAC(s string) net.HardwareAddr {
	mac, err := net.ParseMAC(s)
	if err != nil {
		log.Fatalf("invalid MAC address %q: %v", s, err)
	}
	return mac
}
