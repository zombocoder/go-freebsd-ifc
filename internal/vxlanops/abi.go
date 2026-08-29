//go:build freebsd
// +build freebsd

package vxlanops

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/sockio.h>
#include <net/if.h>
#include <net/if_vxlan.h>
#include <net/ethernet.h>
#include <netinet/in.h>
#include <netinet/udp.h>
#include <netinet/ip.h>
#include <netinet/ip6.h>
*/
import "C"
import "unsafe"

// This file exposes the C ABI of the VXLAN driver structures to Go so the
// tests can pin it. It is not a _test.go file because the go tool does not
// allow cgo in test sources.

// Structure sizes from <net/if_vxlan.h> and <net/if.h>.
const (
	cSizeofIfdrv        = int(C.sizeof_struct_ifdrv)
	cSizeofIfvxlanparam = int(C.sizeof_struct_ifvxlanparam)
	cSizeofIfvxlancfg   = int(C.sizeof_struct_ifvxlancfg)
	cSizeofIfvxlancmd   = int(C.sizeof_struct_ifvxlancmd)
	cSizeofVxlanSa      = int(C.sizeof_union_vxlan_sockaddr)
	cSizeofSockaddrIn   = int(C.sizeof_struct_sockaddr_in)
	cSizeofSockaddrIn6  = int(C.sizeof_struct_sockaddr_in6)
)

// Header lengths that feed the kernel's own MTU accounting in
// vxlan_setup_interface_hdrlen().
const (
	cEtherHdrLen       = int(C.ETHER_HDR_LEN)
	cSizeofUdphdr      = int(C.sizeof_struct_udphdr)
	cSizeofVxlanHeader = int(C.sizeof_struct_vxlan_header)
	cSizeofIP          = int(C.sizeof_struct_ip)
	cSizeofIP6Hdr      = int(C.sizeof_struct_ip6_hdr)
	cEtherMTU          = int(C.ETHERMTU)
)

// Protocol constants.
const (
	cVxlanVniMax    = int(C.VXLAN_VNI_MAX)
	cVxlanPort      = int(C.VXLAN_PORT)
	cVxlanLegacy    = int(C.VXLAN_LEGACY_PORT)
	cEtherAddrLen   = int(C.ETHER_ADDR_LEN)
	cIfnamsiz       = int(C.IFNAMSIZ)
	cSiocsdrvspec   = uint64(C.SIOCSDRVSPEC)
	cSiocgdrvspec   = uint64(C.SIOCGDRVSPEC)
	cFlagFlushAll   = int(C.VXLAN_CMD_FLAG_FLUSH_ALL)
	cFlagLearn      = int(C.VXLAN_CMD_FLAG_LEARN)
	cCmdGetConfig   = int(C.VXLAN_CMD_GET_CONFIG)
	cCmdSetVNI      = int(C.VXLAN_CMD_SET_VNI)
	cCmdSetLocalA   = int(C.VXLAN_CMD_SET_LOCAL_ADDR)
	cCmdSetRemoteA  = int(C.VXLAN_CMD_SET_REMOTE_ADDR)
	cCmdSetLocalP   = int(C.VXLAN_CMD_SET_LOCAL_PORT)
	cCmdSetRemoteP  = int(C.VXLAN_CMD_SET_REMOTE_PORT)
	cCmdSetRange    = int(C.VXLAN_CMD_SET_PORT_RANGE)
	cCmdSetTimeout  = int(C.VXLAN_CMD_SET_FTABLE_TIMEOUT)
	cCmdSetMax      = int(C.VXLAN_CMD_SET_FTABLE_MAX)
	cCmdSetMcastIf  = int(C.VXLAN_CMD_SET_MULTICAST_IF)
	cCmdSetTTL      = int(C.VXLAN_CMD_SET_TTL)
	cCmdSetLearn    = int(C.VXLAN_CMD_SET_LEARN)
	cCmdFtableAdd   = int(C.VXLAN_CMD_FTABLE_ENTRY_ADD)
	cCmdFtableRem   = int(C.VXLAN_CMD_FTABLE_ENTRY_REM)
	cCmdFlush       = int(C.VXLAN_CMD_FLUSH)
	cVxlanHdrVniSh  = int(C.VXLAN_HDR_VNI_SHIFT)
	cVxlanHdrValid  = int(C.VXLAN_HDR_FLAGS_VALID_VNI)
	cParamWithVNI   = int(C.VXLAN_PARAM_WITH_VNI)
	cParamWithLocal = int(C.VXLAN_PARAM_WITH_LOCAL_ADDR4)
)

// cIfdrvOffsets holds offsetof() for struct ifdrv.
var cIfdrvOffsets = func() map[string]uintptr {
	var d C.struct_ifdrv
	return map[string]uintptr{
		"ifd_name": unsafe.Offsetof(d.ifd_name),
		"ifd_cmd":  unsafe.Offsetof(d.ifd_cmd),
		"ifd_len":  unsafe.Offsetof(d.ifd_len),
		"ifd_data": unsafe.Offsetof(d.ifd_data),
	}
}()

// cIfvxlanparamOffsets holds offsetof() for struct ifvxlanparam.
var cIfvxlanparamOffsets = func() map[string]uintptr {
	var p C.struct_ifvxlanparam
	return map[string]uintptr{
		"vxlp_with":           unsafe.Offsetof(p.vxlp_with),
		"vxlp_vni":            unsafe.Offsetof(p.vxlp_vni),
		"vxlp_local_sa":       unsafe.Offsetof(p.vxlp_local_sa),
		"vxlp_remote_sa":      unsafe.Offsetof(p.vxlp_remote_sa),
		"vxlp_local_port":     unsafe.Offsetof(p.vxlp_local_port),
		"vxlp_remote_port":    unsafe.Offsetof(p.vxlp_remote_port),
		"vxlp_min_port":       unsafe.Offsetof(p.vxlp_min_port),
		"vxlp_max_port":       unsafe.Offsetof(p.vxlp_max_port),
		"vxlp_mc_ifname":      unsafe.Offsetof(p.vxlp_mc_ifname),
		"vxlp_ftable_timeout": unsafe.Offsetof(p.vxlp_ftable_timeout),
		"vxlp_ftable_max":     unsafe.Offsetof(p.vxlp_ftable_max),
		"vxlp_ttl":            unsafe.Offsetof(p.vxlp_ttl),
		"vxlp_learn":          unsafe.Offsetof(p.vxlp_learn),
	}
}()

// cIfvxlancfgOffsets holds offsetof() for struct ifvxlancfg.
var cIfvxlancfgOffsets = func() map[string]uintptr {
	var c C.struct_ifvxlancfg
	return map[string]uintptr{
		"vxlc_vni":            unsafe.Offsetof(c.vxlc_vni),
		"vxlc_local_sa":       unsafe.Offsetof(c.vxlc_local_sa),
		"vxlc_remote_sa":      unsafe.Offsetof(c.vxlc_remote_sa),
		"vxlc_mc_ifindex":     unsafe.Offsetof(c.vxlc_mc_ifindex),
		"vxlc_ftable_cnt":     unsafe.Offsetof(c.vxlc_ftable_cnt),
		"vxlc_ftable_max":     unsafe.Offsetof(c.vxlc_ftable_max),
		"vxlc_ftable_timeout": unsafe.Offsetof(c.vxlc_ftable_timeout),
		"vxlc_port_min":       unsafe.Offsetof(c.vxlc_port_min),
		"vxlc_port_max":       unsafe.Offsetof(c.vxlc_port_max),
		"vxlc_learn":          unsafe.Offsetof(c.vxlc_learn),
		"vxlc_ttl":            unsafe.Offsetof(c.vxlc_ttl),
	}
}()

// cIfvxlancmdOffsets holds offsetof() for struct ifvxlancmd.
var cIfvxlancmdOffsets = func() map[string]uintptr {
	var c C.struct_ifvxlancmd
	return map[string]uintptr{
		"vxlcmd_flags":          unsafe.Offsetof(c.vxlcmd_flags),
		"vxlcmd_vni":            unsafe.Offsetof(c.vxlcmd_vni),
		"vxlcmd_ftable_timeout": unsafe.Offsetof(c.vxlcmd_ftable_timeout),
		"vxlcmd_ftable_max":     unsafe.Offsetof(c.vxlcmd_ftable_max),
		"vxlcmd_port":           unsafe.Offsetof(c.vxlcmd_port),
		"vxlcmd_port_min":       unsafe.Offsetof(c.vxlcmd_port_min),
		"vxlcmd_port_max":       unsafe.Offsetof(c.vxlcmd_port_max),
		"vxlcmd_mac":            unsafe.Offsetof(c.vxlcmd_mac),
		"vxlcmd_ttl":            unsafe.Offsetof(c.vxlcmd_ttl),
		"vxlcmd_sa":             unsafe.Offsetof(c.vxlcmd_sa),
		"vxlcmd_ifname":         unsafe.Offsetof(c.vxlcmd_ifname),
	}
}()

// cIfvxlancmdSizes holds sizeof() for each field of struct ifvxlancmd.
var cIfvxlancmdSizes = func() map[string]uintptr {
	var c C.struct_ifvxlancmd
	return map[string]uintptr{
		"vxlcmd_flags":          unsafe.Sizeof(c.vxlcmd_flags),
		"vxlcmd_vni":            unsafe.Sizeof(c.vxlcmd_vni),
		"vxlcmd_ftable_timeout": unsafe.Sizeof(c.vxlcmd_ftable_timeout),
		"vxlcmd_ftable_max":     unsafe.Sizeof(c.vxlcmd_ftable_max),
		"vxlcmd_port":           unsafe.Sizeof(c.vxlcmd_port),
		"vxlcmd_port_min":       unsafe.Sizeof(c.vxlcmd_port_min),
		"vxlcmd_port_max":       unsafe.Sizeof(c.vxlcmd_port_max),
		"vxlcmd_mac":            unsafe.Sizeof(c.vxlcmd_mac),
		"vxlcmd_ttl":            unsafe.Sizeof(c.vxlcmd_ttl),
		"vxlcmd_sa":             unsafe.Sizeof(c.vxlcmd_sa),
		"vxlcmd_ifname":         unsafe.Sizeof(c.vxlcmd_ifname),
	}
}()

// cIfvxlancfgSizes holds sizeof() for each field of struct ifvxlancfg.
var cIfvxlancfgSizes = func() map[string]uintptr {
	var c C.struct_ifvxlancfg
	return map[string]uintptr{
		"vxlc_vni":            unsafe.Sizeof(c.vxlc_vni),
		"vxlc_local_sa":       unsafe.Sizeof(c.vxlc_local_sa),
		"vxlc_remote_sa":      unsafe.Sizeof(c.vxlc_remote_sa),
		"vxlc_mc_ifindex":     unsafe.Sizeof(c.vxlc_mc_ifindex),
		"vxlc_ftable_cnt":     unsafe.Sizeof(c.vxlc_ftable_cnt),
		"vxlc_ftable_max":     unsafe.Sizeof(c.vxlc_ftable_max),
		"vxlc_ftable_timeout": unsafe.Sizeof(c.vxlc_ftable_timeout),
		"vxlc_port_min":       unsafe.Sizeof(c.vxlc_port_min),
		"vxlc_port_max":       unsafe.Sizeof(c.vxlc_port_max),
		"vxlc_learn":          unsafe.Sizeof(c.vxlc_learn),
		"vxlc_ttl":            unsafe.Sizeof(c.vxlc_ttl),
	}
}()
