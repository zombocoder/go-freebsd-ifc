//go:build freebsd
// +build freebsd

package ifc

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <ifaddrs.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
)

// Stats represents interface statistics (packet and byte counters).
type Stats struct {
	InPackets   uint64 // Packets received
	InBytes     uint64 // Bytes received
	InErrors    uint64 // Input errors
	InDropped   uint64 // Input packets dropped
	OutPackets  uint64 // Packets transmitted
	OutBytes    uint64 // Bytes transmitted
	OutErrors   uint64 // Output errors
	OutDropped  uint64 // Output packets dropped
	Collisions  uint64 // Collisions on transmit
	InMulticast uint64 // Multicast packets received
}

// GetStats returns interface statistics.
//
// Returns packet and byte counters for the interface using getifaddrs()
// which provides access to the if_data structure.
//
// Example:
//
//	stats, err := ifc.GetStats("em0")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Received: %d packets, %d bytes\n", stats.InPackets, stats.InBytes)
//	fmt.Printf("Transmitted: %d packets, %d bytes\n", stats.OutPackets, stats.OutBytes)
//	fmt.Printf("Errors: %d in, %d out\n", stats.InErrors, stats.OutErrors)
func GetStats(name string) (*Stats, error) {
	var ifap *C.struct_ifaddrs
	if C.getifaddrs(&ifap) != 0 {
		return nil, fmt.Errorf("getifaddrs failed")
	}
	defer C.freeifaddrs(ifap)

	// Find the interface and get its if_data
	for ifa := ifap; ifa != nil; ifa = ifa.ifa_next {
		ifname := C.GoString(ifa.ifa_name)
		if ifname != name {
			continue
		}

		// Check if this is a link-layer address (AF_LINK)
		if ifa.ifa_addr != nil && ifa.ifa_addr.sa_family == C.AF_LINK {
			// Get if_data from ifa_data
			if ifa.ifa_data != nil {
				ifdata := (*C.struct_if_data)(ifa.ifa_data)

				return &Stats{
					InPackets:   uint64(ifdata.ifi_ipackets),
					InBytes:     uint64(ifdata.ifi_ibytes),
					InErrors:    uint64(ifdata.ifi_ierrors),
					InDropped:   uint64(ifdata.ifi_iqdrops),
					OutPackets:  uint64(ifdata.ifi_opackets),
					OutBytes:    uint64(ifdata.ifi_obytes),
					OutErrors:   uint64(ifdata.ifi_oerrors),
					Collisions:  uint64(ifdata.ifi_collisions),
					InMulticast: uint64(ifdata.ifi_imcasts),
					// Note: OutDropped is not typically available in if_data
				}, nil
			}
		}
	}

	return nil, ErrNotFound
}
