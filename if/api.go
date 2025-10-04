//go:build freebsd
// +build freebsd

package ifc

import (
	"github.com/zombocoder/go-freebsd-ifc/internal/ifops"
)

// List returns all network interfaces on the system.
//
// This function enumerates all network interfaces including physical (em0, igb0),
// virtual (bridge, vlan, epair), and loopback (lo0) interfaces.
//
// The returned interfaces include their current configuration (MTU, flags) and
// all assigned IP addresses (IPv4 and IPv6).
//
// Example:
//
//	ifaces, err := ifc.List()
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, iface := range ifaces {
//		fmt.Printf("%s [%d]: MTU=%d\n", iface.Name, iface.Index, iface.MTU)
//		for _, addr := range iface.Addrs {
//			fmt.Printf("  %s\n", addr.String())
//		}
//	}
func List() ([]Interface, error) {
	ifaces, err := ifops.List()
	if err != nil {
		return nil, err
	}

	result := make([]Interface, len(ifaces))
	for i, iface := range ifaces {
		result[i] = Interface{
			Name:  iface.Name,
			Index: iface.Index,
			MTU:   iface.MTU,
			Flags: InterfaceFlags(iface.Flags),
			Addrs: iface.Addrs,
		}
	}
	return result, nil
}

// Get returns a specific interface by name.
//
// Returns ErrNotFound if the interface does not exist.
//
// Example:
//
//	iface, err := ifc.Get("em0")
//	if err == ifc.ErrNotFound {
//		fmt.Println("Interface not found")
//	} else if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("MTU: %d, Up: %v\n", iface.MTU, iface.Flags.IsUp())
func Get(name string) (*Interface, error) {
	ifaces, err := List()
	if err != nil {
		return nil, err
	}
	for i := range ifaces {
		if ifaces[i].Name == name {
			return &ifaces[i], nil
		}
	}
	return nil, ErrNotFound
}

// SetUp brings an interface up or down.
//
// Requires root privileges. Returns ErrPermission if not running as root.
//
// Example:
//
//	// Bring interface up
//	if err := ifc.SetUp("em0", true); err != nil {
//		log.Fatal(err)
//	}
//
//	// Bring interface down
//	if err := ifc.SetUp("em0", false); err != nil {
//		log.Fatal(err)
//	}
func SetUp(name string, up bool) error {
	return ifops.SetFlags(name, uint32(FlagUp), up)
}

// SetMTU sets the Maximum Transmission Unit (MTU) of an interface.
//
// The MTU determines the maximum size of packets that can be transmitted
// on this interface. Common values are 1500 (Ethernet) or 9000 (Jumbo frames).
//
// Requires root privileges.
//
// Example:
//
//	// Set jumbo frames
//	if err := ifc.SetMTU("em0", 9000); err != nil {
//		log.Fatal(err)
//	}
func SetMTU(name string, mtu int) error {
	return ifops.SetMTU(name, mtu)
}

// Rename changes the name of an interface.
//
// This is useful for creating more meaningful interface names or
// implementing custom naming schemes.
//
// Requires root privileges. Returns ErrNotFound if the old interface
// name doesn't exist, or ErrExists if the new name is already in use.
//
// Example:
//
//	// Rename em0 to wan0
//	if err := ifc.Rename("em0", "wan0"); err != nil {
//		log.Fatal(err)
//	}
func Rename(oldName, newName string) error {
	return ifops.Rename(oldName, newName)
}
