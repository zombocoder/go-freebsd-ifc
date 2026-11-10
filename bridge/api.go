//go:build freebsd
// +build freebsd

package bridge

import (
	"fmt"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/internal/bridgeops"
	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
	"github.com/zombocoder/go-freebsd-ifc/internal/ifops"
)

// Create creates a new bridge interface.
//
// The kernel automatically assigns a name (e.g., "bridge0", "bridge1").
// Requires root privileges.
//
// Example:
//
//	br, err := bridge.Create()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Created: %s\n", br)
//	defer bridge.Destroy(br)
func Create() (string, error) {
	name, err := cloneops.Create("bridge")
	if err != nil {
		return "", fmt.Errorf("create bridge: %w", err)
	}
	return name, nil
}

// Destroy destroys a bridge interface.
//
// The bridge must have no member interfaces before it can be destroyed.
// Requires root privileges.
//
// Example:
//
//	if err := bridge.Destroy("bridge0"); err != nil {
//		log.Fatal(err)
//	}
func Destroy(name string) error {
	if err := cloneops.Destroy(name); err != nil {
		return fmt.Errorf("destroy bridge %s: %w", name, err)
	}
	return nil
}

// Up brings the bridge interface up or down.
//
// A bridge must be "up" to forward traffic between its member interfaces.
// Requires root privileges.
//
// Example:
//
//	// Bring bridge up
//	if err := bridge.Up("bridge0", true); err != nil {
//		log.Fatal(err)
//	}
func Up(name string, up bool) error {
	if err := ifops.SetFlags(name, uint32(ifc.FlagUp), up); err != nil {
		if up {
			return fmt.Errorf("bring bridge %s up: %w", name, err)
		}
		return fmt.Errorf("bring bridge %s down: %w", name, err)
	}
	return nil
}

// AddMember adds a member interface to the bridge.
//
// The member interface will forward traffic to other bridge members.
// This operation is idempotent - returns nil if the member already exists.
// Requires root privileges.
//
// Example:
//
//	if err := bridge.AddMember("bridge0", "em0"); err != nil {
//		log.Fatal(err)
//	}
func AddMember(bridge, member string) error {
	if err := bridgeops.AddMember(bridge, member); err != nil {
		return fmt.Errorf("add member %s to bridge %s: %w", member, bridge, err)
	}
	return nil
}

// DelMember removes a member interface from the bridge.
//
// This operation is idempotent - returns nil if the member doesn't exist.
// Requires root privileges.
//
// Example:
//
//	if err := bridge.DelMember("bridge0", "em0"); err != nil {
//		log.Fatal(err)
//	}
func DelMember(bridge, member string) error {
	if err := bridgeops.DelMember(bridge, member); err != nil {
		return fmt.Errorf("remove member %s from bridge %s: %w", member, bridge, err)
	}
	return nil
}

// Members returns the list of member interfaces in the bridge.
//
// Example:
//
//	members, err := bridge.Members("bridge0")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Bridge members: %v\n", members)
func Members(bridge string) ([]string, error) {
	members, err := bridgeops.GetMembers(bridge)
	if err != nil {
		return nil, fmt.Errorf("get members of bridge %s: %w", bridge, err)
	}
	return members, nil
}

// Get returns complete bridge information including members, MTU, and status.
//
// Example:
//
//	info, err := bridge.Get("bridge0")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("%s: %d members, MTU=%d, Up=%v\n",
//		info.Name, len(info.Members), info.MTU, info.Up)
func Get(bridge string) (Info, error) {
	members, err := Members(bridge)
	if err != nil {
		return Info{}, fmt.Errorf("get bridge %s info: %w", bridge, err)
	}

	iface, err := ifc.Get(bridge)
	if err != nil {
		return Info{}, fmt.Errorf("get bridge %s interface: %w", bridge, err)
	}

	return Info{
		Name:    bridge,
		Members: members,
		MTU:     iface.MTU,
		Up:      iface.Flags.IsUp(),
	}, nil
}

// MemberFlags returns the flags of a bridge member.
//
// Returns ErrNotFound if the bridge does not exist or the interface is not a
// member of it.
//
// Example:
//
//	flags, err := bridge.MemberFlags("bridge0", "tap0")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("tap0: %s (learning=%v)\n", flags, flags.Has(bridge.FlagLearning))
func MemberFlags(bridge, member string) (MemberFlag, error) {
	f, err := bridgeops.GetMemberFlags(bridge, member)
	if err != nil {
		return 0, fmt.Errorf("get flags of member %s of bridge %s: %w", member, bridge, err)
	}
	return MemberFlag(f), nil
}

// SetMemberFlag sets or clears flag on a bridge member and leaves the
// member's other flags unchanged.
//
// The flags are read, modified and written back, so the change is not atomic
// against another process changing the same member. flag may combine several
// bits. FlagSpan, FlagBSTPAdmEdge and FlagBSTPAdmCost cannot be changed this
// way (the kernel refuses SPAN and ignores the other two), so they are
// rejected with ErrInvalidArgument. Requires root privileges.
//
// Example:
//
//	// Stop a member from learning source addresses
//	if err := bridge.SetMemberFlag("bridge0", "tap0", bridge.FlagLearning, false); err != nil {
//		log.Fatal(err)
//	}
func SetMemberFlag(bridge, member string, flag MemberFlag, set bool) error {
	if flag == 0 {
		return fmt.Errorf("set flags of member %s of bridge %s: no flag given", member, bridge)
	}
	if err := bridgeops.ModifyMemberFlag(bridge, member, uint32(flag), set); err != nil {
		return fmt.Errorf("set flags of member %s of bridge %s: %w", member, bridge, err)
	}
	return nil
}

// SetMemberPrivate marks a bridge member as a private segment or clears the
// mark.
//
// Private members do not exchange traffic with each other, only with
// non-private members; this isolates VMs that share a bridge. It is the
// equivalent of "ifconfig bridge0 private tap0" / "-private tap0". The
// member's other flags are preserved. Requires root privileges.
//
// Example:
//
//	// Isolate tap0 and tap1 from each other, both still reach the uplink
//	for _, tap := range []string{"tap0", "tap1"} {
//		if err := bridge.SetMemberPrivate("bridge0", tap, true); err != nil {
//			log.Fatal(err)
//		}
//	}
func SetMemberPrivate(bridge, member string, private bool) error {
	return SetMemberFlag(bridge, member, FlagPrivate, private)
}

// MemberPrivate reports whether a bridge member is marked private.
//
// Example:
//
//	private, err := bridge.MemberPrivate("bridge0", "tap0")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("tap0 private: %v\n", private)
func MemberPrivate(bridge, member string) (bool, error) {
	f, err := MemberFlags(bridge, member)
	if err != nil {
		return false, err
	}
	return f.Has(FlagPrivate), nil
}
