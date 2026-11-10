//go:build freebsd
// +build freebsd

package bridge

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
)

// memberLine returns the "member: X flags=...<...>" line of ifconfig for the
// bridge, or "" when the member is not listed.
func memberLine(t *testing.T, br, member string) string {
	t.Helper()
	out, err := exec.Command("ifconfig", br).CombinedOutput()
	if err != nil {
		t.Fatalf("ifconfig %s: %v\n%s", br, err, out)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "member: "+member+" ") {
			return line
		}
	}
	return ""
}

// TestMemberPrivate creates a bridge and two taps, toggles PRIVATE on one
// member and checks the other flags survive and ifconfig agrees.
func TestMemberPrivate(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	br, err := Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	var taps []string
	t.Cleanup(func() {
		// Taps first, then the bridge: removing the member is implicit in
		// destroying the interface, but be explicit.
		for _, tap := range taps {
			_ = DelMember(br, tap)
			if err := cloneops.Destroy(tap); err != nil {
				t.Errorf("cleanup: destroy %s: %v", tap, err)
			}
		}
		if err := Destroy(br); err != nil {
			t.Errorf("cleanup: destroy %s: %v", br, err)
		}
	})
	for i := 0; i < 2; i++ {
		tap, err := cloneops.Create("tap")
		if err != nil {
			t.Fatalf("create tap: %v", err)
		}
		taps = append(taps, tap)
		if err := AddMember(br, tap); err != nil {
			t.Fatalf("AddMember(%s, %s): %v", br, tap, err)
		}
	}
	a, b := taps[0], taps[1]

	// Defaults: learning+discover, not private.
	before, err := MemberFlags(br, a)
	if err != nil {
		t.Fatalf("MemberFlags: %v", err)
	}
	if !before.Has(FlagLearning | FlagDiscover) {
		t.Errorf("default flags %v lack LEARNING,DISCOVER", before)
	}
	if priv, err := MemberPrivate(br, a); err != nil || priv {
		t.Errorf("MemberPrivate default = %v, %v; want false", priv, err)
	}

	// Turn private on.
	if err := SetMemberPrivate(br, a, true); err != nil {
		t.Fatalf("SetMemberPrivate(true): %v", err)
	}
	if priv, err := MemberPrivate(br, a); err != nil || !priv {
		t.Errorf("MemberPrivate after set = %v, %v; want true", priv, err)
	}
	after, _ := MemberFlags(br, a)
	if after != before|FlagPrivate {
		t.Errorf("flags after private = %v, want %v (other flags must be preserved)", after, before|FlagPrivate)
	}
	if line := memberLine(t, br, a); !strings.Contains(line, "PRIVATE") {
		t.Errorf("ifconfig member line %q lacks PRIVATE", line)
	}

	// The sibling is untouched.
	if priv, _ := MemberPrivate(br, b); priv {
		t.Error("setting private on one member affected another")
	}
	if line := memberLine(t, br, b); strings.Contains(line, "PRIVATE") {
		t.Errorf("ifconfig member line %q for sibling shows PRIVATE", line)
	}

	// Idempotent.
	if err := SetMemberPrivate(br, a, true); err != nil {
		t.Errorf("SetMemberPrivate(true) twice: %v", err)
	}

	// Preservation with a non-default flag: clear LEARNING, keep PRIVATE.
	if err := SetMemberFlag(br, a, FlagLearning, false); err != nil {
		t.Fatalf("SetMemberFlag(learning, false): %v", err)
	}
	f, _ := MemberFlags(br, a)
	if f.Has(FlagLearning) || !f.Has(FlagPrivate|FlagDiscover) {
		t.Errorf("flags = %v, want DISCOVER,PRIVATE without LEARNING", f)
	}
	if line := memberLine(t, br, a); strings.Contains(line, "LEARNING") || !strings.Contains(line, "PRIVATE") {
		t.Errorf("ifconfig member line %q disagrees with %v", line, f)
	}
	// Toggling private off keeps LEARNING cleared.
	if err := SetMemberPrivate(br, a, false); err != nil {
		t.Fatalf("SetMemberPrivate(false): %v", err)
	}
	f, _ = MemberFlags(br, a)
	if f.Has(FlagPrivate) || f.Has(FlagLearning) || !f.Has(FlagDiscover) {
		t.Errorf("flags = %v, want DISCOVER only", f)
	}
	if err := SetMemberFlag(br, a, FlagLearning, true); err != nil {
		t.Fatalf("SetMemberFlag(learning, true): %v", err)
	}
	if f, _ = MemberFlags(br, a); f != before {
		t.Errorf("flags = %v, want original %v", f, before)
	}

	// A tap that is not a member is reported as such.
	if _, err := MemberFlags(br, "lo0"); err == nil {
		t.Error("MemberFlags on a non-member should fail")
	}
	if err := SetMemberPrivate(br, "lo0", true); err == nil {
		t.Error("SetMemberPrivate on a non-member should fail")
	}
}
