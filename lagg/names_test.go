//go:build freebsd
// +build freebsd

package lagg

import (
	"strings"
	"testing"
)

// Empty and over-long names must be errors, not panics.
func TestNameValidation(t *testing.T) {
	for _, name := range []string{"", strings.Repeat("x", 16)} {
		if err := SetProto(name, ProtoFailover); err == nil {
			t.Errorf("SetProto(%q) should fail", name)
		}
		if _, err := Get(name); err == nil {
			t.Errorf("Get(%q) should fail", name)
		}
		if err := AddPort(name, "em0"); err == nil {
			t.Errorf("AddPort(%q, em0) should fail", name)
		}
	}
}
