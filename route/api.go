//go:build freebsd
// +build freebsd

package route

import (
	"fmt"
	"net"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
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
