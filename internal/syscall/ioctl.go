//go:build freebsd
// +build freebsd

package syscall

/*
#include <sys/types.h>
#include <sys/ioctl.h>
#include <string.h>
#include <errno.h>

static int do_ioctl(int fd, unsigned long request, void *data) {
    return ioctl(fd, request, data);
}

static int get_errno() {
    return errno;
}
*/
import "C"
import (
	"syscall"
	"unsafe"
)

// Ioctl performs an ioctl system call
func Ioctl(fd int, request uintptr, data unsafe.Pointer) error {
	ret := C.do_ioctl(C.int(fd), C.ulong(request), data)
	if ret != 0 {
		return mapErrno(syscall.Errno(C.get_errno()))
	}
	return nil
}

// CopyString copies a Go string to a C buffer safely
func CopyString(dst unsafe.Pointer, src string, maxLen int) {
	srcBytes := []byte(src)
	if len(srcBytes) > maxLen {
		srcBytes = srcBytes[:maxLen]
	}
	C.memcpy(dst, unsafe.Pointer(&srcBytes[0]), C.size_t(len(srcBytes)))
}

// CopyBytes copies Go bytes to a C buffer
func CopyBytes(dst, src unsafe.Pointer, n int) {
	C.memcpy(dst, src, C.size_t(n))
}
