//go:build freebsd
// +build freebsd

package routing

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/sysctl.h>
#include <net/if.h>
#include <net/if_dl.h>
#include <net/route.h>
#include <netinet/in.h>
#include <string.h>
#include <stdlib.h>

// route_dump fetches the kernel routing table for the given address family
// (AF_UNSPEC for all) via the CTL_NET/PF_ROUTE/NET_RT_DUMP sysctl. The buffer
// is malloc'd here and must be released with free() by the caller.
static int route_dump(int af, void **bufp, size_t *lenp) {
	int mib[6];
	size_t need;
	void *buf;

	mib[0] = CTL_NET;
	mib[1] = PF_ROUTE;
	mib[2] = 0;
	mib[3] = af;
	mib[4] = NET_RT_DUMP;
	mib[5] = 0;

	if (sysctl(mib, 6, NULL, &need, NULL, 0) < 0)
		return -1;
	if (need == 0) {
		*bufp = NULL;
		*lenp = 0;
		return 0;
	}
	buf = malloc(need);
	if (buf == NULL)
		return -1;
	if (sysctl(mib, 6, buf, &need, NULL, 0) < 0) {
		free(buf);
		return -1;
	}
	*bufp = buf;
	*lenp = need;
	return 0;
}
*/
import "C"
import (
	"bytes"
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// rtMsghdr mirrors struct rt_msghdr from /usr/include/net/route.h.
//
// Layout verified against a C probe on FreeBSD 14.3-RELEASE/amd64:
//
//	field         offset  size
//	rtm_msglen         0     2
//	rtm_version        2     1
//	rtm_type           3     1
//	rtm_index          4     2
//	_rtm_spare1        6     2
//	rtm_flags          8     4
//	rtm_addrs         12     4
//	rtm_pid           16     4
//	rtm_seq           20     4
//	rtm_errno         24     4
//	rtm_fmask         28     4
//	rtm_inits         32     8   (u_long)
//	rtm_rmx           40   112   (struct rt_metrics: 14 x u_long)
//	sizeof(struct rt_msghdr) == 152
//
// routing_abi_test.go pins every one of those numbers so a mismatch is a test
// failure rather than a silent EINVAL from the kernel.
type rtMsghdr struct {
	msglen  uint16     // rtm_msglen
	version uint8      // rtm_version
	msgtype uint8      // rtm_type
	index   uint16     // rtm_index
	spare1  uint16     // _rtm_spare1
	flags   int32      // rtm_flags
	addrs   int32      // rtm_addrs
	pid     int32      // rtm_pid
	seq     int32      // rtm_seq
	errno   int32      // rtm_errno
	fmask   int32      // rtm_fmask
	inits   uint64     // rtm_inits  (u_long on LP64)
	rmx     [14]uint64 // rtm_rmx    (struct rt_metrics, all u_long)
}

// sizeofRtMsghdr is the wire size of a routing message header.
const sizeofRtMsghdr = int(unsafe.Sizeof(rtMsghdr{}))

// Route describes a single entry of the kernel routing table.
type Route struct {
	// Family is the address family the kernel reported in the destination
	// sockaddr: AF_INET or AF_INET6. It is carried explicitly because the
	// shape of Dst cannot be trusted to reveal it. An IPv4-mapped IPv6
	// destination such as ::ffff:0.0.0.0/96 is a real IPv6 route, yet
	// net.IP.To4() returns non-nil for it and net.IPNet.String() prints it
	// as "0.0.0.0/0". sa_family is the only authority.
	Family int
	// Dst is the destination prefix. For host routes the mask is all ones.
	// The address and mask keep the width of their family: 4 bytes for
	// AF_INET, 16 for AF_INET6, never unmapped.
	Dst *net.IPNet
	// Gateway is the next hop, or nil for a link-layer (directly attached)
	// route.
	Gateway net.IP
	// Index is the index of the outgoing interface, 0 when unknown.
	Index int
	// Iface is the name of the outgoing interface, "" when unknown.
	Iface string
	// Flags is the raw RTF_* bitmask from the kernel.
	Flags uint32
}

// bytes returns the raw wire representation of the header in host byte order.
func (h *rtMsghdr) bytes() []byte {
	return (*[sizeofRtMsghdr]byte)(unsafe.Pointer(h))[:]
}

// ModifyRoute adds or deletes a route (supports IPv4 and IPv6)
func ModifyRoute(add bool, dst *net.IPNet, gw net.IP, ifindex int) error {
	if dst == nil {
		return isyscall.NewValidationError("dst", "", "destination network is required")
	}

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

	// Build the sockaddr payload first so that rtm_msglen can be filled in
	// before the header is serialised; the kernel rejects the message
	// outright when rtm_msglen does not equal the write() length
	// (sys/net/rtsock.c, rts_send: "len != mtod(m, struct rt_msghdr *)->rtm_msglen").
	addrs := new(bytes.Buffer)

	if isIPv6 {
		writeSockaddr(addrs, dst.IP, constants.AF_INET6)
	} else {
		writeSockaddr(addrs, dst.IP.To4(), constants.AF_INET)
	}

	if gw != nil {
		if isIPv6 {
			writeSockaddr(addrs, gw, constants.AF_INET6)
		} else {
			writeSockaddr(addrs, gw.To4(), constants.AF_INET)
		}
	} else {
		if isIPv6 {
			writeSockaddr(addrs, net.IPv6zero, constants.AF_INET6)
		} else {
			writeSockaddr(addrs, net.IPv4zero, constants.AF_INET)
		}
	}

	writeSockaddrMask(addrs, dst.Mask, family)

	hdr := rtMsghdr{
		msglen:  uint16(sizeofRtMsghdr + addrs.Len()),
		version: uint8(constants.RTM_VERSION),
		msgtype: uint8(op),
		index:   uint16(ifindex),
		flags:   int32(flags),
		addrs:   constants.RTA_DST | constants.RTA_GATEWAY | constants.RTA_NETMASK,
		pid:     0,
		seq:     1,
	}

	msgBytes := make([]byte, 0, int(hdr.msglen))
	msgBytes = append(msgBytes, hdr.bytes()...)
	msgBytes = append(msgBytes, addrs.Bytes()...)

	n, err := syscall.Write(s.Int(), msgBytes)
	if err != nil {
		errno, ok := err.(syscall.Errno)
		if !ok {
			return err
		}
		if add && errno == syscall.EEXIST {
			return nil // Idempotent
		}
		if !add && (errno == syscall.ESRCH || errno == syscall.ENOENT) {
			return nil // Idempotent
		}
		return isyscall.MapError(errno)
	}

	if n != len(msgBytes) {
		return fmt.Errorf("incomplete write to routing socket: %d of %d bytes", n, len(msgBytes))
	}

	return nil
}

// List returns the kernel routing table for the given address family.
//
// family may be AF_INET, AF_INET6 or AF_UNSPEC (all families). The table is
// read with the CTL_NET/PF_ROUTE/NET_RT_DUMP sysctl, the same interface
// netstat(1) uses, so no privileges are required.
func List(family int) ([]Route, error) {
	var buf unsafe.Pointer
	var length C.size_t

	if C.route_dump(C.int(family), &buf, &length) < 0 {
		return nil, isyscall.MapError(isyscall.GetErrno())
	}
	if buf == nil || length == 0 {
		return []Route{}, nil
	}
	defer C.free(buf)

	raw := unsafe.Slice((*byte)(buf), int(length))
	routes := make([]Route, 0, 16)

	for off := 0; off+sizeofRtMsghdr <= len(raw); {
		hdr := (*rtMsghdr)(unsafe.Pointer(&raw[off]))
		msglen := int(hdr.msglen)
		if msglen < sizeofRtMsghdr || off+msglen > len(raw) {
			break
		}
		if hdr.version == uint8(constants.RTM_VERSION) {
			if r, ok := parseRoute(hdr, raw[off+sizeofRtMsghdr:off+msglen]); ok {
				routes = append(routes, r)
			}
		}
		off += msglen
	}

	return routes, nil
}

// saSize mirrors the SA_SIZE() macro from <net/route.h>: sockaddrs in a
// routing message are padded up to a multiple of sizeof(long), and a zero
// length sockaddr still occupies one such slot.
func saSize(salen int) int {
	if salen == 0 {
		return int(unsafe.Sizeof(C.long(0)))
	}
	align := int(unsafe.Sizeof(C.long(0)))
	return 1 + ((salen - 1) | (align - 1))
}

// splitSockaddrs walks the sockaddr array that follows a routing header and
// returns the segments selected by the rtm_addrs bitmask, indexed by RTAX_*.
func splitSockaddrs(addrs int32, buf []byte) [][]byte {
	const rtaxMax = 8
	out := make([][]byte, rtaxMax)

	pos := 0
	for i := 0; i < rtaxMax; i++ {
		if addrs&(1<<uint(i)) == 0 {
			continue
		}
		if pos >= len(buf) {
			break
		}
		salen := int(buf[pos])
		avail := len(buf) - pos
		if salen > avail {
			salen = avail
		}
		out[i] = buf[pos : pos+salen]
		pos += saSize(salen)
		if pos > len(buf) {
			break
		}
	}
	return out
}

// sockaddrIP extracts an IP address from a raw sockaddr of the given family.
func sockaddrIP(sa []byte, family int) net.IP {
	switch family {
	case constants.AF_INET:
		ip := make(net.IP, net.IPv4len)
		copy(ip, sliceRange(sa, 4, 4))
		return ip
	case constants.AF_INET6:
		ip := make(net.IP, net.IPv6len)
		copy(ip, sliceRange(sa, 8, 16))
		return ip
	}
	return nil
}

// sliceRange returns up to n bytes of sa starting at off, tolerating a
// sockaddr that the kernel truncated (netmasks in particular carry only the
// significant leading bytes).
func sliceRange(sa []byte, off, n int) []byte {
	if off >= len(sa) {
		return nil
	}
	end := off + n
	if end > len(sa) {
		end = len(sa)
	}
	return sa[off:end]
}

// sockaddrDLName extracts the interface name from an AF_LINK sockaddr_dl.
func sockaddrDLName(sa []byte) string {
	// struct sockaddr_dl (verified with a C probe on FreeBSD 14.3/amd64):
	// sdl_len 0, sdl_family 1, sdl_index 2, sdl_type 4, sdl_nlen 5,
	// sdl_alen 6, sdl_slen 7, sdl_data 8. The name occupies the first
	// sdl_nlen bytes of sdl_data.
	const (
		nlenOff = 5
		dataOff = 8
	)
	if len(sa) < dataOff {
		return ""
	}
	nlen := int(sa[nlenOff])
	if nlen <= 0 || dataOff+nlen > len(sa) {
		return ""
	}
	return string(sa[dataOff : dataOff+nlen])
}

func parseRoute(hdr *rtMsghdr, sas []byte) (Route, bool) {
	const (
		rtaxDst     = 0
		rtaxGateway = 1
		rtaxNetmask = 2
	)

	seg := splitSockaddrs(hdr.addrs, sas)

	dstSa := seg[rtaxDst]
	if len(dstSa) < 2 {
		return Route{}, false
	}
	family := int(dstSa[1])
	if family != constants.AF_INET && family != constants.AF_INET6 {
		return Route{}, false
	}

	dstIP := sockaddrIP(dstSa, family)
	if dstIP == nil {
		return Route{}, false
	}

	bits := net.IPv4len * 8
	maskOff := 4
	if family == constants.AF_INET6 {
		bits = net.IPv6len * 8
		maskOff = 8
	}

	var mask net.IPMask
	switch {
	case hdr.flags&int32(constants.RTF_HOST) != 0:
		mask = net.CIDRMask(bits, bits)
	default:
		mask = make(net.IPMask, bits/8)
		copy(mask, sliceRange(seg[rtaxNetmask], maskOff, bits/8))
	}

	r := Route{
		Family: family,
		Dst:    &net.IPNet{IP: dstIP, Mask: mask},
		Index:  int(hdr.index),
		Flags:  uint32(hdr.flags),
	}

	if gwSa := seg[rtaxGateway]; len(gwSa) >= 2 {
		switch int(gwSa[1]) {
		case constants.AF_LINK:
			r.Iface = sockaddrDLName(gwSa)
		case constants.AF_INET, constants.AF_INET6:
			r.Gateway = sockaddrIP(gwSa, int(gwSa[1]))
		}
	}

	if r.Iface == "" && r.Index > 0 {
		var nameBuf [C.IFNAMSIZ]C.char
		if C.if_indextoname(C.uint(r.Index), &nameBuf[0]) != nil {
			r.Iface = C.GoString(&nameBuf[0])
		}
	}

	return r, true
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
