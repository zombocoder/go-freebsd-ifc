//go:build freebsd
// +build freebsd

package constants

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <net/if.h>
#include <net/if_vxlan.h>
#include <net/ethernet.h>
#include <netinet/in.h>
*/
import "C"

// VXLAN driver sub-commands, dispatched through SIOCSDRVSPEC/SIOCGDRVSPEC by
// vxlan_ioctl_drvspec() in sys/net/if_vxlan.c. GET_CONFIG is the only one the
// kernel marks COPYOUT, so it is the only one that goes through SIOCGDRVSPEC;
// every other command is COPYIN and requires PRIV_NET_VXLAN.
const (
	VXLAN_CMD_GET_CONFIG       = C.VXLAN_CMD_GET_CONFIG
	VXLAN_CMD_SET_VNI          = C.VXLAN_CMD_SET_VNI
	VXLAN_CMD_SET_LOCAL_ADDR   = C.VXLAN_CMD_SET_LOCAL_ADDR
	VXLAN_CMD_SET_REMOTE_ADDR  = C.VXLAN_CMD_SET_REMOTE_ADDR
	VXLAN_CMD_SET_LOCAL_PORT   = C.VXLAN_CMD_SET_LOCAL_PORT
	VXLAN_CMD_SET_REMOTE_PORT  = C.VXLAN_CMD_SET_REMOTE_PORT
	VXLAN_CMD_SET_PORT_RANGE   = C.VXLAN_CMD_SET_PORT_RANGE
	VXLAN_CMD_SET_FTABLE_TIMEO = C.VXLAN_CMD_SET_FTABLE_TIMEOUT
	VXLAN_CMD_SET_FTABLE_MAX   = C.VXLAN_CMD_SET_FTABLE_MAX
	VXLAN_CMD_SET_MULTICAST_IF = C.VXLAN_CMD_SET_MULTICAST_IF
	VXLAN_CMD_SET_TTL          = C.VXLAN_CMD_SET_TTL
	VXLAN_CMD_SET_LEARN        = C.VXLAN_CMD_SET_LEARN
	VXLAN_CMD_FTABLE_ENTRY_ADD = C.VXLAN_CMD_FTABLE_ENTRY_ADD
	VXLAN_CMD_FTABLE_ENTRY_REM = C.VXLAN_CMD_FTABLE_ENTRY_REM
	VXLAN_CMD_FLUSH            = C.VXLAN_CMD_FLUSH
)

// VXLAN command flags carried in ifvxlancmd.vxlcmd_flags.
const (
	VXLAN_CMD_FLAG_FLUSH_ALL = C.VXLAN_CMD_FLAG_FLUSH_ALL
	VXLAN_CMD_FLAG_LEARN     = C.VXLAN_CMD_FLAG_LEARN
)

// VXLAN protocol constants from <net/if_vxlan.h>.
const (
	// VXLAN_VNI_MAX is one past the largest usable VNI; vxlan_check_vni()
	// rejects anything >= this value.
	VXLAN_VNI_MAX = C.VXLAN_VNI_MAX
	// VXLAN_PORT is the IANA assigned UDP port.
	VXLAN_PORT = C.VXLAN_PORT
	// VXLAN_LEGACY_PORT is the port early Linux implementations used.
	VXLAN_LEGACY_PORT = C.VXLAN_LEGACY_PORT
)

// Structure sizes for the VXLAN driver interface.
const (
	SizeofIfvxlancfg = C.sizeof_struct_ifvxlancfg
	SizeofIfvxlancmd = C.sizeof_struct_ifvxlancmd
	SizeofIfdrv      = C.sizeof_struct_ifdrv
	EtherAddrLen     = C.ETHER_ADDR_LEN
)
