//go:build freebsd
// +build freebsd

package ifops

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <ifaddrs.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <netinet/in.h>
#include <string.h>
*/
import "C"
import (
	"net"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Interface represents internal interface data
type Interface struct {
	Name  string
	Index int
	MTU   int
	Flags uint32
	Addrs []net.Addr
}

// List returns all network interfaces
func List() ([]Interface, error) {
	var ifap *C.struct_ifaddrs
	if C.getifaddrs(&ifap) != 0 {
		return nil, isyscall.MapError(isyscall.GetErrno())
	}
	defer C.freeifaddrs(ifap)

	ifaceMap := make(map[string]*Interface)

	for ifa := ifap; ifa != nil; ifa = ifa.ifa_next {
		name := C.GoString(ifa.ifa_name)

		iface, exists := ifaceMap[name]
		if !exists {
			iface = &Interface{
				Name:  name,
				Flags: uint32(ifa.ifa_flags),
				Addrs: []net.Addr{},
			}
			ifaceMap[name] = iface
		}

		if ifa.ifa_addr != nil {
			family := ifa.ifa_addr.sa_family

			switch family {
			case constants.AF_LINK:
				sdl := (*C.struct_sockaddr_dl)(unsafe.Pointer(ifa.ifa_addr))
				iface.Index = int(sdl.sdl_index)

			case constants.AF_INET:
				sin := (*C.struct_sockaddr_in)(unsafe.Pointer(ifa.ifa_addr))
				ip := make(net.IP, 4)
				isyscall.CopyBytes(unsafe.Pointer(&ip[0]), unsafe.Pointer(&sin.sin_addr), 4)

				var mask net.IPMask
				if ifa.ifa_netmask != nil {
					sinMask := (*C.struct_sockaddr_in)(unsafe.Pointer(ifa.ifa_netmask))
					mask = make(net.IPMask, 4)
					isyscall.CopyBytes(unsafe.Pointer(&mask[0]), unsafe.Pointer(&sinMask.sin_addr), 4)
				}

				iface.Addrs = append(iface.Addrs, &net.IPNet{IP: ip, Mask: mask})

			case constants.AF_INET6:
				sin6 := (*C.struct_sockaddr_in6)(unsafe.Pointer(ifa.ifa_addr))
				ip := make(net.IP, 16)
				isyscall.CopyBytes(unsafe.Pointer(&ip[0]), unsafe.Pointer(&sin6.sin6_addr), 16)

				var mask net.IPMask
				if ifa.ifa_netmask != nil {
					sin6Mask := (*C.struct_sockaddr_in6)(unsafe.Pointer(ifa.ifa_netmask))
					mask = make(net.IPMask, 16)
					isyscall.CopyBytes(unsafe.Pointer(&mask[0]), unsafe.Pointer(&sin6Mask.sin6_addr), 16)
				}

				iface.Addrs = append(iface.Addrs, &net.IPNet{IP: ip, Mask: mask})
			}
		}
	}

	// Get MTU for each interface
	for name, iface := range ifaceMap {
		mtu, err := GetMTU(name)
		if err == nil {
			iface.MTU = mtu
		}
	}

	result := make([]Interface, 0, len(ifaceMap))
	for _, iface := range ifaceMap {
		result = append(result, *iface)
	}

	return result, nil
}
