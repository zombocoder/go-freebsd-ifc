//go:build freebsd
// +build freebsd

package routing

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <net/route.h>
*/
import "C"
import "unsafe"

// This file exposes the C ABI of struct rt_msghdr to Go so that the tests can
// compare it against the hand-written rtMsghdr in routing.go. It lives outside
// the _test.go files because the go tool does not allow cgo in test sources.

// Sizes taken straight from <net/route.h>.
const (
	cSizeofRtMsghdr  = int(C.sizeof_struct_rt_msghdr)
	cSizeofRtMetrics = int(C.sizeof_struct_rt_metrics)
	cSizeofULong     = int(C.sizeof_u_long)
)

// cRtMsghdrOffsets holds offsetof() for every named field of struct rt_msghdr.
var cRtMsghdrOffsets = func() map[string]uintptr {
	var h C.struct_rt_msghdr
	return map[string]uintptr{
		"rtm_msglen":  unsafe.Offsetof(h.rtm_msglen),
		"rtm_version": unsafe.Offsetof(h.rtm_version),
		"rtm_type":    unsafe.Offsetof(h.rtm_type),
		"rtm_index":   unsafe.Offsetof(h.rtm_index),
		"rtm_flags":   unsafe.Offsetof(h.rtm_flags),
		"rtm_addrs":   unsafe.Offsetof(h.rtm_addrs),
		"rtm_pid":     unsafe.Offsetof(h.rtm_pid),
		"rtm_seq":     unsafe.Offsetof(h.rtm_seq),
		"rtm_errno":   unsafe.Offsetof(h.rtm_errno),
		"rtm_fmask":   unsafe.Offsetof(h.rtm_fmask),
		"rtm_inits":   unsafe.Offsetof(h.rtm_inits),
		"rtm_rmx":     unsafe.Offsetof(h.rtm_rmx),
	}
}()
