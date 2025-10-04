//go:build freebsd
// +build freebsd

package syscall

import (
	"fmt"
	"syscall"
)

// MapErrno maps a syscall errno to a typed error
func mapErrno(err syscall.Errno) error {
	switch err {
	case syscall.EPERM, syscall.EACCES:
		return ErrPermission
	case syscall.ENOENT, syscall.ENXIO:
		return ErrNotFound
	case syscall.EEXIST:
		return ErrExists
	case syscall.EINVAL:
		return ErrInvalidArgument
	case syscall.EBUSY:
		return ErrBusy
	case syscall.EOPNOTSUPP:
		return ErrNotSupported
	case syscall.ENETDOWN:
		return ErrNetworkDown
	case syscall.EADDRINUSE:
		return ErrAddressInUse
	default:
		return fmt.Errorf("%w: %v", ErrSyscall, err)
	}
}

// MapError is the public version of mapErrno
func MapError(err syscall.Errno) error {
	return mapErrno(err)
}
