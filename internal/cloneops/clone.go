//go:build freebsd
// +build freebsd

package cloneops

/*
#include <sys/types.h>
#include <net/if.h>
#include <string.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Create creates a cloned interface with the given prefix.
//
// If the kernel does not know the cloner (SIOCIFCREATE fails with EINVAL,
// which is what happens when e.g. if_epair.ko is not loaded), the matching
// if_<driver> module is loaded and the create is retried once. This mirrors
// what ifconfig(8) does in ifmaybeload().
func Create(prefix string) (string, error) {
	name, err := create(prefix)
	if err == nil || !errors.Is(err, isyscall.ErrInvalidArgument) {
		return name, err
	}

	loaded, kerr := isyscall.LoadKernelModule("if_" + driverName(prefix))
	if kerr != nil || !loaded {
		// Module missing, already loaded, or no privilege to load it:
		// the original EINVAL is the meaningful error.
		return "", err
	}
	return create(prefix)
}

func create(prefix string) (string, error) {
	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return "", err
	}
	defer s.Close()

	var ifr C.struct_ifreq
	if len(prefix) >= constants.IFNAMSIZ {
		return "", fmt.Errorf("interface name too long: %s", prefix)
	}
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), prefix, constants.IFNAMSIZ)

	if err := isyscall.Ioctl(s.Int(), constants.SIOCIFCREATE, unsafe.Pointer(&ifr)); err != nil {
		return "", err
	}

	return C.GoString(&ifr.ifr_name[0]), nil
}

// driverName strips a trailing unit number: "bridge5" -> "bridge".
func driverName(prefix string) string {
	return strings.TrimRight(prefix, "0123456789")
}

// Destroy destroys a cloned interface
func Destroy(name string) error {
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

	return isyscall.Ioctl(s.Int(), constants.SIOCIFDESTROY, unsafe.Pointer(&ifr))
}
