//go:build freebsd
// +build freebsd

package constants

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <net/route.h>
*/
import "C"

// Routing message types
const (
	RTM_ADD     = C.RTM_ADD
	RTM_DELETE  = C.RTM_DELETE
	RTM_VERSION = C.RTM_VERSION
)

// Routing flags
const (
	RTF_UP      = C.RTF_UP
	RTF_GATEWAY = C.RTF_GATEWAY
	RTF_HOST    = C.RTF_HOST
	RTF_STATIC  = C.RTF_STATIC
)

// Routing address types
const (
	RTA_DST     = C.RTA_DST
	RTA_GATEWAY = C.RTA_GATEWAY
	RTA_NETMASK = C.RTA_NETMASK
)
