//go:build freebsd
// +build freebsd

package vlanops

/*
#include <sys/types.h>
#include <net/if.h>
#include <net/if_vlan_var.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// VLANConfig represents VLAN configuration
type VLANConfig struct {
	Tag    uint16
	Parent string
}

// Configure sets the VLAN tag and parent interface
func Configure(name string, tag uint16, parent string) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var vlanreq C.struct_vlanreq

	if len(name) >= constants.IFNAMSIZ || len(parent) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("interface name", name, "name too long")
	}

	if tag == 0 || tag > 4094 {
		return isyscall.NewValidationError("tag", fmt.Sprintf("%d", tag), "must be between 1 and 4094")
	}

	isyscall.CopyString(unsafe.Pointer(&vlanreq.vlr_parent[0]), parent, constants.IFNAMSIZ)
	vlanreq.vlr_tag = C.ushort(tag)

	var ifr C.struct_ifreq
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)
	*(*C.caddr_t)(unsafe.Pointer(&ifr.ifr_ifru)) = C.caddr_t(unsafe.Pointer(&vlanreq))

	if err := isyscall.Ioctl(s.Int(), constants.SIOCSETVLAN, unsafe.Pointer(&ifr)); err != nil {
		return err
	}

	return nil
}

// Get returns VLAN configuration
func Get(name string) (VLANConfig, error) {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return VLANConfig{}, err
	}
	defer s.Close()

	var vlanreq C.struct_vlanreq

	if len(name) >= constants.IFNAMSIZ {
		return VLANConfig{}, isyscall.NewValidationError("name", name, "interface name too long")
	}

	var ifr C.struct_ifreq
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)
	*(*C.caddr_t)(unsafe.Pointer(&ifr.ifr_ifru)) = C.caddr_t(unsafe.Pointer(&vlanreq))

	if err := isyscall.Ioctl(s.Int(), constants.SIOCGETVLAN, unsafe.Pointer(&ifr)); err != nil {
		return VLANConfig{}, err
	}

	tag := uint16(vlanreq.vlr_tag)
	parent := C.GoString(&vlanreq.vlr_parent[0])

	return VLANConfig{
		Tag:    tag,
		Parent: parent,
	}, nil
}
