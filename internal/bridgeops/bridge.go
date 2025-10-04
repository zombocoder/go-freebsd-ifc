//go:build freebsd
// +build freebsd

package bridgeops

/*
#include <sys/types.h>
#include <net/if.h>
#include <net/if_bridgevar.h>
#include <string.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// AddMember adds a member interface to a bridge
func AddMember(bridge, member string) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var req C.struct_ifbreq

	if len(bridge) >= constants.IFNAMSIZ || len(member) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long")
	}

	isyscall.CopyString(unsafe.Pointer(&req.ifbr_ifsname[0]), member, constants.IFNAMSIZ)

	// Allocate C memory for ifdrv to avoid Go pointer issues
	ifd := (*C.struct_ifdrv)(C.malloc(C.size_t(unsafe.Sizeof(C.struct_ifdrv{}))))
	defer C.free(unsafe.Pointer(ifd))

	isyscall.CopyString(unsafe.Pointer(&ifd.ifd_name[0]), bridge, constants.IFNAMSIZ)
	ifd.ifd_cmd = C.ulong(constants.BRDGADD)
	ifd.ifd_len = C.size_t(unsafe.Sizeof(req))
	ifd.ifd_data = unsafe.Pointer(&req)

	err = isyscall.Ioctl(s.Int(), constants.SIOCSDRVSPEC, unsafe.Pointer(ifd))
	if err != nil && err == isyscall.ErrExists {
		return nil // Idempotent
	}
	return err
}

// DelMember removes a member interface from a bridge
func DelMember(bridge, member string) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var req C.struct_ifbreq

	if len(bridge) >= constants.IFNAMSIZ || len(member) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long")
	}

	isyscall.CopyString(unsafe.Pointer(&req.ifbr_ifsname[0]), member, constants.IFNAMSIZ)

	// Allocate C memory for ifdrv to avoid Go pointer issues
	ifd := (*C.struct_ifdrv)(C.malloc(C.size_t(unsafe.Sizeof(C.struct_ifdrv{}))))
	defer C.free(unsafe.Pointer(ifd))

	isyscall.CopyString(unsafe.Pointer(&ifd.ifd_name[0]), bridge, constants.IFNAMSIZ)
	ifd.ifd_cmd = C.ulong(constants.BRDGDEL)
	ifd.ifd_len = C.size_t(unsafe.Sizeof(req))
	ifd.ifd_data = unsafe.Pointer(&req)

	err = isyscall.Ioctl(s.Int(), constants.SIOCSDRVSPEC, unsafe.Pointer(ifd))
	if err != nil && err == isyscall.ErrNotFound {
		return nil // Idempotent
	}
	return err
}

// GetMembers returns the list of member interfaces in a bridge
func GetMembers(bridge string) ([]string, error) {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	bufSize := 32
	reqSize := int(unsafe.Sizeof(C.struct_ifbreq{}))
	bufLen := bufSize * reqSize
	buf := make([]byte, bufLen)

	type ifbifconf struct {
		len uint32
		buf uintptr
	}

	bifc := ifbifconf{
		len: uint32(bufLen),
		buf: uintptr(unsafe.Pointer(&buf[0])),
	}

	if len(bridge) >= constants.IFNAMSIZ {
		return nil, fmt.Errorf("interface name too long")
	}

	// Allocate C memory for ifdrv to avoid Go pointer issues
	ifd := (*C.struct_ifdrv)(C.malloc(C.size_t(unsafe.Sizeof(C.struct_ifdrv{}))))
	defer C.free(unsafe.Pointer(ifd))

	isyscall.CopyString(unsafe.Pointer(&ifd.ifd_name[0]), bridge, constants.IFNAMSIZ)
	ifd.ifd_cmd = C.ulong(constants.BRDGGIFS)
	ifd.ifd_len = C.size_t(unsafe.Sizeof(bifc))
	ifd.ifd_data = unsafe.Pointer(&bifc)

	if err := isyscall.Ioctl(s.Int(), constants.SIOCGDRVSPEC, unsafe.Pointer(ifd)); err != nil {
		return nil, err
	}

	numMembers := int(bifc.len) / reqSize
	members := make([]string, 0, numMembers)

	for i := 0; i < numMembers; i++ {
		req := (*C.struct_ifbreq)(unsafe.Pointer(&buf[i*reqSize]))
		name := C.GoString(&req.ifbr_ifsname[0])
		members = append(members, name)
	}

	return members, nil
}
