//go:build freebsd
// +build freebsd

package syscall

import (
	"errors"
	"testing"
)

// Path-like names are refused before kldload(2), so this runs without root.
func TestLoadKernelModuleRefusesPaths(t *testing.T) {
	for _, name := range []string{"", "if_/tmp/x", "/boot/kernel/if_wg.ko", "../if_wg", "if_wg\x00"} {
		loaded, err := LoadKernelModule(name)
		if loaded || !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("LoadKernelModule(%q) = %v, %v; want false, ErrInvalidArgument", name, loaded, err)
		}
	}
}
