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

// MoveToVNET moves an interface from the current vnet into the vnet of the
// jail with the given JID. The jail must have been created with vnet.
//
// This is the equivalent of "ifconfig <name> vnet <jid>". After the move the
// interface is no longer visible on the host; use ReclaimFromVNET to get it
// back, or destroy the jail (which returns it automatically).
//
// Requires root privileges. Returns ErrNotFound if the interface or jail
// does not exist.
//
// Example:
//
//	if err := ifc.MoveToVNET("epair0b", jid); err != nil {
//		log.Fatal(err)
//	}
func MoveToVNET(name string, jid int) error {
	return ifops.SetVNET(name, jid)
}

// ReclaimFromVNET moves an interface out of the vnet of the jail with the
// given JID back into the current vnet. name is the interface name as seen
// inside the jail.
//
// This is the equivalent of "ifconfig <name> -vnet <jid>".
//
// Requires root privileges.
//
// Example:
//
//	if err := ifc.ReclaimFromVNET("epair0b", jid); err != nil {
//		log.Fatal(err)
//	}
func ReclaimFromVNET(name string, jid int) error {
	return ifops.ReclaimVNET(name, jid)
}

// SetDescription sets the description of an interface, the text shown by
// "ifconfig <name>" as "description:". An empty string clears it.
//
// The kernel limits the size to net.ifdescr_maxlen (default 1024) bytes
// including a terminating NUL, so the longest description is 1023 bytes by
// default; a longer one is refused and the old description is kept. The
// description must not contain NUL bytes.
//
// Requires root privileges.
//
// Example:
//
//	// Mark a tap interface
//	if err := ifc.SetDescription("tap0", "microvm-runner vm=abc123"); err != nil {
//		log.Fatal(err)
//	}
//
//	// Remove the description
//	if err := ifc.SetDescription("tap0", ""); err != nil {
//		log.Fatal(err)
//	}
func SetDescription(name, descr string) error {
	return ifops.SetDescription(name, descr)
}

// Description returns the description of an interface, or "" if none is set.
//
// Returns ErrNotFound if the interface does not exist.
//
// Example:
//
//	descr, err := ifc.Description("tap0")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("tap0: %q\n", descr)
func Description(name string) (string, error) {
	return ifops.GetDescription(name)
}
