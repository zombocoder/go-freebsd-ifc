//go:build freebsd
// +build freebsd

package route

import (
	"fmt"
	"net"
	"net/netip"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	"github.com/zombocoder/go-freebsd-ifc/internal/routing"
)

// AddDefault4 adds an IPv4 default route
func AddDefault4(iface string, gw net.IP) error {
	if gw.To4() == nil {
		return fmt.Errorf("not an IPv4 address: %v", gw)
	}
	_, defaultNet, _ := net.ParseCIDR("0.0.0.0/0")

	ifindex := 0
	if iface != "" {
		ifc, err := ifc.Get(iface)
		if err != nil {
			return err
		}
		ifindex = ifc.Index
	}

	return routing.ModifyRoute(true, defaultNet, gw.To4(), ifindex)
}

// DelDefault4 deletes an IPv4 default route
func DelDefault4(iface string, gw net.IP) error {
	if gw.To4() == nil {
		return fmt.Errorf("not an IPv4 address: %v", gw)
	}
	_, defaultNet, _ := net.ParseCIDR("0.0.0.0/0")

	ifindex := 0
	if iface != "" {
		ifc, err := ifc.Get(iface)
		if err != nil {
			return err
		}
		ifindex = ifc.Index
	}

	return routing.ModifyRoute(false, defaultNet, gw.To4(), ifindex)
}

// AddRoute4 adds an IPv4 route
func AddRoute4(dst *net.IPNet, gw net.IP, iface string) error {
	if dst == nil {
		return fmt.Errorf("destination network is required")
	}
	if dst.IP.To4() == nil {
		return fmt.Errorf("not an IPv4 network: %v", dst)
	}
	if gw.To4() == nil {
		return fmt.Errorf("not an IPv4 address: %v", gw)
	}

	ifindex := 0
	if iface != "" {
		ifc, err := ifc.Get(iface)
		if err != nil {
			return err
		}
		ifindex = ifc.Index
	}

	return routing.ModifyRoute(true, dst, gw.To4(), ifindex)
}

// DelRoute4 deletes an IPv4 route
func DelRoute4(dst *net.IPNet, gw net.IP, iface string) error {
	if dst == nil {
		return fmt.Errorf("destination network is required")
	}
	if dst.IP.To4() == nil {
		return fmt.Errorf("not an IPv4 network: %v", dst)
	}
	if gw.To4() == nil {
		return fmt.Errorf("not an IPv4 address: %v", gw)
	}

	ifindex := 0
	if iface != "" {
		ifc, err := ifc.Get(iface)
		if err != nil {
			return err
		}
		ifindex = ifc.Index
	}

	return routing.ModifyRoute(false, dst, gw.To4(), ifindex)
}

// AddDefault6 adds an IPv6 default route
func AddDefault6(iface string, gw net.IP) error {
	if gw.To4() != nil {
		return fmt.Errorf("not an IPv6 address: %v", gw)
	}
	_, defaultNet, _ := net.ParseCIDR("::/0")

	ifindex := 0
	if iface != "" {
		ifc, err := ifc.Get(iface)
		if err != nil {
			return err
		}
		ifindex = ifc.Index
	}

	return routing.ModifyRoute(true, defaultNet, gw, ifindex)
}

// DelDefault6 deletes an IPv6 default route
func DelDefault6(iface string, gw net.IP) error {
	if gw.To4() != nil {
		return fmt.Errorf("not an IPv6 address: %v", gw)
	}
	_, defaultNet, _ := net.ParseCIDR("::/0")

	ifindex := 0
	if iface != "" {
		ifc, err := ifc.Get(iface)
		if err != nil {
			return err
		}
		ifindex = ifc.Index
	}

	return routing.ModifyRoute(false, defaultNet, gw, ifindex)
}

// AddRoute6 adds an IPv6 route
func AddRoute6(dst *net.IPNet, gw net.IP, iface string) error {
	if dst == nil {
		return fmt.Errorf("destination network is required")
	}
	if dst.IP.To4() != nil {
		return fmt.Errorf("not an IPv6 network: %v", dst)
	}
	if gw.To4() != nil {
		return fmt.Errorf("not an IPv6 address: %v", gw)
	}

	ifindex := 0
	if iface != "" {
		ifc, err := ifc.Get(iface)
		if err != nil {
			return err
		}
		ifindex = ifc.Index
	}

	return routing.ModifyRoute(true, dst, gw, ifindex)
}

// DelRoute6 deletes an IPv6 route
func DelRoute6(dst *net.IPNet, gw net.IP, iface string) error {
	if dst == nil {
		return fmt.Errorf("destination network is required")
	}
	if dst.IP.To4() != nil {
		return fmt.Errorf("not an IPv6 network: %v", dst)
	}
	if gw.To4() != nil {
		return fmt.Errorf("not an IPv6 address: %v", gw)
	}

	ifindex := 0
	if iface != "" {
		ifc, err := ifc.Get(iface)
		if err != nil {
			return err
		}
		ifindex = ifc.Index
	}

	return routing.ModifyRoute(false, dst, gw, ifindex)
}

// Flags is the raw RTF_* bitmask the kernel keeps for a route.
type Flags uint32

// IsUp reports whether the route is usable (RTF_UP).
func (f Flags) IsUp() bool { return f&Flags(constants.RTF_UP) != 0 }

// IsGateway reports whether the route forwards to a next hop (RTF_GATEWAY).
func (f Flags) IsGateway() bool { return f&Flags(constants.RTF_GATEWAY) != 0 }

// IsHost reports whether the route is a host route (RTF_HOST).
func (f Flags) IsHost() bool { return f&Flags(constants.RTF_HOST) != 0 }

// IsStatic reports whether the route was added manually (RTF_STATIC).
func (f Flags) IsStatic() bool { return f&Flags(constants.RTF_STATIC) != 0 }

// Family is the address family of a route.
type Family int

const (
	// FamilyIPv4 is an AF_INET route.
	FamilyIPv4 Family = constants.AF_INET
	// FamilyIPv6 is an AF_INET6 route.
	FamilyIPv6 Family = constants.AF_INET6
)

// String returns "inet", "inet6" or "unknown".
func (f Family) String() string {
	switch f {
	case FamilyIPv4:
		return "inet"
	case FamilyIPv6:
		return "inet6"
	}
	return "unknown"
}

// Route describes one entry of the kernel routing table.
type Route struct {
	// Family is the address family the kernel reported in the route's
	// destination sockaddr.
	//
	// It is carried explicitly rather than inferred from Dst, because Go's
	// net.IP deliberately conflates the two families for IPv4-mapped
	// addresses. FreeBSD's IPv6 table really does contain a
	// ::ffff:0.0.0.0/96 route; for that value net.IP.To4() returns non-nil
	// and net.IPNet.String() prints "0.0.0.0/0". Test this field, never the
	// shape of Dst, to decide which family a route belongs to.
	Family Family

	// Dst is the destination prefix. Host routes carry an all-ones mask.
	// The address and mask keep the width of their family: 4 bytes for
	// FamilyIPv4 and 16 for FamilyIPv6, never unmapped.
	Dst *net.IPNet

	// Gateway is the next hop, or nil for a directly attached route whose
	// gateway the kernel reports as a link-layer address.
	Gateway net.IP

	// Index is the index of the outgoing interface, 0 when unknown.
	Index int

	// Iface is the name of the outgoing interface, "" when unknown.
	Iface string

	// Flags is the RTF_* bitmask reported by the kernel.
	Flags Flags
}

// Is4 reports whether the route is an IPv4 route.
func (r Route) Is4() bool { return r.Family == FamilyIPv4 }

// Is6 reports whether the route is an IPv6 route. This is true for an
// IPv4-mapped destination such as ::ffff:0.0.0.0/96, which is an IPv6 route
// however much it looks like an IPv4 one.
func (r Route) Is6() bool { return r.Family == FamilyIPv6 }

// String renders the route roughly the way netstat -rn does.
//
// The prefix is formatted through net/netip rather than net.IPNet.String(),
// which would print the IPv6 route ::ffff:0.0.0.0/96 as "0.0.0.0/0":
// netip.Addr keeps Is4In6 distinct from Is4, so the family survives.
func (r Route) String() string {
	gw := "link"
	if r.Gateway != nil {
		gw = formatAddr(r.Gateway, r.Family)
	}
	return fmt.Sprintf("%s via %s dev %s flags 0x%x", r.prefixString(), gw, r.Iface, uint32(r.Flags))
}

// prefixString formats Dst without letting Go collapse a 4-in-6 address.
func (r Route) prefixString() string {
	if r.Dst == nil {
		return "<nil>"
	}
	addr, ok := netipAddr(r.Dst.IP, r.Family)
	if !ok {
		return r.Dst.String()
	}
	ones, _ := r.Dst.Mask.Size()
	return netip.PrefixFrom(addr, ones).String()
}

// formatAddr renders an address in the notation of the given family.
func formatAddr(ip net.IP, family Family) string {
	addr, ok := netipAddr(ip, family)
	if !ok {
		return ip.String()
	}
	return addr.String()
}

// netipAddr converts a net.IP to a netip.Addr of the width the family calls
// for, so an IPv4-mapped IPv6 address stays a 16-byte Is4In6 address instead
// of being unmapped to an Is4 one.
func netipAddr(ip net.IP, family Family) (netip.Addr, bool) {
	switch family {
	case FamilyIPv4:
		if v4 := ip.To4(); v4 != nil {
			return netip.AddrFrom4([4]byte(v4)), true
		}
	case FamilyIPv6:
		if len(ip) == net.IPv6len {
			return netip.AddrFrom16([16]byte(ip)), true
		}
		if v16 := ip.To16(); v16 != nil {
			return netip.AddrFrom16([16]byte(v16)), true
		}
	}
	return netip.Addr{}, false
}

func convertRoutes(in []routing.Route) []Route {
	out := make([]Route, len(in))
	for i, r := range in {
		out[i] = Route{
			Family:  Family(r.Family),
			Dst:     r.Dst,
			Gateway: r.Gateway,
			Index:   r.Index,
			Iface:   r.Iface,
			Flags:   Flags(r.Flags),
		}
	}
	return out
}

// List returns the whole kernel routing table, both IPv4 and IPv6.
//
// The table is read with the same CTL_NET/PF_ROUTE/NET_RT_DUMP sysctl that
// netstat(1) uses, so no privileges are required. It is the read side needed
// to reconcile a desired set of routes against what the kernel actually has.
//
// Example:
//
//	routes, err := route.List()
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, r := range routes {
//		fmt.Println(r)
//	}
func List() ([]Route, error) {
	rs, err := routing.List(constants.AF_UNSPEC)
	if err != nil {
		return nil, fmt.Errorf("list routes: %w", err)
	}
	return convertRoutes(rs), nil
}

// List4 returns the IPv4 routing table.
func List4() ([]Route, error) {
	rs, err := routing.List(constants.AF_INET)
	if err != nil {
		return nil, fmt.Errorf("list IPv4 routes: %w", err)
	}
	return convertRoutes(rs), nil
}

// List6 returns the IPv6 routing table.
func List6() ([]Route, error) {
	rs, err := routing.List(constants.AF_INET6)
	if err != nil {
		return nil, fmt.Errorf("list IPv6 routes: %w", err)
	}
	return convertRoutes(rs), nil
}
