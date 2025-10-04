//go:build freebsd
// +build freebsd

package ip

import (
	"fmt"
	"net"

	"github.com/zombocoder/go-freebsd-ifc/internal/ipaddr"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Add4 adds an IPv4 address to an interface.
//
// Returns a validation error if the IP is not IPv4 or the mask is invalid.
// This operation is idempotent - returns nil if the address already exists.
func Add4(iface string, ip net.IP, mask net.IPMask) error {
	if ip.To4() == nil {
		return isyscall.NewValidationError("ip", ip.String(), "not an IPv4 address")
	}
	if len(mask) != 4 {
		return isyscall.NewValidationError("mask", fmt.Sprintf("%v", mask), "invalid IPv4 mask length")
	}
	if err := ipaddr.Add4(iface, ip.To4(), mask); err != nil {
		return fmt.Errorf("add IPv4 %s/%s to %s: %w", ip, mask, iface, err)
	}
	return nil
}

// Del4 removes an IPv4 address from an interface.
//
// Returns a validation error if the IP is not IPv4 or the mask is invalid.
// This operation is idempotent - returns nil if the address doesn't exist.
func Del4(iface string, ip net.IP, mask net.IPMask) error {
	if ip.To4() == nil {
		return isyscall.NewValidationError("ip", ip.String(), "not an IPv4 address")
	}
	if len(mask) != 4 {
		return isyscall.NewValidationError("mask", fmt.Sprintf("%v", mask), "invalid IPv4 mask length")
	}
	if err := ipaddr.Del4(iface, ip.To4(), mask); err != nil {
		return fmt.Errorf("delete IPv4 %s/%s from %s: %w", ip, mask, iface, err)
	}
	return nil
}

// Add6 adds an IPv6 address to an interface.
//
// Returns a validation error if the IP is not IPv6 or prefixLen is invalid.
// This operation is idempotent - returns nil if the address already exists.
func Add6(iface string, ip net.IP, prefixLen int) error {
	if ip.To4() != nil {
		return isyscall.NewValidationError("ip", ip.String(), "not an IPv6 address")
	}
	if prefixLen < 0 || prefixLen > 128 {
		return isyscall.NewValidationError("prefixLen", fmt.Sprintf("%d", prefixLen), "must be between 0 and 128")
	}
	if err := ipaddr.Add6(iface, ip, prefixLen); err != nil {
		return fmt.Errorf("add IPv6 %s/%d to %s: %w", ip, prefixLen, iface, err)
	}
	return nil
}

// Del6 removes an IPv6 address from an interface.
//
// Returns a validation error if the IP is not IPv6 or prefixLen is invalid.
// This operation is idempotent - returns nil if the address doesn't exist.
func Del6(iface string, ip net.IP, prefixLen int) error {
	if ip.To4() != nil {
		return isyscall.NewValidationError("ip", ip.String(), "not an IPv6 address")
	}
	if prefixLen < 0 || prefixLen > 128 {
		return isyscall.NewValidationError("prefixLen", fmt.Sprintf("%d", prefixLen), "must be between 0 and 128")
	}
	if err := ipaddr.Del6(iface, ip, prefixLen); err != nil {
		return fmt.Errorf("delete IPv6 %s/%d from %s: %w", ip, prefixLen, iface, err)
	}
	return nil
}
