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

// drvspec issues a bridge ioctl (SIOCSDRVSPEC / SIOCGDRVSPEC) with cmd and
// the C buffer data of length n. Both the ifdrv and the buffer it points to
// are C memory: a Go pointer stored in C memory would break the cgo pointer
// rules. The ifdrv is zeroed because the name copy writes no NUL.
func drvspec(bridge string, ioc uintptr, cmd uintptr, data unsafe.Pointer, n uintptr) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	ifd := (*C.struct_ifdrv)(C.calloc(1, C.size_t(unsafe.Sizeof(C.struct_ifdrv{}))))
	defer C.free(unsafe.Pointer(ifd))

	isyscall.CopyString(unsafe.Pointer(&ifd.ifd_name[0]), bridge, constants.IFNAMSIZ)
	ifd.ifd_cmd = C.ulong(cmd)
	ifd.ifd_len = C.size_t(n)
	ifd.ifd_data = data

	return isyscall.Ioctl(s.Int(), ioc, unsafe.Pointer(ifd))
}

// newReq allocates a zeroed ifbreq in C memory naming member.
func newReq(member string) *C.struct_ifbreq {
	req := (*C.struct_ifbreq)(C.calloc(1, C.size_t(unsafe.Sizeof(C.struct_ifbreq{}))))
	isyscall.CopyString(unsafe.Pointer(&req.ifbr_ifsname[0]), member, constants.IFNAMSIZ)
	return req
}

// AddMember adds a member interface to a bridge
func AddMember(bridge, member string) error {
	if err := checkMemberNames(bridge, member); err != nil {
		return err
	}
	req := newReq(member)
	defer C.free(unsafe.Pointer(req))

	err := drvspec(bridge, constants.SIOCSDRVSPEC, constants.BRDGADD, unsafe.Pointer(req), unsafe.Sizeof(*req))
	if err != nil && err == isyscall.ErrExists {
		return nil // Idempotent
	}
	return err
}

// DelMember removes a member interface from a bridge
func DelMember(bridge, member string) error {
	if err := checkMemberNames(bridge, member); err != nil {
		return err
	}
	req := newReq(member)
	defer C.free(unsafe.Pointer(req))

	err := drvspec(bridge, constants.SIOCSDRVSPEC, constants.BRDGDEL, unsafe.Pointer(req), unsafe.Sizeof(*req))
	if err != nil && err == isyscall.ErrNotFound {
		return nil // Idempotent
	}
	return err
}

// GetMembers returns the list of member interfaces in a bridge
func GetMembers(bridge string) ([]string, error) {
	if err := checkBridgeName(bridge); err != nil {
		return nil, err
	}

	bufSize := 32
	reqSize := int(unsafe.Sizeof(C.struct_ifbreq{}))
	bufLen := bufSize * reqSize

	// The kernel copies the member list out through ifbic_buf, so both the
	// list buffer and the ifbifconf that points to it are C memory.
	buf := C.calloc(1, C.size_t(bufLen))
	if buf == nil {
		return nil, fmt.Errorf("allocate member buffer for %s", bridge)
	}
	defer C.free(buf)

	type ifbifconf struct {
		len uint32
		buf uintptr
	}
	bifc := (*ifbifconf)(C.calloc(1, C.size_t(unsafe.Sizeof(ifbifconf{}))))
	defer C.free(unsafe.Pointer(bifc))
	bifc.len = uint32(bufLen)
	bifc.buf = uintptr(buf)

	if err := drvspec(bridge, constants.SIOCGDRVSPEC, constants.BRDGGIFS, unsafe.Pointer(bifc), unsafe.Sizeof(*bifc)); err != nil {
		return nil, err
	}

	numMembers := int(bifc.len) / reqSize
	members := make([]string, 0, numMembers)

	for i := 0; i < numMembers; i++ {
		req := (*C.struct_ifbreq)(unsafe.Add(buf, i*reqSize))
		name := C.GoString(&req.ifbr_ifsname[0])
		members = append(members, name)
	}

	return members, nil
}

func checkBridgeName(bridge string) error {
	if bridge == "" {
		return isyscall.NewValidationError("bridge", "", "interface name is empty")
	}
	if len(bridge) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("bridge", bridge, "interface name too long")
	}
	return nil
}

func checkMemberNames(bridge, member string) error {
	if err := checkBridgeName(bridge); err != nil {
		return err
	}
	if member == "" {
		return isyscall.NewValidationError("member", "", "interface name is empty")
	}
	if len(member) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("member", member, "interface name too long")
	}
	return nil
}

// memberReq runs a BRDGGIFFLGS/BRDGSIFFLGS request for member on bridge. For
// a set, flags is sent in ifbr_ifsflags. It returns ifbr_ifsflags as the
// kernel left it.
func memberReq(bridge, member string, cmd uintptr, ioc uintptr, flags uint32) (uint32, error) {
	req := newReq(member)
	defer C.free(unsafe.Pointer(req))
	req.ifbr_ifsflags = C.uint32_t(flags)

	if err := drvspec(bridge, ioc, cmd, unsafe.Pointer(req), unsafe.Sizeof(*req)); err != nil {
		return 0, err
	}
	return uint32(req.ifbr_ifsflags), nil
}

// GetMemberFlags returns the flags (IFBIF_*) of a bridge member
// (BRDGGIFFLGS).
func GetMemberFlags(bridge, member string) (uint32, error) {
	if err := checkMemberNames(bridge, member); err != nil {
		return 0, err
	}
	return memberReq(bridge, member, constants.BRDGGIFFLGS, constants.SIOCGDRVSPEC, 0)
}

// SetMemberFlags replaces the flags of a bridge member (BRDGSIFFLGS).
func SetMemberFlags(bridge, member string, flags uint32) error {
	if err := checkMemberNames(bridge, member); err != nil {
		return err
	}
	_, err := memberReq(bridge, member, constants.BRDGSIFFLGS, constants.SIOCSDRVSPEC, flags)
	return err
}

// UnstorableMemberFlags are the flags BRDGSIFFLGS cannot change: SPAN is
// read-only (the kernel answers EINVAL) and ADMEDGE/ADMCOST are reported by
// BRDGGIFFLGS but ignored when set (outside what if_bridge.c handles).
const UnstorableMemberFlags = constants.IFBIF_SPAN | constants.IFBIF_BSTP_ADMEDGE | constants.IFBIF_BSTP_ADMCOST

// ModifyMemberFlag sets or clears the bits in flag on a bridge member and
// leaves its other flags as they are. The flags are read, changed and written
// back, so this is not atomic against another process changing the same
// member. A flag that is already in the requested state is not written.
func ModifyMemberFlag(bridge, member string, flag uint32, set bool) error {
	if err := checkMemberNames(bridge, member); err != nil {
		return err
	}
	if flag&UnstorableMemberFlags != 0 {
		return isyscall.NewValidationError("flag", fmt.Sprintf("%#x", flag),
			"flag cannot be changed on a bridge member")
	}
	cur, err := GetMemberFlags(bridge, member)
	if err != nil {
		return err
	}
	want := cur &^ flag
	if set {
		want = cur | flag
	}
	if want == cur {
		return nil
	}
	return SetMemberFlags(bridge, member, want)
}
