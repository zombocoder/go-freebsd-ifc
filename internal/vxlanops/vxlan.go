//go:build freebsd
// +build freebsd

package vxlanops

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/sysctl.h>
#include <net/if.h>
#include <net/if_vxlan.h>
#include <net/ethernet.h>
#include <netinet/in.h>
#include <string.h>
#include <stdlib.h>

// sysctl_string reads a string sysctl by name into a zeroed buffer. On
// success *lenp is the number of bytes the kernel actually returned. The
// caller frees *bufp. Returns -1 on failure with errno set.
static int sysctl_string(const char *name, char **bufp, size_t *lenp) {
	size_t need;
	char *buf;

	if (sysctlbyname(name, NULL, &need, NULL, 0) < 0)
		return -1;
	// Ask for a little slack: the table can grow between the sizing call
	// and the fetch. calloc, not malloc, so a short reply cannot leave
	// stale heap bytes behind the data.
	need += 4096;
	buf = calloc(1, need);
	if (buf == NULL)
		return -1;
	if (sysctlbyname(name, buf, &need, NULL, 0) < 0) {
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
	"fmt"
	"net"
	"strconv"
	"strings"
	"unsafe"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Config mirrors struct ifvxlancfg, the reply to VXLAN_CMD_GET_CONFIG.
type Config struct {
	VNI            uint32
	Local          net.IP
	LocalPort      uint16
	Remote         net.IP
	RemotePort     uint16
	MulticastIndex int
	FtableCount    uint32
	FtableMax      uint32
	FtableTimeout  uint32
	PortMin        uint16
	PortMax        uint16
	Learn          bool
	TTL            uint8
}

// FtableEntry is one row of a vxlan interface's forwarding table: the remote
// VTEP that frames for a given MAC address are sent to.
type FtableEntry struct {
	MAC    net.HardwareAddr
	Remote net.IP
	Static bool
	Expire int64
}

// htons converts a host order port to network byte order.
func htons(v uint16) uint16 {
	return v<<8 | v>>8
}

// ntohs converts a network order port to host byte order.
func ntohs(v uint16) uint16 {
	return v<<8 | v>>8
}

// doCmd issues one VXLAN driver command.
//
// The kernel dispatches these through vxlan_ioctl_drvspec() in
// sys/net/if_vxlan.c: SIOCGDRVSPEC for the single COPYOUT command
// (VXLAN_CMD_GET_CONFIG) and SIOCSDRVSPEC for every COPYIN command. ifd_len
// must equal the exact vxlc_argsize the control table declares or the kernel
// answers EINVAL.
//
// data must point at C memory: the kernel dereferences it during the ioctl and
// storing a Go pointer inside the C-allocated ifdrv would break the cgo
// pointer rules.
func doCmd(name string, cmd int, data unsafe.Pointer, size int, set bool) error {
	if name == "" || len(name) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("name", name, "invalid interface name")
	}

	s, err := isyscall.CreateInetSocket()
	if err != nil {
		return err
	}
	defer s.Close()

	ifd := (*C.struct_ifdrv)(C.calloc(1, C.sizeof_struct_ifdrv))
	if ifd == nil {
		return fmt.Errorf("allocate ifdrv for %s", name)
	}
	defer C.free(unsafe.Pointer(ifd))

	isyscall.CopyString(unsafe.Pointer(&ifd.ifd_name[0]), name, constants.IFNAMSIZ-1)
	ifd.ifd_cmd = C.ulong(cmd)
	ifd.ifd_len = C.size_t(size)
	ifd.ifd_data = data

	req := uintptr(constants.SIOCSDRVSPEC)
	if !set {
		req = uintptr(constants.SIOCGDRVSPEC)
	}
	return isyscall.Ioctl(s.Int(), req, unsafe.Pointer(ifd))
}

// newCmd allocates a zeroed struct ifvxlancmd in C memory.
func newCmd() (*C.struct_ifvxlancmd, func(), error) {
	p := (*C.struct_ifvxlancmd)(C.calloc(1, C.sizeof_struct_ifvxlancmd))
	if p == nil {
		return nil, nil, fmt.Errorf("allocate ifvxlancmd")
	}
	return p, func() { C.free(unsafe.Pointer(p)) }, nil
}

// setCmd runs a COPYIN command with a freshly zeroed ifvxlancmd that fill
// populates.
func setCmd(name string, cmd int, fill func(*C.struct_ifvxlancmd)) error {
	c, free, err := newCmd()
	if err != nil {
		return err
	}
	defer free()

	fill(c)
	return doCmd(name, cmd, unsafe.Pointer(c), constants.SizeofIfvxlancmd, true)
}

// writeSockaddr fills a union vxlan_sockaddr with ip and port. The kernel
// requires sa_family to be AF_INET or AF_INET6 (VXLAN_SOCKADDR_IS_IPV46).
func writeSockaddr(dst unsafe.Pointer, ip net.IP, port uint16) error {
	if ip == nil {
		return isyscall.NewValidationError("addr", "", "address is required")
	}
	if ip4 := ip.To4(); ip4 != nil {
		sin := (*C.struct_sockaddr_in)(dst)
		sin.sin_len = C.uchar(constants.SizeofSockaddrIn)
		sin.sin_family = C.uchar(constants.AF_INET)
		sin.sin_port = C.ushort(htons(port))
		isyscall.CopyBytes(unsafe.Pointer(&sin.sin_addr), unsafe.Pointer(&ip4[0]), 4)
		return nil
	}
	ip16 := ip.To16()
	if ip16 == nil {
		return isyscall.NewValidationError("addr", ip.String(), "not an IP address")
	}
	sin6 := (*C.struct_sockaddr_in6)(dst)
	sin6.sin6_len = C.uchar(constants.SizeofSockaddrIn6)
	sin6.sin6_family = C.uchar(constants.AF_INET6)
	sin6.sin6_port = C.ushort(htons(port))
	isyscall.CopyBytes(unsafe.Pointer(&sin6.sin6_addr), unsafe.Pointer(&ip16[0]), 16)
	return nil
}

// readSockaddr decodes a union vxlan_sockaddr the kernel filled in.
//
// The port is read regardless of the address family. vxlan_set_default_config()
// stores the default port straight into vxl_src_addr.in4.sin_port without ever
// setting sin_family, so a freshly cloned interface has a valid port and an
// AF_UNSPEC address; sin_port and sin6_port share offset 2, so one read covers
// both. The address itself is only returned once the family says which of the
// two it is.
func readSockaddr(src unsafe.Pointer) (net.IP, uint16) {
	sin := (*C.struct_sockaddr_in)(src)
	port := ntohs(uint16(sin.sin_port))

	sa := (*C.struct_sockaddr)(src)
	switch int(sa.sa_family) {
	case constants.AF_INET:
		ip := make(net.IP, net.IPv4len)
		isyscall.CopyBytes(unsafe.Pointer(&ip[0]), unsafe.Pointer(&sin.sin_addr), 4)
		return ip, port
	case constants.AF_INET6:
		sin6 := (*C.struct_sockaddr_in6)(src)
		ip := make(net.IP, net.IPv6len)
		isyscall.CopyBytes(unsafe.Pointer(&ip[0]), unsafe.Pointer(&sin6.sin6_addr), 16)
		return ip, port
	}
	return nil, port
}

// GetConfig returns the kernel's view of a vxlan interface.
func GetConfig(name string) (Config, error) {
	cfg := (*C.struct_ifvxlancfg)(C.calloc(1, C.sizeof_struct_ifvxlancfg))
	if cfg == nil {
		return Config{}, fmt.Errorf("allocate ifvxlancfg")
	}
	defer C.free(unsafe.Pointer(cfg))

	if err := doCmd(name, constants.VXLAN_CMD_GET_CONFIG, unsafe.Pointer(cfg),
		constants.SizeofIfvxlancfg, false); err != nil {
		return Config{}, err
	}

	local, localPort := readSockaddr(unsafe.Pointer(&cfg.vxlc_local_sa[0]))
	remote, remotePort := readSockaddr(unsafe.Pointer(&cfg.vxlc_remote_sa[0]))

	return Config{
		VNI:            uint32(cfg.vxlc_vni),
		Local:          local,
		LocalPort:      localPort,
		Remote:         remote,
		RemotePort:     remotePort,
		MulticastIndex: int(cfg.vxlc_mc_ifindex),
		FtableCount:    uint32(cfg.vxlc_ftable_cnt),
		FtableMax:      uint32(cfg.vxlc_ftable_max),
		FtableTimeout:  uint32(cfg.vxlc_ftable_timeout),
		PortMin:        uint16(cfg.vxlc_port_min),
		PortMax:        uint16(cfg.vxlc_port_max),
		Learn:          cfg.vxlc_learn != 0,
		TTL:            uint8(cfg.vxlc_ttl),
	}, nil
}

// SetVNI sets the virtual network identifier.
func SetVNI(name string, vni uint32) error {
	if vni >= constants.VXLAN_VNI_MAX {
		return isyscall.NewValidationError("vni", strconv.FormatUint(uint64(vni), 10),
			fmt.Sprintf("must be less than %d", constants.VXLAN_VNI_MAX))
	}
	return setCmd(name, constants.VXLAN_CMD_SET_VNI, func(c *C.struct_ifvxlancmd) {
		c.vxlcmd_vni = C.uint32_t(vni)
	})
}

// SetLocalAddr sets the source address of the tunnel endpoint. The kernel
// rejects multicast addresses here.
func SetLocalAddr(name string, ip net.IP) error {
	var saErr error
	err := setCmd(name, constants.VXLAN_CMD_SET_LOCAL_ADDR, func(c *C.struct_ifvxlancmd) {
		saErr = writeSockaddr(unsafe.Pointer(&c.vxlcmd_sa[0]), ip, 0)
	})
	if saErr != nil {
		return saErr
	}
	return err
}

// SetRemoteAddr sets the destination address of the tunnel. A multicast
// address here makes the interface join that group instead of using a single
// unicast peer.
func SetRemoteAddr(name string, ip net.IP) error {
	var saErr error
	err := setCmd(name, constants.VXLAN_CMD_SET_REMOTE_ADDR, func(c *C.struct_ifvxlancmd) {
		saErr = writeSockaddr(unsafe.Pointer(&c.vxlcmd_sa[0]), ip, 0)
	})
	if saErr != nil {
		return saErr
	}
	return err
}

// SetLocalPort sets the source UDP port. vxlcmd_port is in host byte order;
// the kernel applies htons() itself.
func SetLocalPort(name string, port uint16) error {
	if port == 0 {
		return isyscall.NewValidationError("port", "0", "must be non-zero")
	}
	return setCmd(name, constants.VXLAN_CMD_SET_LOCAL_PORT, func(c *C.struct_ifvxlancmd) {
		c.vxlcmd_port = C.uint16_t(port)
	})
}

// SetRemotePort sets the destination UDP port (host byte order).
func SetRemotePort(name string, port uint16) error {
	if port == 0 {
		return isyscall.NewValidationError("port", "0", "must be non-zero")
	}
	return setCmd(name, constants.VXLAN_CMD_SET_REMOTE_PORT, func(c *C.struct_ifvxlancmd) {
		c.vxlcmd_port = C.uint16_t(port)
	})
}

// SetPortRange sets the range of source ports used for entropy.
func SetPortRange(name string, min, max uint16) error {
	if max < min {
		return isyscall.NewValidationError("port range",
			fmt.Sprintf("%d-%d", min, max), "max must not be less than min")
	}
	return setCmd(name, constants.VXLAN_CMD_SET_PORT_RANGE, func(c *C.struct_ifvxlancmd) {
		c.vxlcmd_port_min = C.uint16_t(min)
		c.vxlcmd_port_max = C.uint16_t(max)
	})
}

// SetMulticastIf sets the interface used to send multicast traffic.
func SetMulticastIf(name, dev string) error {
	if len(dev) >= constants.IFNAMSIZ {
		return isyscall.NewValidationError("dev", dev, "interface name too long")
	}
	return setCmd(name, constants.VXLAN_CMD_SET_MULTICAST_IF, func(c *C.struct_ifvxlancmd) {
		if dev != "" {
			isyscall.CopyString(unsafe.Pointer(&c.vxlcmd_ifname[0]), dev, constants.IFNAMSIZ-1)
		}
	})
}

// SetTTL sets the TTL (IPv4) or hop limit (IPv6) of the outer header.
func SetTTL(name string, ttl uint8) error {
	return setCmd(name, constants.VXLAN_CMD_SET_TTL, func(c *C.struct_ifvxlancmd) {
		c.vxlcmd_ttl = C.uint8_t(ttl)
	})
}

// SetLearn enables or disables learning of remote MAC addresses.
func SetLearn(name string, learn bool) error {
	return setCmd(name, constants.VXLAN_CMD_SET_LEARN, func(c *C.struct_ifvxlancmd) {
		if learn {
			c.vxlcmd_flags |= C.uint32_t(constants.VXLAN_CMD_FLAG_LEARN)
		}
	})
}

// SetFtableTimeout sets how long a learned forwarding table entry lives.
func SetFtableTimeout(name string, seconds uint32) error {
	return setCmd(name, constants.VXLAN_CMD_SET_FTABLE_TIMEO, func(c *C.struct_ifvxlancmd) {
		c.vxlcmd_ftable_timeout = C.uint32_t(seconds)
	})
}

// SetFtableMax sets the maximum number of forwarding table entries.
func SetFtableMax(name string, max uint32) error {
	return setCmd(name, constants.VXLAN_CMD_SET_FTABLE_MAX, func(c *C.struct_ifvxlancmd) {
		c.vxlcmd_ftable_max = C.uint32_t(max)
	})
}

// FtableAdd installs a static forwarding table entry: frames for mac are sent
// to the VTEP at remote:port.
//
// vxlan_ctrl_ftable_entry_add() requires the address family to match the
// interface's configured remote address, rejects unspecified and multicast
// addresses, and falls back to the interface's remote port when port is 0.
func FtableAdd(name string, mac net.HardwareAddr, remote net.IP, port uint16) error {
	if len(mac) != constants.EtherAddrLen {
		return isyscall.NewValidationError("mac", mac.String(),
			fmt.Sprintf("must be %d bytes", constants.EtherAddrLen))
	}

	var saErr error
	err := setCmd(name, constants.VXLAN_CMD_FTABLE_ENTRY_ADD, func(c *C.struct_ifvxlancmd) {
		isyscall.CopyBytes(unsafe.Pointer(&c.vxlcmd_mac[0]), unsafe.Pointer(&mac[0]),
			constants.EtherAddrLen)
		saErr = writeSockaddr(unsafe.Pointer(&c.vxlcmd_sa[0]), remote, port)
	})
	if saErr != nil {
		return saErr
	}
	return err
}

// FtableDel removes the forwarding table entry for mac. The kernel looks the
// entry up by MAC alone and answers ENOENT when there is none.
func FtableDel(name string, mac net.HardwareAddr) error {
	if len(mac) != constants.EtherAddrLen {
		return isyscall.NewValidationError("mac", mac.String(),
			fmt.Sprintf("must be %d bytes", constants.EtherAddrLen))
	}
	return setCmd(name, constants.VXLAN_CMD_FTABLE_ENTRY_REM, func(c *C.struct_ifvxlancmd) {
		isyscall.CopyBytes(unsafe.Pointer(&c.vxlcmd_mac[0]), unsafe.Pointer(&mac[0]),
			constants.EtherAddrLen)
	})
}

// Flush drops forwarding table entries. all also removes static entries;
// otherwise only dynamically learned ones go.
func Flush(name string, all bool) error {
	return setCmd(name, constants.VXLAN_CMD_FLUSH, func(c *C.struct_ifvxlancmd) {
		if all {
			c.vxlcmd_flags |= C.uint32_t(constants.VXLAN_CMD_FLAG_FLUSH_ALL)
		}
	})
}

// unitFromName extracts the clone unit from a default vxlan interface name.
func unitFromName(name string) (int, error) {
	const prefix = "vxlan"
	if !strings.HasPrefix(name, prefix) {
		return 0, isyscall.NewValidationError("name", name,
			"forwarding table dump needs the kernel-assigned vxlanN name")
	}
	unit, err := strconv.Atoi(name[len(prefix):])
	if err != nil || unit < 0 {
		return 0, isyscall.NewValidationError("name", name,
			"forwarding table dump needs the kernel-assigned vxlanN name")
	}
	return unit, nil
}

// FtableList returns the forwarding table of a vxlan interface.
//
// There is no ioctl for reading the table back; the only kernel interface is
// the per-unit sysctl net.link.vxlan.<unit>.ftable.dump registered by
// vxlan_sysctl_setup(). Two consequences follow, both from the kernel:
//
//   - The node is keyed by the clone unit, not the interface name, so this
//     only works while the interface still carries its kernel-assigned
//     "vxlanN" name.
//   - The dump is capped at one PAGE_SIZE sbuf. sys/net/if_vxlan.c says so
//     itself: "This is mostly intended for debugging during development. It is
//     not practical to dump an entire large table this way." A table larger
//     than a page comes back truncated at a line boundary.
//
// Use GetConfig().FtableCount for an exact, unconditional count.
func FtableList(name string) ([]FtableEntry, error) {
	unit, err := unitFromName(name)
	if err != nil {
		return nil, err
	}

	oid := fmt.Sprintf("net.link.vxlan.%d.ftable.dump", unit)
	coid := C.CString(oid)
	defer C.free(unsafe.Pointer(coid))

	var buf *C.char
	var length C.size_t
	if C.sysctl_string(coid, &buf, &length) < 0 {
		return nil, isyscall.MapError(isyscall.GetErrno())
	}
	defer C.free(unsafe.Pointer(buf))

	// Bound the string by the length the kernel reported rather than by a
	// NUL: sysctl_handle_string is not obliged to terminate the reply.
	return parseFtableDump(C.GoStringN(buf, C.int(length))), nil
}

// parseFtableDump decodes the text produced by vxlan_ftable_entry_dump():
//
//	<D|S> 0x%02X <mac> <address> <expire>
func parseFtableDump(dump string) []FtableEntry {
	entries := make([]FtableEntry, 0, 8)

	for _, line := range strings.Split(dump, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		if fields[0] != "S" && fields[0] != "D" {
			continue
		}
		mac, err := net.ParseMAC(fields[2])
		if err != nil {
			continue
		}
		ip := net.ParseIP(fields[3])
		if ip == nil {
			continue
		}
		var expire int64
		if len(fields) >= 5 {
			expire, _ = strconv.ParseInt(fields[4], 10, 64)
		}
		entries = append(entries, FtableEntry{
			MAC:    mac,
			Remote: ip,
			Static: fields[0] == "S",
			Expire: expire,
		})
	}

	return entries
}
