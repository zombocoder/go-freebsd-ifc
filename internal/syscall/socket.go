//go:build freebsd
// +build freebsd

package syscall

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/ioctl.h>
#include <errno.h>
#include <unistd.h>

static int create_inet_socket() {
    return socket(AF_INET, SOCK_DGRAM, 0);
}

static int create_inet6_socket() {
    return socket(AF_INET6, SOCK_DGRAM, 0);
}

static int create_route_socket() {
    return socket(PF_ROUTE, SOCK_RAW, AF_INET);
}

static int get_errno() {
    return errno;
}

static void close_fd(int fd) {
    close(fd);
}
*/
import "C"
import (
	"syscall"
)

// Socket represents a system socket file descriptor
type Socket int

// CreateInetSocket creates an AF_INET SOCK_DGRAM socket for ioctl operations
func CreateInetSocket() (Socket, error) {
	s := Socket(C.create_inet_socket())
	if s < 0 {
		return -1, mapErrno(syscall.Errno(C.get_errno()))
	}
	return s, nil
}

// CreateInet6Socket creates an AF_INET6 SOCK_DGRAM socket for ioctl operations
func CreateInet6Socket() (Socket, error) {
	s := Socket(C.create_inet6_socket())
	if s < 0 {
		return -1, mapErrno(syscall.Errno(C.get_errno()))
	}
	return s, nil
}

// CreateRouteSocket creates a PF_ROUTE socket for routing operations
func CreateRouteSocket() (Socket, error) {
	s := Socket(C.create_route_socket())
	if s < 0 {
		return -1, mapErrno(syscall.Errno(C.get_errno()))
	}
	return s, nil
}

// Close closes the socket
func (s Socket) Close() {
	C.close_fd(C.int(s))
}

// Int returns the socket as an int
func (s Socket) Int() int {
	return int(s)
}

// GetErrno returns the current errno value
func GetErrno() syscall.Errno {
	return syscall.Errno(C.get_errno())
}
