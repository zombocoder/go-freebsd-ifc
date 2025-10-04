//go:build freebsd
// +build freebsd

package ifc

import (
	"github.com/zombocoder/go-freebsd-ifc/internal/ifops"
)

// SetPromisc enables or disables promiscuous mode on an interface.
//
// Promiscuous mode allows the interface to receive all packets on the network,
// not just those destined for its MAC address. This is commonly used for
// packet capture, network monitoring, and intrusion detection.
//
// Requires root privileges.
//
// Example:
//
//	// Enable promiscuous mode for packet capture
//	if err := ifc.SetPromisc("em0", true); err != nil {
//		log.Fatal(err)
//	}
//
//	// Disable promiscuous mode
//	if err := ifc.SetPromisc("em0", false); err != nil {
//		log.Fatal(err)
//	}
func SetPromisc(name string, enable bool) error {
	return ifops.SetFlags(name, uint32(FlagPromisc), enable)
}

// IsPromisc checks if an interface is in promiscuous mode.
//
// Example:
//
//	promisc, err := ifc.IsPromisc("em0")
//	if err != nil {
//		log.Fatal(err)
//	}
//	if promisc {
//		fmt.Println("Interface is in promiscuous mode")
//	}
func IsPromisc(name string) (bool, error) {
	iface, err := Get(name)
	if err != nil {
		return false, err
	}
	return iface.Flags&FlagPromisc != 0, nil
}
