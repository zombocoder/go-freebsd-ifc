//go:build freebsd
// +build freebsd

package syscall

/*
#include <sys/param.h>
#include <sys/linker.h>
#include <stdlib.h>
*/
import "C"
import (
	"strings"
	"syscall"
	"unsafe"
)

// LoadKernelModule loads the named kernel module with kldload(2).
//
// It reports loaded=true only when this call actually loaded the module.
// A module that is already present is not an error: loaded=false, err=nil.
//
// name must be a bare module name. kldload(2) treats a name containing '/' as
// a file path, so such names (and names containing NUL) are refused with
// ErrInvalidArgument without calling the kernel.
func LoadKernelModule(name string) (loaded bool, err error) {
	if name == "" || strings.ContainsAny(name, "/\x00") {
		return false, ErrInvalidArgument
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if ret, errno := C.kldload(cname); ret < 0 {
		if errno == syscall.EEXIST {
			return false, nil
		}
		if e, ok := errno.(syscall.Errno); ok {
			return false, mapErrno(e)
		}
		return false, errno
	}
	return true, nil
}
