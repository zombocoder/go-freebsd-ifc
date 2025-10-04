//go:build freebsd
// +build freebsd

package ifc

import (
	"net"

	"github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Interface represents a network interface with its configuration and addresses.
type Interface struct {
	Name  string         // Interface name (e.g., "em0", "bridge0")
	Index int            // Kernel interface index
	MTU   int            // Maximum Transmission Unit
	Flags InterfaceFlags // Interface flags (up, running, etc.)
	Addrs []net.Addr     // Assigned IP addresses (IPv4 and IPv6)
}

// InterfaceFlags represents interface flags (IFF_*).
//
// Use the helper methods (IsUp, IsRunning, IsLoopback) for common checks.
type InterfaceFlags uint32

const (
	FlagUp           InterfaceFlags = 0x1
	FlagBroadcast    InterfaceFlags = 0x2
	FlagDebug        InterfaceFlags = 0x4
	FlagLoopback     InterfaceFlags = 0x8
	FlagPointToPoint InterfaceFlags = 0x10
	FlagNoTrailers   InterfaceFlags = 0x20
	FlagRunning      InterfaceFlags = 0x40
	FlagNoARP        InterfaceFlags = 0x80
	FlagPromisc      InterfaceFlags = 0x100
	FlagAllMulti     InterfaceFlags = 0x200
	FlagMulticast    InterfaceFlags = 0x8000
)

func (f InterfaceFlags) IsUp() bool       { return f&FlagUp != 0 }
func (f InterfaceFlags) IsRunning() bool  { return f&FlagRunning != 0 }
func (f InterfaceFlags) IsLoopback() bool { return f&FlagLoopback != 0 }

// Re-export common errors from internal package
var (
	ErrPermission = syscall.ErrPermission
	ErrNotFound   = syscall.ErrNotFound
	ErrExists     = syscall.ErrExists
	ErrSyscall    = syscall.ErrSyscall
)
