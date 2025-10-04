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
	"fmt"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Create creates a cloned interface with the given prefix
func Create(prefix string) (string, error) {
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
