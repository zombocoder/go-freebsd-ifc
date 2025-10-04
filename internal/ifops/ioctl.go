//go:build freebsd
// +build freebsd

package ifops

/*
#include <sys/types.h>
#include <net/if.h>
#include <string.h>
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
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var ifr C.struct_ifreq

	if len(oldName) >= constants.IFNAMSIZ || len(newName) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long")
	}

	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), oldName, constants.IFNAMSIZ)
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_ifru), newName, constants.IFNAMSIZ)

	return isyscall.Ioctl(s.Int(), constants.SIOCSIFNAME, unsafe.Pointer(&ifr))
}
