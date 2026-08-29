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

// SetFlags modifies interface flags
func SetFlags(name string, flag uint32, set bool) error {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var ifr C.struct_ifreq
	if len(name) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long: %s", name)
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
	if len(name) >= constants.IFNAMSIZ {
		return 0, fmt.Errorf("interface name too long: %s", name)
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
	if len(name) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long: %s", name)
	}
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)

	*(*C.int)(unsafe.Pointer(&ifr.ifr_ifru)) = C.int(mtu)

	return isyscall.Ioctl(s.Int(), constants.SIOCSIFMTU, unsafe.Pointer(&ifr))
}

// Rename renames an interface
func Rename(oldName, newName string) error {
	if len(oldName) >= constants.IFNAMSIZ || len(newName) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long")
	}
	if oldName == "" || newName == "" {
		return fmt.Errorf("interface name is empty")
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
