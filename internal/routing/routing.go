//go:build freebsd
// +build freebsd

package routing

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <net/if.h>
#include <net/route.h>
#include <netinet/in.h>
#include <string.h>
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

type rtMsghdr struct {
	msglen  uint16
	version uint8
	msgtype uint8
	index   uint16
	flags   int32
	addrs   int32
	pid     int32
	seq     int32
	errno   int32
	use     int32
	inits   uint32
	rmx     [14]uint32
}

// ModifyRoute adds or deletes a route (supports IPv4 and IPv6)
func ModifyRoute(add bool, dst *net.IPNet, gw net.IP, ifindex int) error {
	s, err := isyscall.CreateRouteSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	var op int
	if add {
		op = constants.RTM_ADD
	} else {
		op = constants.RTM_DELETE
	}

	buf := new(bytes.Buffer)

	flags := constants.RTF_UP | constants.RTF_STATIC
	if gw != nil {
		flags |= constants.RTF_GATEWAY
	}
	ones, bits := dst.Mask.Size()
	if ones == bits {
		flags |= constants.RTF_HOST
	}

	// Detect IPv4 vs IPv6
	isIPv6 := dst.IP.To4() == nil
	var family int32
	if isIPv6 {
		family = constants.AF_INET6
	} else {
		family = constants.AF_INET
	}

	hdr := rtMsghdr{
		msglen:  0,
		version: uint8(constants.RTM_VERSION),
		msgtype: uint8(op),
		index:   uint16(ifindex),
		flags:   int32(flags),
		addrs:   constants.RTA_DST | constants.RTA_GATEWAY | constants.RTA_NETMASK,
		pid:     0,
		seq:     1,
	}

	binary.Write(buf, binary.LittleEndian, hdr)

	if isIPv6 {
		writeSockaddr(buf, dst.IP, constants.AF_INET6)
	} else {
		writeSockaddr(buf, dst.IP.To4(), constants.AF_INET)
	}

	if gw != nil {
		if isIPv6 {
			writeSockaddr(buf, gw, constants.AF_INET6)
		} else {
			writeSockaddr(buf, gw.To4(), constants.AF_INET)
		}
	} else {
		if isIPv6 {
			writeSockaddr(buf, net.IPv6zero, constants.AF_INET6)
		} else {
			writeSockaddr(buf, net.IPv4zero, constants.AF_INET)
		}
	}

	writeSockaddrMask(buf, dst.Mask, family)

	msgBytes := buf.Bytes()
	msglen := uint16(len(msgBytes))
	binary.LittleEndian.PutUint16(msgBytes[0:2], msglen)

	n, err := syscall.Write(s.Int(), msgBytes)
	if err != nil {
		if add && err == syscall.EEXIST {
			return nil // Idempotent
		}
		if !add && (err == syscall.ESRCH || err == syscall.ENOENT) {
			return nil // Idempotent
		}
		return isyscall.MapError(err.(syscall.Errno))
	}

	if n != len(msgBytes) {
		return fmt.Errorf("incomplete write to routing socket: %d of %d bytes", n, len(msgBytes))
	}

	return nil
}

func writeSockaddr(buf *bytes.Buffer, ip net.IP, family int32) {
	for buf.Len()%int(unsafe.Sizeof(C.long(0))) != 0 {
		buf.WriteByte(0)
	}

	if family == constants.AF_INET {
		sa := make([]byte, syscall.SizeofSockaddrInet4)
		sa[0] = byte(syscall.SizeofSockaddrInet4)
		sa[1] = byte(syscall.AF_INET)
		copy(sa[4:8], ip.To4())
		buf.Write(sa)
	} else if family == constants.AF_INET6 {
		sa := make([]byte, syscall.SizeofSockaddrInet6)
		sa[0] = byte(syscall.SizeofSockaddrInet6)
		sa[1] = byte(syscall.AF_INET6)
		copy(sa[8:24], ip.To16())
		buf.Write(sa)
	}
}

func writeSockaddrMask(buf *bytes.Buffer, mask net.IPMask, family int32) {
	for buf.Len()%int(unsafe.Sizeof(C.long(0))) != 0 {
		buf.WriteByte(0)
	}

	if family == constants.AF_INET {
		sa := make([]byte, syscall.SizeofSockaddrInet4)
		sa[0] = byte(syscall.SizeofSockaddrInet4)
		sa[1] = byte(syscall.AF_INET)
		copy(sa[4:8], mask)
		buf.Write(sa)
	} else if family == constants.AF_INET6 {
		sa := make([]byte, syscall.SizeofSockaddrInet6)
		sa[0] = byte(syscall.SizeofSockaddrInet6)
		sa[1] = byte(syscall.AF_INET6)
		copy(sa[8:24], mask)
		buf.Write(sa)
	}
}
