//go:build freebsd
// +build freebsd

package bridge

import (
	"errors"
	"strings"
	"testing"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
)

// Bad names are refused before any ioctl, so this runs without root.
func TestBridgeNameValidation(t *testing.T) {
	long := strings.Repeat("x", 16)
	for _, c := range []struct{ br, member string }{
		{"", "tap0"}, {"bridge0", ""}, {long, "tap0"}, {"bridge0", long},
	} {
		if err := AddMember(c.br, c.member); !errors.Is(err, ifc.ErrInvalidArgument) {
			t.Errorf("AddMember(%q, %q) = %v, want ErrInvalidArgument", c.br, c.member, err)
		}
		if err := DelMember(c.br, c.member); !errors.Is(err, ifc.ErrInvalidArgument) {
			t.Errorf("DelMember(%q, %q) = %v, want ErrInvalidArgument", c.br, c.member, err)
		}
		if err := SetMemberPrivate(c.br, c.member, true); !errors.Is(err, ifc.ErrInvalidArgument) {
			t.Errorf("SetMemberPrivate(%q, %q) = %v, want ErrInvalidArgument", c.br, c.member, err)
		}
		if _, err := MemberFlags(c.br, c.member); !errors.Is(err, ifc.ErrInvalidArgument) {
			t.Errorf("MemberFlags(%q, %q) = %v, want ErrInvalidArgument", c.br, c.member, err)
		}
	}
	for _, br := range []string{"", long} {
		if _, err := Members(br); !errors.Is(err, ifc.ErrInvalidArgument) {
			t.Errorf("Members(%q) = %v, want ErrInvalidArgument", br, err)
		}
	}
}

// Flags the kernel ignores or refuses via BRDGSIFFLGS are rejected up front
// instead of being silently dropped. No root needed: validation comes first.
func TestSetMemberFlagRejectsUnstorable(t *testing.T) {
	for _, f := range []MemberFlag{FlagSpan, FlagBSTPAdmEdge, FlagBSTPAdmCost, FlagPrivate | FlagBSTPAdmCost} {
		for _, set := range []bool{true, false} {
			err := SetMemberFlag("bridge0", "tap0", f, set)
			if !errors.Is(err, ifc.ErrInvalidArgument) {
				t.Errorf("SetMemberFlag(%v, %v) = %v, want ErrInvalidArgument", f, set, err)
			}
		}
	}
	// Bits outside every defined flag are not rejected here (the kernel
	// masks them); a settable flag passes validation (and then fails on the
	// missing bridge with something other than a validation error).
	err := SetMemberFlag("bridge999999", "tap999999", FlagPrivate, true)
	if err == nil || errors.Is(err, ifc.ErrInvalidArgument) {
		t.Errorf("SetMemberFlag(FlagPrivate) on missing bridge = %v, want a lookup error", err)
	}
}
