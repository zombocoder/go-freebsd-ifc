//go:build freebsd
// +build freebsd

package ifc

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
)

// ifconfigOut returns the output of "ifconfig <name>".
func ifconfigOut(t *testing.T, name string) string {
	t.Helper()
	out, err := exec.Command("ifconfig", name).CombinedOutput()
	if err != nil {
		t.Fatalf("ifconfig %s: %v\n%s", name, err, out)
	}
	return string(out)
}

// ifconfigDescr extracts the "description:" line value ("" when absent).
func ifconfigDescr(t *testing.T, name string) string {
	t.Helper()
	for _, line := range strings.Split(ifconfigOut(t, name), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "description: ") {
			return strings.TrimPrefix(line, "description: ")
		}
	}
	return ""
}

// descrMaxLen reads the kernel limit (includes the terminating NUL).
func descrMaxLen(t *testing.T) int {
	t.Helper()
	// Named net.link.ifdescr_maxlen in some documentation, net.ifdescr_maxlen
	// on current FreeBSD.
	for _, oid := range []string{"net.ifdescr_maxlen", "net.link.ifdescr_maxlen"} {
		out, err := exec.Command("sysctl", "-n", oid).Output()
		if err != nil {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err == nil && n > 0 {
			return n
		}
	}
	t.Skip("cannot read ifdescr_maxlen sysctl")
	return 0
}

func TestDescriptionValidation(t *testing.T) {
	long := strings.Repeat("x", 16)
	if err := SetDescription("", "d"); err == nil {
		t.Error("SetDescription with empty name should fail")
	}
	if err := SetDescription(long, "d"); err == nil {
		t.Error("SetDescription with over-long name should fail")
	}
	if err := SetDescription("lo0", "a\x00b"); err == nil {
		t.Error("SetDescription with embedded NUL should fail")
	}
	if _, err := Description(""); err == nil {
		t.Error("Description with empty name should fail")
	}
	if _, err := Description(long); err == nil {
		t.Error("Description with over-long name should fail")
	}
}

func TestDescriptionNotFound(t *testing.T) {
	if _, err := Description("nonexist999"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Description(nonexistent) = %v, want ErrNotFound", err)
	}
}

// TestDescription sets, reads, replaces and clears a description on a tap
// created by the test and cross-checks ifconfig(8).
func TestDescription(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	name, err := cloneops.Create("tap")
	if err != nil {
		t.Fatalf("create tap: %v", err)
	}
	t.Cleanup(func() {
		if err := cloneops.Destroy(name); err != nil {
			t.Errorf("cleanup: destroy %s: %v", name, err)
		}
	})

	// A fresh interface has no description.
	got, err := Description(name)
	if err != nil {
		t.Fatalf("Description on fresh interface: %v", err)
	}
	if got != "" {
		t.Errorf("fresh description = %q, want empty", got)
	}

	const d1 = "microvm-runner vm=abc123"
	if err := SetDescription(name, d1); err != nil {
		t.Fatalf("SetDescription: %v", err)
	}
	if got, err := Description(name); err != nil || got != d1 {
		t.Errorf("Description = %q, %v; want %q", got, err, d1)
	}
	if got := ifconfigDescr(t, name); got != d1 {
		t.Errorf("ifconfig description = %q, want %q", got, d1)
	}

	// Replace.
	const d2 = "second"
	if err := SetDescription(name, d2); err != nil {
		t.Fatalf("SetDescription (replace): %v", err)
	}
	if got, _ := Description(name); got != d2 {
		t.Errorf("after replace Description = %q, want %q", got, d2)
	}

	// Clear.
	if err := SetDescription(name, ""); err != nil {
		t.Fatalf("SetDescription (clear): %v", err)
	}
	if got, err := Description(name); err != nil || got != "" {
		t.Errorf("after clear Description = %q, %v; want empty", got, err)
	}
	if got := ifconfigDescr(t, name); got != "" {
		t.Errorf("ifconfig still shows description %q after clear", got)
	}
	// Clearing twice is fine.
	if err := SetDescription(name, ""); err != nil {
		t.Errorf("clearing an empty description: %v", err)
	}
}

// TestDescriptionLimit checks the longest accepted description and that one
// byte more is refused without disturbing the stored value.
func TestDescriptionLimit(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)
	max := descrMaxLen(t)

	name, err := cloneops.Create("tap")
	if err != nil {
		t.Fatalf("create tap: %v", err)
	}
	t.Cleanup(func() {
		if err := cloneops.Destroy(name); err != nil {
			t.Errorf("cleanup: destroy %s: %v", name, err)
		}
	})

	// The kernel limit counts the terminating NUL.
	fits := strings.Repeat("a", max-1)
	if err := SetDescription(name, fits); err != nil {
		t.Fatalf("SetDescription of %d bytes: %v", len(fits), err)
	}
	got, err := Description(name)
	if err != nil {
		t.Fatalf("Description: %v", err)
	}
	if got != fits {
		t.Errorf("long description round trip: got %d bytes, want %d", len(got), len(fits))
	}
	if got := ifconfigDescr(t, name); got != fits {
		t.Errorf("ifconfig long description: got %d bytes, want %d", len(got), len(fits))
	}

	tooLong := strings.Repeat("b", max)
	if err := SetDescription(name, tooLong); err == nil {
		t.Errorf("SetDescription of %d bytes should fail", len(tooLong))
	}
	if got, _ := Description(name); got != fits {
		t.Errorf("failed set changed the stored description (%d bytes)", len(got))
	}
}
