//go:build freebsd
// +build freebsd

package ifc

import (
	"fmt"
	"os"
	"testing"

	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
)

func skipIfNotRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}
}

func skipIfNotE2E(t *testing.T) {
	t.Helper()
	if os.Getenv("IFCLIB_E2E") != "1" {
		t.Skip("E2E tests disabled. Set IFCLIB_E2E=1 to enable")
	}
}

// TestRename renames a real interface and checks the new name took effect.
//
// SIOCSIFNAME does not carry the new name inline in ifr_ifru: sys/net/if.c
// calls copyinstr(ifr_data_get_ptr(ifr), ...), so the union must hold a
// pointer. Copying the ASCII name into the union instead makes the kernel
// dereference "tap0\0..." as a userland address and return EFAULT.
func TestRename(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	created, err := cloneops.Create("tap")
	if err != nil {
		t.Fatalf("create tap for rename test: %v", err)
	}

	// current tracks whichever name the interface answers to, so cleanup
	// works on every error path including a partially applied rename.
	current := created
	t.Cleanup(func() {
		if err := cloneops.Destroy(current); err != nil {
			t.Errorf("cleanup: destroy %s: %v", current, err)
		}
	})

	renamed := fmt.Sprintf("ifctest%d", os.Getpid()%10000)

	if err := Rename(created, renamed); err != nil {
		t.Fatalf("Rename(%s, %s) failed: %v", created, renamed, err)
	}
	current = renamed

	iface, err := Get(renamed)
	if err != nil {
		t.Fatalf("Get(%s) after rename failed: %v", renamed, err)
	}
	if iface.Name != renamed {
		t.Errorf("interface name after rename = %q, want %q", iface.Name, renamed)
	}

	if _, err := Get(created); err == nil {
		t.Errorf("old name %q is still present after rename", created)
	}

	// Rename back, so the interface is destroyed under the name the kernel
	// handed out.
	if err := Rename(renamed, created); err != nil {
		t.Fatalf("Rename(%s, %s) back failed: %v", renamed, created, err)
	}
	current = created

	if _, err := Get(created); err != nil {
		t.Errorf("Get(%s) after renaming back failed: %v", created, err)
	}
}

// TestRenameTooLong checks the length guard rejects over-long names.
func TestRenameTooLong(t *testing.T) {
	long := "abcdefghijklmnopqrstuvwxyz"
	if err := Rename("lo0", long); err == nil {
		t.Error("Rename() with a name longer than IFNAMSIZ should fail")
	}
}
