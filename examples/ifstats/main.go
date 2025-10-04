//go:build freebsd
// +build freebsd

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

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
			fmt.Println("Usage: ifstats show <interface>")
			os.Exit(1)
		}
		showStats(os.Args[2])
	case "list":
		listAllStats()
	case "watch":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ifstats watch <interface> [interval]")
			os.Exit(1)
		}
		interval := 1
		if len(os.Args) >= 4 {
			fmt.Sscanf(os.Args[3], "%d", &interval)
		}
		watchStats(os.Args[2], interval)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Interface Statistics Viewer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ifstats show <interface>               # Show interface statistics")
	fmt.Println("  ifstats list                           # List statistics for all interfaces")
	fmt.Println("  ifstats watch <interface> [interval]   # Watch statistics (default 1s)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Show em0 statistics")
	fmt.Println("  go run examples/ifstats/main.go show em0")
	fmt.Println()
	fmt.Println("  # List all interface statistics")
	fmt.Println("  go run examples/ifstats/main.go list")
	fmt.Println()
	fmt.Println("  # Watch em0 statistics every 2 seconds")
	fmt.Println("  go run examples/ifstats/main.go watch em0 2")
}

func showStats(name string) {
	stats, err := ifc.GetStats(name)
	if err == ifc.ErrNotFound {
		log.Fatalf("Interface not found: %s", name)
	} else if err != nil {
		log.Fatalf("Failed to get statistics: %v", err)
	}

	fmt.Printf("Interface: %s\n", name)
	fmt.Println("==================")
	fmt.Println()
	fmt.Println("Receive (RX):")
	fmt.Printf("  Packets:   %s\n", formatNumber(stats.InPackets))
	fmt.Printf("  Bytes:     %s (%s)\n", formatNumber(stats.InBytes), formatBytes(stats.InBytes))
	fmt.Printf("  Errors:    %s\n", formatNumber(stats.InErrors))
	fmt.Printf("  Dropped:   %s\n", formatNumber(stats.InDropped))
	fmt.Printf("  Multicast: %s\n", formatNumber(stats.InMulticast))
	fmt.Println()
	fmt.Println("Transmit (TX):")
	fmt.Printf("  Packets:   %s\n", formatNumber(stats.OutPackets))
	fmt.Printf("  Bytes:     %s (%s)\n", formatNumber(stats.OutBytes), formatBytes(stats.OutBytes))
	fmt.Printf("  Errors:    %s\n", formatNumber(stats.OutErrors))
	fmt.Println()
	fmt.Println("Other:")
	fmt.Printf("  Collisions: %s\n", formatNumber(stats.Collisions))
}

func listAllStats() {
	ifaces, err := ifc.List()
	if err != nil {
		log.Fatalf("Failed to list interfaces: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Interface\tRX Packets\tRX Bytes\tRX Errors\tTX Packets\tTX Bytes\tTX Errors")
	fmt.Fprintln(w, strings.Repeat("-", 100))

	for _, iface := range ifaces {
		stats, err := ifc.GetStats(iface.Name)
		if err != nil {
			continue
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			iface.Name,
			formatNumber(stats.InPackets),
			formatBytes(stats.InBytes),
			formatNumber(stats.InErrors),
			formatNumber(stats.OutPackets),
			formatBytes(stats.OutBytes),
			formatNumber(stats.OutErrors))
	}

	w.Flush()
}

func watchStats(name string, interval int) {
	fmt.Printf("Watching %s statistics (Ctrl+C to stop)\n\n", name)

	var lastStats *ifc.Stats

	for {
		stats, err := ifc.GetStats(name)
		if err != nil {
			log.Fatalf("Failed to get statistics: %v", err)
		}

		// Clear screen (ANSI escape code)
		fmt.Print("\033[H\033[2J")

		fmt.Printf("Interface: %s (refreshing every %ds)\n", name, interval)
		fmt.Println(strings.Repeat("=", 60))

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "\tTotal\tRate/s")
		fmt.Fprintln(w, strings.Repeat("-", 60))

		if lastStats != nil {
			// Calculate rates
			rxPktsRate := (stats.InPackets - lastStats.InPackets) / uint64(interval)
			rxBytesRate := (stats.InBytes - lastStats.InBytes) / uint64(interval)
			txPktsRate := (stats.OutPackets - lastStats.OutPackets) / uint64(interval)
			txBytesRate := (stats.OutBytes - lastStats.OutBytes) / uint64(interval)

			fmt.Fprintf(w, "RX Packets:\t%s\t%s pps\n",
				formatNumber(stats.InPackets), formatNumber(rxPktsRate))
			fmt.Fprintf(w, "RX Bytes:\t%s\t%s/s\n",
				formatBytes(stats.InBytes), formatBytes(rxBytesRate))
			fmt.Fprintf(w, "RX Errors:\t%s\t\n", formatNumber(stats.InErrors))
			fmt.Fprintf(w, "RX Dropped:\t%s\t\n", formatNumber(stats.InDropped))
			fmt.Fprintln(w)
			fmt.Fprintf(w, "TX Packets:\t%s\t%s pps\n",
				formatNumber(stats.OutPackets), formatNumber(txPktsRate))
			fmt.Fprintf(w, "TX Bytes:\t%s\t%s/s\n",
				formatBytes(stats.OutBytes), formatBytes(txBytesRate))
			fmt.Fprintf(w, "TX Errors:\t%s\t\n", formatNumber(stats.OutErrors))
			fmt.Fprintln(w)
			fmt.Fprintf(w, "Collisions:\t%s\t\n", formatNumber(stats.Collisions))
			fmt.Fprintf(w, "Multicast:\t%s\t\n", formatNumber(stats.InMulticast))
		} else {
			fmt.Fprintf(w, "RX Packets:\t%s\t\n", formatNumber(stats.InPackets))
			fmt.Fprintf(w, "RX Bytes:\t%s\t\n", formatBytes(stats.InBytes))
			fmt.Fprintf(w, "RX Errors:\t%s\t\n", formatNumber(stats.InErrors))
			fmt.Fprintf(w, "RX Dropped:\t%s\t\n", formatNumber(stats.InDropped))
			fmt.Fprintln(w)
			fmt.Fprintf(w, "TX Packets:\t%s\t\n", formatNumber(stats.OutPackets))
			fmt.Fprintf(w, "TX Bytes:\t%s\t\n", formatBytes(stats.OutBytes))
			fmt.Fprintf(w, "TX Errors:\t%s\t\n", formatNumber(stats.OutErrors))
			fmt.Fprintln(w)
			fmt.Fprintf(w, "Collisions:\t%s\t\n", formatNumber(stats.Collisions))
			fmt.Fprintf(w, "Multicast:\t%s\t\n", formatNumber(stats.InMulticast))
		}

		w.Flush()

		lastStats = stats
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func formatNumber(n uint64) string {
	if n == 0 {
		return "0"
	}

	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
