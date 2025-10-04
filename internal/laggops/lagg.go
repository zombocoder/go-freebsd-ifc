//go:build freebsd
// +build freebsd

package laggops

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <net/if.h>
#include <net/ethernet.h>
#include <net/if_lagg.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// LAGGConfig represents LAGG configuration
type LAGGConfig struct {
	Proto int
	Ports []string
}

// SetProto sets the LAGG protocol
func SetProto(name string, proto int) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	if len(name) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("name", name, "interface name too long")
	}

	var laggreq C.struct_lagg_reqall
	var ifr C.struct_ifreq

	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)

	// Set protocol
	laggreq.ra_proto = C.u_int(proto)
	*(*C.caddr_t)(unsafe.Pointer(&ifr.ifr_ifru)) = C.caddr_t(unsafe.Pointer(&laggreq))

	if err := isyscall.Ioctl(s.Int(), constants.SIOCSLAGG, unsafe.Pointer(&ifr)); err != nil {
		return err
	}

	return nil
}

// AddPort adds a port to the LAGG interface
func AddPort(lagg, port string) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	if len(lagg) >= constants.IFNAMSIZ || len(port) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("interface name", lagg, "name too long")
	}

	var laggreq C.struct_lagg_reqport
	var ifr C.struct_ifreq

	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), lagg, constants.IFNAMSIZ)
	isyscall.CopyString(unsafe.Pointer(&laggreq.rp_portname[0]), port, constants.IFNAMSIZ)

	*(*C.caddr_t)(unsafe.Pointer(&ifr.ifr_ifru)) = C.caddr_t(unsafe.Pointer(&laggreq))

	if err := isyscall.Ioctl(s.Int(), constants.SIOCSLAGGPORT, unsafe.Pointer(&ifr)); err != nil {
		return err
	}

	return nil
}

// DelPort removes a port from the LAGG interface
func DelPort(lagg, port string) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	if len(lagg) >= constants.IFNAMSIZ || len(port) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("interface name", lagg, "name too long")
	}

	var laggreq C.struct_lagg_reqport
	var ifr C.struct_ifreq

	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), lagg, constants.IFNAMSIZ)
	isyscall.CopyString(unsafe.Pointer(&laggreq.rp_portname[0]), port, constants.IFNAMSIZ)

	*(*C.caddr_t)(unsafe.Pointer(&ifr.ifr_ifru)) = C.caddr_t(unsafe.Pointer(&laggreq))

	if err := isyscall.Ioctl(s.Int(), constants.SIOCSLAGGDELPORT, unsafe.Pointer(&ifr)); err != nil {
		return err
	}

	return nil
}

// Get returns LAGG configuration
func Get(name string) (LAGGConfig, error) {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return LAGGConfig{}, err
	}
	defer s.Close()

	if len(name) >= constants.IFNAMSIZ {
		return LAGGConfig{}, isyscall.NewValidationError("name", name, "interface name too long")
	}

	var laggreq C.struct_lagg_reqall
	var ifr C.struct_ifreq

	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)

	// Allocate space for ports
	const maxPorts = 32
	var ports [maxPorts]C.struct_lagg_reqport
	laggreq.ra_port = &ports[0]
	laggreq.ra_size = C.ulong(maxPorts * C.sizeof_struct_lagg_reqport)

	*(*C.caddr_t)(unsafe.Pointer(&ifr.ifr_ifru)) = C.caddr_t(unsafe.Pointer(&laggreq))

	if err := isyscall.Ioctl(s.Int(), constants.SIOCGLAGG, unsafe.Pointer(&ifr)); err != nil {
		return LAGGConfig{}, err
	}

	// Extract protocol
	proto := int(laggreq.ra_proto)

	// Extract port names
	numPorts := int(laggreq.ra_ports)
	portNames := make([]string, 0, numPorts)
	for i := 0; i < numPorts && i < maxPorts; i++ {
		portName := C.GoString(&ports[i].rp_portname[0])
		if portName != "" {
			portNames = append(portNames, portName)
		}
	}

	return LAGGConfig{
		Proto: proto,
		Ports: portNames,
	}, nil
}
