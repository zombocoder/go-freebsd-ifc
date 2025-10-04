//go:build freebsd
// +build freebsd

package ipaddr

/*
#include <sys/types.h>
#include <net/if.h>
#include <netinet/in.h>
#include <netinet6/in6_var.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"net"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Add4 adds an IPv4 address to an interface
func Add4(iface string, ip net.IP, mask net.IPMask) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var req C.struct_ifaliasreq
	if len(iface) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long: %s", iface)
	}
	isyscall.CopyString(unsafe.Pointer(&req.ifra_name[0]), iface, constants.IFNAMSIZ)

	// Set address
	addr := (*C.struct_sockaddr_in)(unsafe.Pointer(&req.ifra_addr))
	addr.sin_family = constants.AF_INET
	addr.sin_len = constants.SizeofSockaddrIn
	isyscall.CopyBytes(unsafe.Pointer(&addr.sin_addr), unsafe.Pointer(&ip[0]), 4)

	// Set netmask
	netmask := (*C.struct_sockaddr_in)(unsafe.Pointer(&req.ifra_mask))
	netmask.sin_family = constants.AF_INET
	netmask.sin_len = constants.SizeofSockaddrIn
	isyscall.CopyBytes(unsafe.Pointer(&netmask.sin_addr), unsafe.Pointer(&mask[0]), 4)

	// Calculate broadcast address
	bcast := (*C.struct_sockaddr_in)(unsafe.Pointer(&req.ifra_broadaddr))
	bcast.sin_family = constants.AF_INET
	bcast.sin_len = constants.SizeofSockaddrIn
	bcastIP := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		bcastIP[i] = ip[i] | ^mask[i]
	}
	isyscall.CopyBytes(unsafe.Pointer(&bcast.sin_addr), unsafe.Pointer(&bcastIP[0]), 4)

	err = isyscall.Ioctl(s.Int(), constants.SIOCAIFADDR, unsafe.Pointer(&req))
	if err != nil && err == isyscall.ErrExists {
		return nil // Idempotent
	}
	return err
}

// Del4 removes an IPv4 address from an interface
func Del4(iface string, ip net.IP, mask net.IPMask) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var req C.struct_ifreq
	if len(iface) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long: %s", iface)
	}
	isyscall.CopyString(unsafe.Pointer(&req.ifr_name[0]), iface, constants.IFNAMSIZ)

	addr := (*C.struct_sockaddr_in)(unsafe.Pointer(&req.ifr_ifru))
	addr.sin_family = constants.AF_INET
	addr.sin_len = constants.SizeofSockaddrIn
	isyscall.CopyBytes(unsafe.Pointer(&addr.sin_addr), unsafe.Pointer(&ip[0]), 4)

	err = isyscall.Ioctl(s.Int(), constants.SIOCDIFADDR, unsafe.Pointer(&req))
	if err != nil && err == isyscall.ErrNotFound {
		return nil // Idempotent
	}
	return err
}

// Add6 adds an IPv6 address to an interface
func Add6(iface string, ip net.IP, prefixLen int) error {
	s, err := isyscall.CreateInet6Socket()
	if err != nil {
		return err
	}
	defer s.Close()

	var req C.struct_in6_aliasreq
	if len(iface) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long: %s", iface)
	}
	isyscall.CopyString(unsafe.Pointer(&req.ifra_name[0]), iface, constants.IFNAMSIZ)

	// Set address
	addr := (*C.struct_sockaddr_in6)(unsafe.Pointer(&req.ifra_addr))
	addr.sin6_family = constants.AF_INET6
	addr.sin6_len = constants.SizeofSockaddrIn6
	isyscall.CopyBytes(unsafe.Pointer(&addr.sin6_addr), unsafe.Pointer(&ip[0]), 16)

	// Set prefix mask
	mask := (*C.struct_sockaddr_in6)(unsafe.Pointer(&req.ifra_prefixmask))
	mask.sin6_family = constants.AF_INET6
	mask.sin6_len = constants.SizeofSockaddrIn6
	prefixMask := net.CIDRMask(prefixLen, 128)
	isyscall.CopyBytes(unsafe.Pointer(&mask.sin6_addr), unsafe.Pointer(&prefixMask[0]), 16)

	// Set lifetime
	req.ifra_lifetime.ia6t_vltime = constants.ND6_INFINITE_LIFETIME
	req.ifra_lifetime.ia6t_pltime = constants.ND6_INFINITE_LIFETIME

	err = isyscall.Ioctl(s.Int(), constants.SIOCAIFADDR_IN6, unsafe.Pointer(&req))
	if err != nil && err == isyscall.ErrExists {
		return nil // Idempotent
	}
	return err
}

// Del6 removes an IPv6 address from an interface
func Del6(iface string, ip net.IP, prefixLen int) error {
	s, err := isyscall.CreateInet6Socket()
	if err != nil {
		return err
	}
	defer s.Close()

	var req C.struct_in6_ifreq
	if len(iface) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long: %s", iface)
	}
	isyscall.CopyString(unsafe.Pointer(&req.ifr_name[0]), iface, constants.IFNAMSIZ)

	addr := (*C.struct_sockaddr_in6)(unsafe.Pointer(&req.ifr_ifru))
	addr.sin6_family = constants.AF_INET6
	addr.sin6_len = constants.SizeofSockaddrIn6
	isyscall.CopyBytes(unsafe.Pointer(&addr.sin6_addr), unsafe.Pointer(&ip[0]), 16)

	err = isyscall.Ioctl(s.Int(), constants.SIOCDIFADDR_IN6, unsafe.Pointer(&req))
	if err != nil && err == isyscall.ErrNotFound {
		return nil // Idempotent
	}
	return err
}
