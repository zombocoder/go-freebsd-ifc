//go:build freebsd
// +build freebsd

package clone

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
)

func skipIfNotRoot(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}
}

func skipIfNotE2E(t *testing.T) {
	if os.Getenv("IFCLIB_E2E") != "1" {
		t.Skip("E2E tests disabled. Set IFCLIB_E2E=1 to enable")
	}
}

// cleanup destroys name when the test ends, unless the test already did.
func cleanup(t *testing.T, name string) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := ifc.Get(name); errors.Is(err, ifc.ErrNotFound) {
			return
		}
		if err := Destroy(name); err != nil {
			t.Logf("cleanup: Destroy(%q): %v", name, err)
		}
	})
}

func TestCreateDestroyNamed(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := Create("tun4242")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	cleanup(t, name)
	if name != "tun4242" {
		t.Fatalf("name = %q, want tun4242", name)
	}
	if _, err := ifc.Get(name); err != nil {
		t.Fatalf("ifc.Get(%q) failed: %v", name, err)
	}
	if out, err := exec.Command("/sbin/ifconfig", name).CombinedOutput(); err != nil {
		t.Fatalf("ifconfig %s: %v: %s", name, err, out)
	}
	if err := Destroy(name); err != nil {
		t.Fatalf("Destroy() failed: %v", err)
	}
	if _, err := ifc.Get(name); !errors.Is(err, ifc.ErrNotFound) {
		t.Fatalf("ifc.Get after Destroy = %v, want ErrNotFound", err)
	}
}

func TestCreateBareDriver(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := Create("tun")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	cleanup(t, name)
	if !strings.HasPrefix(name, "tun") || name == "tun" {
		t.Fatalf("name = %q, want tun<unit>", name)
	}
}

func TestCreateDuplicate(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := Create("tun4243")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	cleanup(t, name)

	_, err = Create("tun4243")
	if !errors.Is(err, syscall.EEXIST) {
		t.Fatalf("second Create() = %v, want EEXIST", err)
	}
}

func TestCreateLoadsModule(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := Create("wg")
	if err != nil {
		t.Fatalf("Create(wg) failed: %v", err)
	}
	cleanup(t, name)
	if !strings.HasPrefix(name, "wg") {
		t.Fatalf("name = %q, want wg prefix", name)
	}
}

func TestNameTooLong(t *testing.T) {
	long := strings.Repeat("a", 16)
	if _, err := Create(long); err == nil {
		t.Fatal("Create() with 16-char name succeeded, want error")
	}
	if err := Destroy(long); err == nil {
		t.Fatal("Destroy() with 16-char name succeeded, want error")
	}
}

// Names that could turn into a kldload(2) path must be rejected before any
// syscall, so this runs without root.
func TestCreateRejectsUnsafeNames(t *testing.T) {
	for _, name := range []string{"../x", "/tmp/x", "if_/x", "wg\x00", "0wg", ".wg", "wg/0", "wg 0"} {
		if _, err := Create(name); err == nil {
			t.Errorf("Create(%q) succeeded, want error", name)
		} else if !strings.Contains(err.Error(), "invalid interface name") {
			t.Errorf("Create(%q) = %v, want validation error", name, err)
		}
	}
}
