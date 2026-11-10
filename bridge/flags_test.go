//go:build freebsd
// +build freebsd

package bridge

import (
	"strings"
	"testing"
)

// The values are the kernel ABI (net/if_bridgevar.h).
func TestMemberFlagValues(t *testing.T) {
	cases := []struct {
		name string
		got  MemberFlag
		want uint32
	}{
		{"Learning", FlagLearning, 0x0001},
		{"Discover", FlagDiscover, 0x0002},
		{"STP", FlagSTP, 0x0004},
		{"Span", FlagSpan, 0x0008},
		{"Sticky", FlagSticky, 0x0010},
		{"BSTPEdge", FlagBSTPEdge, 0x0020},
		{"BSTPAutoEdge", FlagBSTPAutoEdge, 0x0040},
		{"BSTPPTP", FlagBSTPPTP, 0x0080},
		{"BSTPAutoPTP", FlagBSTPAutoPTP, 0x0100},
		{"BSTPAdmEdge", FlagBSTPAdmEdge, 0x0200},
		{"BSTPAdmCost", FlagBSTPAdmCost, 0x0400},
		{"Private", FlagPrivate, 0x0800},
		{"QinQ", FlagQinQ, 0x1000},
	}
	for _, c := range cases {
		if uint32(c.got) != c.want {
			t.Errorf("Flag%s = %#x, want %#x", c.name, uint32(c.got), c.want)
		}
	}
}

func TestMemberFlagString(t *testing.T) {
	cases := []struct {
		f    MemberFlag
		want string
	}{
		{0, "0"},
		{FlagLearning | FlagDiscover, "LEARNING,DISCOVER"},
		{FlagPrivate, "PRIVATE"},
		{FlagLearning | FlagDiscover | FlagPrivate, "LEARNING,DISCOVER,PRIVATE"},
		{FlagSTP | FlagSpan | FlagSticky, "STP,SPAN,STICKY"},
		{MemberFlag(0x80000000), "0x80000000"},
		{FlagPrivate | MemberFlag(0x80000000), "PRIVATE,0x80000000"},
	}
	for _, c := range cases {
		if got := c.f.String(); got != c.want {
			t.Errorf("MemberFlag(%#x).String() = %q, want %q", uint32(c.f), got, c.want)
		}
	}
}

func TestMemberFlagHas(t *testing.T) {
	f := FlagLearning | FlagPrivate
	if !f.Has(FlagPrivate) || !f.Has(FlagLearning|FlagPrivate) {
		t.Error("Has should report set flags")
	}
	if f.Has(FlagSTP) || f.Has(FlagPrivate|FlagSTP) {
		t.Error("Has should require every requested flag")
	}
}

func TestMemberFlagsValidation(t *testing.T) {
	long := strings.Repeat("x", 16)
	if _, err := MemberFlags("", "tap0"); err == nil {
		t.Error("MemberFlags with empty bridge should fail")
	}
	if _, err := MemberFlags("bridge0", ""); err == nil {
		t.Error("MemberFlags with empty member should fail")
	}
	if _, err := MemberFlags(long, "tap0"); err == nil {
		t.Error("MemberFlags with over-long bridge should fail")
	}
	if _, err := MemberFlags("bridge0", long); err == nil {
		t.Error("MemberFlags with over-long member should fail")
	}
	if err := SetMemberFlag(long, "tap0", FlagPrivate, true); err == nil {
		t.Error("SetMemberFlag with over-long bridge should fail")
	}
	if err := SetMemberFlag("bridge0", long, FlagPrivate, true); err == nil {
		t.Error("SetMemberFlag with over-long member should fail")
	}
	if err := SetMemberFlag("bridge0", "tap0", 0, true); err == nil {
		t.Error("SetMemberFlag with no flag should fail")
	}
	if err := SetMemberPrivate(long, "tap0", true); err == nil {
		t.Error("SetMemberPrivate with over-long bridge should fail")
	}
	if _, err := MemberPrivate("bridge0", long); err == nil {
		t.Error("MemberPrivate with over-long member should fail")
	}
}

func TestMemberFlagsNotAMember(t *testing.T) {
	// Neither exists: must be an error, whatever the exact kind.
	if _, err := MemberFlags("bridge999999", "tap999999"); err == nil {
		t.Error("MemberFlags on nonexistent bridge should fail")
	}
}
