//go:build freebsd
// +build freebsd

package routing

import (
	"testing"
	"unsafe"
)

// The reference numbers below come from a C probe compiled against the system
// headers on FreeBSD 14.3-RELEASE-p3/amd64:
//
//	sizeof(struct rt_msghdr)   = 152
//	sizeof(struct rt_metrics)  = 112
//	rtm_msglen  off=  0  rtm_version off=  2  rtm_type  off=  3
//	rtm_index   off=  4  _rtm_spare1 off=  6  rtm_flags off=  8
//	rtm_addrs   off= 12  rtm_pid     off= 16  rtm_seq   off= 20
//	rtm_errno   off= 24  rtm_fmask   off= 28  rtm_inits off= 32 (size 8)
//	rtm_rmx     off= 40  (size 112)
const (
	probeSizeofRtMsghdr  = 152
	probeSizeofRtMetrics = 112
)

// TestRtMsghdrSize pins the wire size of rtMsghdr.
//
// sys/net/rtsock.c:1076-1079 rejects any routing message whose length is
// smaller than sizeof(struct rt_msghdr) or whose rtm_msglen disagrees with the
// write() length, so a Go struct of the wrong size makes every RTM_ADD and
// RTM_DELETE fail with EINVAL before the kernel ever parses it.
func TestRtMsghdrSize(t *testing.T) {
	if got := int(unsafe.Sizeof(rtMsghdr{})); got != probeSizeofRtMsghdr {
		t.Errorf("unsafe.Sizeof(rtMsghdr{}) = %d, want %d", got, probeSizeofRtMsghdr)
	}
	if sizeofRtMsghdr != probeSizeofRtMsghdr {
		t.Errorf("sizeofRtMsghdr = %d, want %d", sizeofRtMsghdr, probeSizeofRtMsghdr)
	}
	if cSizeofRtMsghdr != probeSizeofRtMsghdr {
		t.Errorf("C sizeof(struct rt_msghdr) = %d, want %d (system header changed?)",
			cSizeofRtMsghdr, probeSizeofRtMsghdr)
	}
	if got := int(unsafe.Sizeof(rtMsghdr{})); got != cSizeofRtMsghdr {
		t.Errorf("Go rtMsghdr is %d bytes, C struct rt_msghdr is %d bytes", got, cSizeofRtMsghdr)
	}
	if got := len((&rtMsghdr{}).bytes()); got != probeSizeofRtMsghdr {
		t.Errorf("rtMsghdr.bytes() length = %d, want %d", got, probeSizeofRtMsghdr)
	}
}

// TestRtMsghdrOffsets pins every field offset against the C header. Two structs
// can have the same size and still disagree field by field, so size alone is
// not enough.
func TestRtMsghdrOffsets(t *testing.T) {
	var h rtMsghdr

	cases := []struct {
		cName string
		got   uintptr
		probe uintptr
	}{
		{"rtm_msglen", unsafe.Offsetof(h.msglen), 0},
		{"rtm_version", unsafe.Offsetof(h.version), 2},
		{"rtm_type", unsafe.Offsetof(h.msgtype), 3},
		{"rtm_index", unsafe.Offsetof(h.index), 4},
		{"rtm_flags", unsafe.Offsetof(h.flags), 8},
		{"rtm_addrs", unsafe.Offsetof(h.addrs), 12},
		{"rtm_pid", unsafe.Offsetof(h.pid), 16},
		{"rtm_seq", unsafe.Offsetof(h.seq), 20},
		{"rtm_errno", unsafe.Offsetof(h.errno), 24},
		{"rtm_fmask", unsafe.Offsetof(h.fmask), 28},
		{"rtm_inits", unsafe.Offsetof(h.inits), 32},
		{"rtm_rmx", unsafe.Offsetof(h.rmx), 40},
	}

	for _, tc := range cases {
		cOff, ok := cRtMsghdrOffsets[tc.cName]
		if !ok {
			t.Fatalf("%s: missing from cRtMsghdrOffsets", tc.cName)
		}
		if tc.got != cOff {
			t.Errorf("%s: Go offset %d != C offset %d", tc.cName, tc.got, cOff)
		}
		if tc.got != tc.probe {
			t.Errorf("%s: Go offset %d != C-probe offset %d", tc.cName, tc.got, tc.probe)
		}
	}

	// _rtm_spare1 sits between rtm_index and rtm_flags. cgo gives the field
	// no usable Go name, so it is pinned directly: without it rtm_flags
	// still lands on offset 8 by alignment, but the message would be two
	// bytes short of the kernel's idea of the header.
	if got, want := unsafe.Offsetof(h.spare1), uintptr(6); got != want {
		t.Errorf("_rtm_spare1 offset = %d, want %d", got, want)
	}
	if got, want := unsafe.Sizeof(h.spare1), uintptr(2); got != want {
		t.Errorf("_rtm_spare1 size = %d, want %d", got, want)
	}
}

// TestRtMetricsSize pins struct rt_metrics: every member is u_long, so the
// trailing rmx array is 112 bytes on LP64, not 56, and rtm_inits is 8 bytes,
// not 4.
func TestRtMetricsSize(t *testing.T) {
	var h rtMsghdr

	if got := int(unsafe.Sizeof(h.rmx)); got != probeSizeofRtMetrics {
		t.Errorf("rmx size = %d, want %d", got, probeSizeofRtMetrics)
	}
	if cSizeofRtMetrics != probeSizeofRtMetrics {
		t.Errorf("C sizeof(struct rt_metrics) = %d, want %d", cSizeofRtMetrics, probeSizeofRtMetrics)
	}
	if got := int(unsafe.Sizeof(h.rmx)); got != cSizeofRtMetrics {
		t.Errorf("Go rmx is %d bytes, C struct rt_metrics is %d bytes", got, cSizeofRtMetrics)
	}
	if got := int(unsafe.Sizeof(h.inits)); got != cSizeofULong {
		t.Errorf("rtm_inits is %d bytes in Go, u_long is %d bytes", got, cSizeofULong)
	}
}

// TestSaSize pins the SA_SIZE() rounding used to walk sockaddr arrays.
func TestSaSize(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, 8}, {1, 8}, {7, 8}, {8, 8}, {9, 16}, {16, 16}, {17, 24}, {28, 32},
	}
	for _, tc := range cases {
		if got := saSize(tc.in); got != tc.want {
			t.Errorf("saSize(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
