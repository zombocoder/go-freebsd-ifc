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
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// descrRetries bounds the size/read loop in GetDescription. The description
// can only change between the two ioctls if someone else sets it at that
// moment, so a handful of attempts is plenty.
const descrRetries = 5

func checkDescrName(name string) error {
	if name == "" {
		return fmt.Errorf("interface name is empty")
	}
	if len(name) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long: %s", name)
	}
	return nil
}

// SetDescription sets the description of an interface (SIOCSIFDESCR). An
// empty string clears it.
//
// ifr_ifru holds a struct ifreq_buffer {length, buffer}: the kernel copies
// length-1 bytes from buffer and counts the terminating NUL in length, so
// length is len(descr)+1. A length of 0 removes the description. A length
// above net.ifdescr_maxlen is refused with ENAMETOOLONG.
func SetDescription(name, descr string) error {
	if err := checkDescrName(name); err != nil {
		return err
	}
	if strings.IndexByte(descr, 0) >= 0 {
		return fmt.Errorf("description contains a NUL byte")
	}

	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var ifr C.struct_ifreq
	isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)
	rb := (*C.struct_ifreq_buffer)(unsafe.Pointer(&ifr.ifr_ifru))

	if descr != "" {
		// The kernel reads the string from inside the ioctl, so it lives
		// in C memory (see Rename). calloc provides the trailing NUL.
		n := len(descr) + 1
		buf := C.calloc(1, C.size_t(n))
		if buf == nil {
			return fmt.Errorf("allocate description buffer for %s", name)
		}
		defer C.free(buf)
		isyscall.CopyString(buf, descr, len(descr))
		rb.length = C.size_t(n)
		rb.buffer = buf
	}

	err = isyscall.Ioctl(s.Int(), constants.SIOCSIFDESCR, unsafe.Pointer(&ifr))
	if errors.Is(err, syscall.ENAMETOOLONG) {
		return fmt.Errorf("description of %d bytes exceeds net.ifdescr_maxlen: %w", len(descr), err)
	}
	return err
}

// GetDescription returns the description of an interface (SIOCGIFDESCR), or
// "" when none is set.
//
// The kernel never truncates: when the buffer is too small it returns success
// with buffer cleared and length set to the size needed (NUL included). So
// the first call probes with length 0, the second reads into a buffer of that
// size. An interface without a description answers ENOMSG.
func GetDescription(name string) (string, error) {
	if err := checkDescrName(name); err != nil {
		return "", err
	}

	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return "", err
	}
	defer s.Close()

	size := C.size_t(0) // 0 = probe
	for i := 0; i < descrRetries; i++ {
		var ifr C.struct_ifreq
		isyscall.CopyString(unsafe.Pointer(&ifr.ifr_name[0]), name, constants.IFNAMSIZ)
		rb := (*C.struct_ifreq_buffer)(unsafe.Pointer(&ifr.ifr_ifru))

		var buf unsafe.Pointer
		if size > 0 {
			buf = C.calloc(1, size)
			if buf == nil {
				return "", fmt.Errorf("allocate description buffer for %s", name)
			}
			rb.length = size
			rb.buffer = buf
		}

		err := isyscall.Ioctl(s.Int(), constants.SIOCGIFDESCR, unsafe.Pointer(&ifr))
		var out string
		if err == nil && size > 0 && rb.length <= size {
			out = C.GoString((*C.char)(buf))
		}
		got := rb.length
		if buf != nil {
			C.free(buf)
		}

		if errors.Is(err, syscall.ENOMSG) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		if size > 0 && got <= size {
			return out, nil
		}
		// Probe answer, or the description grew in between: retry with
		// the size the kernel reported.
		size = got
	}
	return "", fmt.Errorf("description of %s kept changing", name)
}
