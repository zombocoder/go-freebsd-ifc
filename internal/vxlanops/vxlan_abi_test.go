//go:build freebsd
// +build freebsd

package vxlanops

import (
	"net"
	"testing"

	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
)

// The reference numbers below come from a C probe compiled against the system
// headers on FreeBSD 14.3-RELEASE-p3/amd64. See docs/fixes-report.md for the
// full probe output.
//
//	sizeof(struct ifdrv)          =  40
//	sizeof(union vxlan_sockaddr)  =  28
//	sizeof(struct ifvxlanparam)   = 104
//	sizeof(struct ifvxlancfg)     =  84
//	sizeof(struct ifvxlancmd)     =  76

func TestStructSizes(t *testing.T) {
	cases := []struct {
		name  string
		got   int
		probe int
	}{
		{"struct ifdrv", cSizeofIfdrv, 40},
		{"union vxlan_sockaddr", cSizeofVxlanSa, 28},
		{"struct ifvxlanparam", cSizeofIfvxlanparam, 104},
		{"struct ifvxlancfg", cSizeofIfvxlancfg, 84},
		{"struct ifvxlancmd", cSizeofIfvxlancmd, 76},
		{"struct sockaddr_in", cSizeofSockaddrIn, 16},
		{"struct sockaddr_in6", cSizeofSockaddrIn6, 28},
	}
	for _, tc := range cases {
		if tc.got != tc.probe {
			t.Errorf("%s: cgo sizeof = %d, C-probe sizeof = %d", tc.name, tc.got, tc.probe)
		}
	}

	// The sizes the ioctl path actually sends. vxlan_ioctl_drvspec() rejects
	// the request with EINVAL unless ifd_len equals the vxlc_argsize the
	// control table declares for the command.
	if constants.SizeofIfvxlancfg != cSizeofIfvxlancfg {
		t.Errorf("constants.SizeofIfvxlancfg = %d, want %d",
			constants.SizeofIfvxlancfg, cSizeofIfvxlancfg)
	}
	if constants.SizeofIfvxlancmd != cSizeofIfvxlancmd {
		t.Errorf("constants.SizeofIfvxlancmd = %d, want %d",
			constants.SizeofIfvxlancmd, cSizeofIfvxlancmd)
	}
	if constants.SizeofIfdrv != cSizeofIfdrv {
		t.Errorf("constants.SizeofIfdrv = %d, want %d", constants.SizeofIfdrv, cSizeofIfdrv)
	}
}

func TestIfdrvOffsets(t *testing.T) {
	want := map[string]uintptr{
		"ifd_name": 0,
		"ifd_cmd":  16,
		"ifd_len":  24,
		"ifd_data": 32,
	}
	checkOffsets(t, "struct ifdrv", cIfdrvOffsets, want)
}

func TestIfvxlanparamOffsets(t *testing.T) {
	want := map[string]uintptr{
		"vxlp_with":           0,
		"vxlp_vni":            8,
		"vxlp_local_sa":       12,
		"vxlp_remote_sa":      40,
		"vxlp_local_port":     68,
		"vxlp_remote_port":    70,
		"vxlp_min_port":       72,
		"vxlp_max_port":       74,
		"vxlp_mc_ifname":      76,
		"vxlp_ftable_timeout": 92,
		"vxlp_ftable_max":     96,
		"vxlp_ttl":            100,
		"vxlp_learn":          101,
	}
	checkOffsets(t, "struct ifvxlanparam", cIfvxlanparamOffsets, want)
}

func TestIfvxlancfgOffsets(t *testing.T) {
	want := map[string]uintptr{
		"vxlc_vni":            0,
		"vxlc_local_sa":       4,
		"vxlc_remote_sa":      32,
		"vxlc_mc_ifindex":     60,
		"vxlc_ftable_cnt":     64,
		"vxlc_ftable_max":     68,
		"vxlc_ftable_timeout": 72,
		"vxlc_port_min":       76,
		"vxlc_port_max":       78,
		"vxlc_learn":          80,
		"vxlc_ttl":            81,
	}
	checkOffsets(t, "struct ifvxlancfg", cIfvxlancfgOffsets, want)

	wantSizes := map[string]uintptr{
		"vxlc_vni":            4,
		"vxlc_local_sa":       28,
		"vxlc_remote_sa":      28,
		"vxlc_mc_ifindex":     4,
		"vxlc_ftable_cnt":     4,
		"vxlc_ftable_max":     4,
		"vxlc_ftable_timeout": 4,
		"vxlc_port_min":       2,
		"vxlc_port_max":       2,
		"vxlc_learn":          1,
		"vxlc_ttl":            1,
	}
	checkOffsets(t, "struct ifvxlancfg sizes", cIfvxlancfgSizes, wantSizes)
}

func TestIfvxlancmdOffsets(t *testing.T) {
	want := map[string]uintptr{
		"vxlcmd_flags":          0,
		"vxlcmd_vni":            4,
		"vxlcmd_ftable_timeout": 8,
		"vxlcmd_ftable_max":     12,
		"vxlcmd_port":           16,
		"vxlcmd_port_min":       18,
		"vxlcmd_port_max":       20,
		"vxlcmd_mac":            22,
		"vxlcmd_ttl":            28,
		"vxlcmd_sa":             32,
		"vxlcmd_ifname":         60,
	}
	checkOffsets(t, "struct ifvxlancmd", cIfvxlancmdOffsets, want)

	wantSizes := map[string]uintptr{
		"vxlcmd_flags":          4,
		"vxlcmd_vni":            4,
		"vxlcmd_ftable_timeout": 4,
		"vxlcmd_ftable_max":     4,
		"vxlcmd_port":           2,
		"vxlcmd_port_min":       2,
		"vxlcmd_port_max":       2,
		"vxlcmd_mac":            6,
		"vxlcmd_ttl":            1,
		"vxlcmd_sa":             28,
		"vxlcmd_ifname":         16,
	}
	checkOffsets(t, "struct ifvxlancmd sizes", cIfvxlancmdSizes, wantSizes)
}

// TestVxlanConstants pins the driver command numbers and protocol constants.
func TestVxlanConstants(t *testing.T) {
	cases := []struct {
		name       string
		got, probe int
	}{
		{"VXLAN_VNI_MAX", cVxlanVniMax, 1 << 24},
		{"VXLAN_PORT", cVxlanPort, 4789},
		{"VXLAN_LEGACY_PORT", cVxlanLegacy, 8472},
		{"ETHER_ADDR_LEN", cEtherAddrLen, 6},
		{"IFNAMSIZ", cIfnamsiz, 16},
		{"VXLAN_HDR_VNI_SHIFT", cVxlanHdrVniSh, 8},
		{"VXLAN_HDR_FLAGS_VALID_VNI", cVxlanHdrValid, 0x08000000},
		{"VXLAN_PARAM_WITH_VNI", cParamWithVNI, 0x0001},
		{"VXLAN_PARAM_WITH_LOCAL_ADDR4", cParamWithLocal, 0x0002},
		{"VXLAN_CMD_FLAG_FLUSH_ALL", cFlagFlushAll, 0x0001},
		{"VXLAN_CMD_FLAG_LEARN", cFlagLearn, 0x0002},
		{"VXLAN_CMD_GET_CONFIG", cCmdGetConfig, 0},
		{"VXLAN_CMD_SET_VNI", cCmdSetVNI, 1},
		{"VXLAN_CMD_SET_LOCAL_ADDR", cCmdSetLocalA, 2},
		{"VXLAN_CMD_SET_REMOTE_ADDR", cCmdSetRemoteA, 4},
		{"VXLAN_CMD_SET_LOCAL_PORT", cCmdSetLocalP, 5},
		{"VXLAN_CMD_SET_REMOTE_PORT", cCmdSetRemoteP, 6},
		{"VXLAN_CMD_SET_PORT_RANGE", cCmdSetRange, 7},
		{"VXLAN_CMD_SET_FTABLE_TIMEOUT", cCmdSetTimeout, 8},
		{"VXLAN_CMD_SET_FTABLE_MAX", cCmdSetMax, 9},
		{"VXLAN_CMD_SET_MULTICAST_IF", cCmdSetMcastIf, 10},
		{"VXLAN_CMD_SET_TTL", cCmdSetTTL, 11},
		{"VXLAN_CMD_SET_LEARN", cCmdSetLearn, 12},
		{"VXLAN_CMD_FTABLE_ENTRY_ADD", cCmdFtableAdd, 13},
		{"VXLAN_CMD_FTABLE_ENTRY_REM", cCmdFtableRem, 14},
		{"VXLAN_CMD_FLUSH", cCmdFlush, 15},
	}
	for _, tc := range cases {
		if tc.got != tc.probe {
			t.Errorf("%s: cgo = %d, C-probe = %d", tc.name, tc.got, tc.probe)
		}
	}

	// The ioctl numbers vxlan_ioctl() dispatches on, encoding
	// sizeof(struct ifdrv) == 40 == 0x28 in the length field.
	if cSiocsdrvspec != 0x8028697b {
		t.Errorf("SIOCSDRVSPEC = %#x, want 0x8028697b", cSiocsdrvspec)
	}
	if cSiocgdrvspec != 0xc028697b {
		t.Errorf("SIOCGDRVSPEC = %#x, want 0xc028697b", cSiocgdrvspec)
	}

	// The constants package must agree with the headers.
	pairs := []struct {
		name      string
		got, want int
	}{
		{"VXLAN_CMD_GET_CONFIG", constants.VXLAN_CMD_GET_CONFIG, cCmdGetConfig},
		{"VXLAN_CMD_SET_VNI", constants.VXLAN_CMD_SET_VNI, cCmdSetVNI},
		{"VXLAN_CMD_SET_LOCAL_ADDR", constants.VXLAN_CMD_SET_LOCAL_ADDR, cCmdSetLocalA},
		{"VXLAN_CMD_SET_REMOTE_ADDR", constants.VXLAN_CMD_SET_REMOTE_ADDR, cCmdSetRemoteA},
		{"VXLAN_CMD_SET_LOCAL_PORT", constants.VXLAN_CMD_SET_LOCAL_PORT, cCmdSetLocalP},
		{"VXLAN_CMD_SET_REMOTE_PORT", constants.VXLAN_CMD_SET_REMOTE_PORT, cCmdSetRemoteP},
		{"VXLAN_CMD_SET_PORT_RANGE", constants.VXLAN_CMD_SET_PORT_RANGE, cCmdSetRange},
		{"VXLAN_CMD_FTABLE_ENTRY_ADD", constants.VXLAN_CMD_FTABLE_ENTRY_ADD, cCmdFtableAdd},
		{"VXLAN_CMD_FTABLE_ENTRY_REM", constants.VXLAN_CMD_FTABLE_ENTRY_REM, cCmdFtableRem},
		{"VXLAN_CMD_FLUSH", constants.VXLAN_CMD_FLUSH, cCmdFlush},
		{"VXLAN_VNI_MAX", constants.VXLAN_VNI_MAX, cVxlanVniMax},
		{"VXLAN_PORT", constants.VXLAN_PORT, cVxlanPort},
		{"EtherAddrLen", constants.EtherAddrLen, cEtherAddrLen},
	}
	for _, p := range pairs {
		if p.got != p.want {
			t.Errorf("constants.%s = %d, want %d", p.name, p.got, p.want)
		}
	}
}

// TestEncapOverhead pins the kernel's own MTU accounting from
// vxlan_setup_interface_hdrlen():
//
//	if_hdrlen = ETHER_HDR_LEN + sizeof(struct vxlanudphdr) [+ outer IP]
//	if_mtu    = ETHERMTU - if_hdrlen
func TestEncapOverhead(t *testing.T) {
	vxlanudphdr := cSizeofUdphdr + cSizeofVxlanHeader

	if cEtherHdrLen != 14 {
		t.Errorf("ETHER_HDR_LEN = %d, want 14", cEtherHdrLen)
	}
	if cSizeofUdphdr != 8 {
		t.Errorf("sizeof(struct udphdr) = %d, want 8", cSizeofUdphdr)
	}
	if cSizeofVxlanHeader != 8 {
		t.Errorf("sizeof(struct vxlan_header) = %d, want 8", cSizeofVxlanHeader)
	}
	if cSizeofIP != 20 {
		t.Errorf("sizeof(struct ip) = %d, want 20", cSizeofIP)
	}
	if cSizeofIP6Hdr != 40 {
		t.Errorf("sizeof(struct ip6_hdr) = %d, want 40", cSizeofIP6Hdr)
	}

	if got, want := cEtherHdrLen+vxlanudphdr+cSizeofIP, 50; got != want {
		t.Errorf("IPv4 if_hdrlen = %d, want %d", got, want)
	}
	if got, want := cEtherHdrLen+vxlanudphdr+cSizeofIP6Hdr, 70; got != want {
		t.Errorf("IPv6 if_hdrlen = %d, want %d", got, want)
	}
	if got, want := cEtherMTU-(cEtherHdrLen+vxlanudphdr+cSizeofIP), 1450; got != want {
		t.Errorf("default IPv4 MTU = %d, want %d", got, want)
	}
}

func TestParseFtableDump(t *testing.T) {
	// Exactly the layout vxlan_ftable_entry_dump() produces:
	//   "%c 0x%02X " then the MAC, then a right-aligned address, then
	//   "%08jd" for the expiry.
	dump := "\n" +
		"S 0x02 02:11:22:33:44:55         192.0.2.10 00000000\n" +
		"D 0x01 02:AA:BB:CC:DD:EE         192.0.2.11 00001234\n" +
		"S 0x02 02:11:22:33:44:66  2001:db8::1 00000000\n"

	got := parseFtableDump(dump)
	if len(got) != 3 {
		t.Fatalf("parsed %d entries, want 3: %#v", len(got), got)
	}

	if got[0].MAC.String() != "02:11:22:33:44:55" {
		t.Errorf("entry 0 MAC = %s", got[0].MAC)
	}
	if !got[0].Remote.Equal(net.ParseIP("192.0.2.10")) {
		t.Errorf("entry 0 remote = %s", got[0].Remote)
	}
	if !got[0].Static {
		t.Error("entry 0 should be static")
	}
	if got[1].Static {
		t.Error("entry 1 should be dynamic")
	}
	if got[1].Expire != 1234 {
		t.Errorf("entry 1 expire = %d, want 1234", got[1].Expire)
	}
	if !got[2].Remote.Equal(net.ParseIP("2001:db8::1")) {
		t.Errorf("entry 2 remote = %s", got[2].Remote)
	}
}

func TestByteOrderHelpers(t *testing.T) {
	if got := htons(4789); got != 0xb512 {
		t.Errorf("htons(4789) = %#x, want 0xb512", got)
	}
	if got := ntohs(htons(4789)); got != 4789 {
		t.Errorf("ntohs(htons(4789)) = %d, want 4789", got)
	}
}

func TestUnitFromName(t *testing.T) {
	if u, err := unitFromName("vxlan7"); err != nil || u != 7 {
		t.Errorf("unitFromName(vxlan7) = %d, %v; want 7, nil", u, err)
	}
	for _, bad := range []string{"tap0", "vxlan", "vxlanx", "", "vxlan-1"} {
		if _, err := unitFromName(bad); err == nil {
			t.Errorf("unitFromName(%q) should fail", bad)
		}
	}
}

func checkOffsets(t *testing.T, what string, got, want map[string]uintptr) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("%s: %d fields measured, %d expected", what, len(got), len(want))
	}
	for field, wantOff := range want {
		gotOff, ok := got[field]
		if !ok {
			t.Errorf("%s: %s not measured", what, field)
			continue
		}
		if gotOff != wantOff {
			t.Errorf("%s: %s = %d, C-probe says %d", what, field, gotOff, wantOff)
		}
	}
}
