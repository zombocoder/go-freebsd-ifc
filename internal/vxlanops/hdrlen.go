//go:build freebsd
// +build freebsd

package vxlanops

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <ifaddrs.h>
*/
import "C"
import (
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// HeaderLen returns if_hdrlen for an interface, i.e. the number of bytes of
// media header the kernel accounts for on every frame.
//
// For a vxlan interface this is the encapsulation overhead as the kernel
// itself computes it in vxlan_setup_interface_hdrlen() (sys/net/if_vxlan.c):
//
//	ifp->if_hdrlen = ETHER_HDR_LEN + sizeof(struct vxlanudphdr);
//	if (VXLAN_SOCKADDR_IS_IPV4(&sc->vxl_dst_addr))
//		ifp->if_hdrlen += sizeof(struct ip);
//	else if (VXLAN_SOCKADDR_IS_IPV6(&sc->vxl_dst_addr))
//		ifp->if_hdrlen += sizeof(struct ip6_hdr);
//	if ((sc->vxl_flags & VXLAN_FLAG_USER_MTU) == 0)
//		ifp->if_mtu = ETHERMTU - ifp->if_hdrlen;
//
// The value is read from if_data.ifi_hdrlen via getifaddrs(3), so it is the
// kernel's own number for that specific interface rather than arithmetic done
// here.
func HeaderLen(name string) (int, error) {
	var ifap *C.struct_ifaddrs
	if C.getifaddrs(&ifap) != 0 {
		return 0, isyscall.MapError(isyscall.GetErrno())
	}
	defer C.freeifaddrs(ifap)

	for ifa := ifap; ifa != nil; ifa = ifa.ifa_next {
		if C.GoString(ifa.ifa_name) != name {
			continue
		}
		if ifa.ifa_addr == nil || ifa.ifa_addr.sa_family != C.AF_LINK {
			continue
		}
		if ifa.ifa_data == nil {
			continue
		}
		ifdata := (*C.struct_if_data)(ifa.ifa_data)
		return int(ifdata.ifi_hdrlen), nil
	}

	return 0, isyscall.ErrNotFound
}
