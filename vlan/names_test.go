//go:build freebsd
// +build freebsd

package vlan

import (
	"strings"
	"testing"
)

// Empty and over-long names must be errors, not panics.
func TestNameValidation(t *testing.T) {
	for _, name := range []string{"", strings.Repeat("x", 16)} {
		if _, err := Get(name); err == nil {
			t.Errorf("Get(%q) should fail", name)
		}
		if err := Configure(name, 10, "em0"); err == nil {
			t.Errorf("Configure(%q, 10, em0) should fail", name)
		}
	}
	if err := Configure("vlan0", 10, strings.Repeat("x", 16)); err == nil {
		t.Error("Configure with over-long parent should fail")
	}
}
