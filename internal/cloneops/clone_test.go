//go:build freebsd
// +build freebsd

package cloneops

import "testing"

func TestIsModuleSafeDriver(t *testing.T) {
	for drv, want := range map[string]bool{
		"wg": true, "epair": true, "if_x": true, "vxlan": true,
		"": false, "../x": false, "/tmp/x": false, "a/b": false,
		"wg.x": false, "0wg": false, "Wg": false, "wg-x": false,
	} {
		if got := isModuleSafeDriver(drv); got != want {
			t.Errorf("isModuleSafeDriver(%q) = %v, want %v", drv, got, want)
		}
	}
}
