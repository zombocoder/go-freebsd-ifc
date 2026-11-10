//go:build freebsd
// +build freebsd

package ifops

/*
#include <sys/types.h>
#include <net/if.h>
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

// checkIfName validates an interface name before it is copied into a fixed
// IFNAMSIZ buffer (a name of IFNAMSIZ or more bytes would not be NUL
// terminated). The error matches ErrInvalidArgument.
func checkIfName(name string) error {
	if name == "" {
		return isyscall.NewValidationError("interface name", "", "name is empty")
	}
	if len(name) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("interface name", name, "name too long")
	}
	return nil
}

// SetFlags modifies interface flags
func SetFlags(name string, flag uint32, set bool) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var ifr C.struct_ifreq
	if err := checkIfName(name); err != nil {
		return err
	}
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)

	if err := isyscall.Ioctl(s.Int(), constants.SIOCGIFFLAGS, unsafe.Pointer(&ifr)); err != nil {
		return err
	}

	oldFlags := *(*C.int)(unsafe.Pointer(&ifr.ifr_ifru))
	if set {
		*(*C.int)(unsafe.Pointer(&ifr.ifr_ifru)) = oldFlags | C.int(flag)
	} else {
		*(*C.int)(unsafe.Pointer(&ifr.ifr_ifru)) = oldFlags & ^C.int(flag)
	}

	return isyscall.Ioctl(s.Int(), constants.SIOCSIFFLAGS, unsafe.Pointer(&ifr))
}

// GetMTU returns the MTU of an interface
func GetMTU(name string) (int, error) {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return 0, err
	}
	defer s.Close()

	var ifr C.struct_ifreq
	if err := checkIfName(name); err != nil {
		return 0, err
	}
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)

	if err := isyscall.Ioctl(s.Int(), constants.SIOCGIFMTU, unsafe.Pointer(&ifr)); err != nil {
		return 0, err
	}

	return int(*(*C.int)(unsafe.Pointer(&ifr.ifr_ifru))), nil
}

// SetMTU sets the MTU of an interface
func SetMTU(name string, mtu int) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var ifr C.struct_ifreq
	if err := checkIfName(name); err != nil {
		return err
	}
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)

	*(*C.int)(unsafe.Pointer(&ifr.ifr_ifru)) = C.int(mtu)

	return isyscall.Ioctl(s.Int(), constants.SIOCSIFMTU, unsafe.Pointer(&ifr))
}

// Rename renames an interface
func Rename(oldName, newName string) error {
	if err := checkIfName(oldName); err != nil {
		return err
	}
	if err := checkIfName(newName); err != nil {
		return err
	}

	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	// SIOCSIFNAME does not take the new name inline. sys/net/if.c:2715 does
	//
	//	error = copyinstr(ifr_data_get_ptr(ifr), new_name, IFNAMSIZ, NULL);
	//
	// so ifr_ifru holds a *pointer* to a NUL-terminated string in the
	// caller's address space. Copying the ASCII name into the union makes
	// copyinstr() treat those bytes as an address and fail with EFAULT.
	//
	// The buffer is allocated with calloc() rather than taken from a Go
	// slice: the kernel dereferences it from inside the ioctl, and a Go
	// pointer parked in a struct that is handed to C is both a cgo pointer
	// rule violation and invisible to the collector, so nothing would keep
	// the buffer alive for the duration of the call.
	buf := C.calloc(1, C.size_t(constants.IFNAMSIZ))
	if buf == nil {
		return fmt.Errorf("allocate rename buffer for %s", newName)
	}
	defer C.free(buf)
	// Leave room for the terminating NUL calloc already wrote.
	isyscall.CopyString(buf, newName, constants.IFNAMSIZ-1)

	var ifr C.struct_ifreq
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), oldName, constants.IFNAMSIZ)
	*(**C.char)(unsafe.Pointer(&ifr.ifr_ifru)) = (*C.char)(buf)

	return isyscall.Ioctl(s.Int(), constants.SIOCSIFNAME, unsafe.Pointer(&ifr))
}

// SetVNET moves the host interface name into the vnet of jail jid
// (SIOCSIFVNET, the ioctl behind "ifconfig <name> vnet <jid>").
func SetVNET(name string, jid int) error {
	return vnetIoctl(constants.SIOCSIFVNET, name, jid)
}

// ReclaimVNET moves interface name, as it is called inside jail jid, back
// into the caller's vnet (SIOCSIFRVNET, "ifconfig <name> -vnet <jid>").
func ReclaimVNET(name string, jid int) error {
	return vnetIoctl(constants.SIOCSIFRVNET, name, jid)
}

func vnetIoctl(req uintptr, name string, jid int) error {
	if err := checkIfName(name); err != nil {
		return err
	}
	if jid <= 0 {
		return fmt.Errorf("invalid jail id: %d", jid)
	}

	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var ifr C.struct_ifreq
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)
	// ifr_jid is the int member of the ifr_ifru union.
	*(*C.int)(unsafe.Pointer(&ifr.ifr_ifru)) = C.int(jid)

	return isyscall.Ioctl(s.Int(), req, unsafe.Pointer(&ifr))
}
