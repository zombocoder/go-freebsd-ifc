//go:build freebsd
// +build freebsd

package laggops

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <net/if.h>
#include <net/ethernet.h>
#include <net/if_lagg.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Kernel protocol numbers (enum lagg_proto in net/if_lagg.h).
const (
	ProtoNone        = C.LAGG_PROTO_NONE
	ProtoRoundRobin  = C.LAGG_PROTO_ROUNDROBIN
	ProtoFailover    = C.LAGG_PROTO_FAILOVER
	ProtoLoadBalance = C.LAGG_PROTO_LOADBALANCE
	ProtoLACP        = C.LAGG_PROTO_LACP
	ProtoBroadcast   = C.LAGG_PROTO_BROADCAST
)

// maxPorts bounds the port list returned by Get.
const maxPorts = 64

// LAGGConfig represents LAGG configuration
type LAGGConfig struct {
	Proto int
	Ports []string
}

// The lagg ioctls take their request struct directly, with the lagg name
// in its first field (ra_ifname / rp_ifname), not wrapped in a struct
// ifreq. See setlaggproto() and setlaggport() in sbin/ifconfig/iflagg.c.

func checkName(name string) error {
	if name == "" {
		return isyscall.NewValidationError("name", "", "interface name is empty")
	}
	if len(name) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("name", name, "interface name too long")
	}
	return nil
}

// SetProto sets the LAGG protocol
func SetProto(name string, proto int) error {
	if err := checkName(name); err != nil {
		return err
	}

	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var ra C.struct_lagg_reqall
	isyscall.CopyString(unsafe.Pointer(&ra.ra_ifname[0]), name, constants.IFNAMSIZ)
	ra.ra_proto = C.u_int(proto)

	return isyscall.Ioctl(s.Int(), constants.SIOCSLAGG, unsafe.Pointer(&ra))
}

// AddPort adds a port to the LAGG interface
func AddPort(lagg, port string) error {
	return portIoctl(constants.SIOCSLAGGPORT, lagg, port)
}

// DelPort removes a port from the LAGG interface
func DelPort(lagg, port string) error {
	return portIoctl(constants.SIOCSLAGGDELPORT, lagg, port)
}

func portIoctl(req uintptr, lagg, port string) error {
	if len(lagg) >= constants.IFNAMSIZ || len(port) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("interface name", lagg, "name too long")
	}
	if lagg == "" || port == "" {
		return isyscall.NewValidationError("interface name", lagg, "name is empty")
	}

	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var rp C.struct_lagg_reqport
	isyscall.CopyString(unsafe.Pointer(&rp.rp_ifname[0]), lagg, constants.IFNAMSIZ)
	isyscall.CopyString(unsafe.Pointer(&rp.rp_portname[0]), port, constants.IFNAMSIZ)

	return isyscall.Ioctl(s.Int(), req, unsafe.Pointer(&rp))
}

// Get returns LAGG configuration
func Get(name string) (LAGGConfig, error) {
	if err := checkName(name); err != nil {
		return LAGGConfig{}, err
	}

	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return LAGGConfig{}, err
	}
	defer s.Close()

	// The kernel copies port entries out through ra_port. The buffer is C
	// memory: a Go pointer stored inside a struct passed to C would violate
	// the cgo pointer rules.
	size := C.size_t(maxPorts) * C.size_t(C.sizeof_struct_lagg_reqport)
	buf := (*C.struct_lagg_reqport)(C.calloc(1, size))
	if buf == nil {
		return LAGGConfig{}, isyscall.NewValidationError("ports", name, "cannot allocate port buffer")
	}
	defer C.free(unsafe.Pointer(buf))

	var ra C.struct_lagg_reqall
	isyscall.CopyString(unsafe.Pointer(&ra.ra_ifname[0]), name, constants.IFNAMSIZ)
	ra.ra_port = buf
	ra.ra_size = size

	if err := isyscall.Ioctl(s.Int(), constants.SIOCGLAGG, unsafe.Pointer(&ra)); err != nil {
		return LAGGConfig{}, err
	}

	ports := unsafe.Slice(buf, maxPorts)
	n := int(ra.ra_ports)
	if n > maxPorts {
		n = maxPorts
	}
	portNames := make([]string, 0, n)
	for i := 0; i < n; i++ {
		if p := C.GoString(&ports[i].rp_portname[0]); p != "" {
			portNames = append(portNames, p)
		}
	}

	return LAGGConfig{
		Proto: int(ra.ra_proto),
		Ports: portNames,
	}, nil
}
