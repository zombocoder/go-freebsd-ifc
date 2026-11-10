//go:build freebsd
// +build freebsd

package bridge

import (
	"os"
	"testing"

	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
)

// skipIfNotRoot skips the test if not running as root
func skipIfNotRoot(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}
}

// skipIfNotE2E skips the test if not running E2E tests
func skipIfNotE2E(t *testing.T) {
	if os.Getenv("IFCLIB_E2E") != "1" {
		t.Skip("E2E tests disabled. Set IFCLIB_E2E=1 to enable")
	}
}

// TestCreateDestroy tests bridge creation and destruction
func TestCreateDestroy(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Create bridge
	br, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if br == "" {
		t.Error("Create() returned empty bridge name")
	}

	// Verify it exists
	info, err := Get(br)
	if err != nil {
		t.Errorf("Get(%s) failed: %v", br, err)
	}

	if info.Name != br {
		t.Errorf("Bridge name mismatch: got %s, expected %s", info.Name, br)
	}

	// Destroy bridge
	if err := Destroy(br); err != nil {
		t.Errorf("Destroy(%s) failed: %v", br, err)
	}

	// Verify it's gone
	_, err = Get(br)
	if err == nil {
		t.Errorf("Get(%s) should fail after Destroy()", br)
	}
}

// TestUpDown tests bringing bridge up and down
func TestUpDown(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	br, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(br)

	// Bring up
	if err := Up(br, true); err != nil {
		t.Errorf("Up(%s, true) failed: %v", br, err)
	}

	info, _ := Get(br)
	if !info.Up {
		t.Error("Bridge should be up after Up(true)")
	}

	// Bring down
	if err := Up(br, false); err != nil {
		t.Errorf("Up(%s, false) failed: %v", br, err)
	}

	info, _ = Get(br)
	if info.Up {
		t.Error("Bridge should be down after Up(false)")
	}
}

// TestMembers tests adding and removing bridge members
func TestMembers(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Create bridge
	br, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(br)

	// Initially should have no members
	members, err := Members(br)
	if err != nil {
		t.Fatalf("Members(%s) failed: %v", br, err)
	}

	if len(members) != 0 {
		t.Errorf("New bridge should have 0 members, got %d", len(members))
	}
}

// TestAddDelMemberIdempotent tests idempotent add/delete operations
func TestAddDelMemberIdempotent(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	br, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	defer Destroy(br)

	// Delete non-existent member should not error (idempotent)
	if err := DelMember(br, "nonexistent999"); err != nil {
		t.Errorf("DelMember() with non-existent member should be idempotent, got error: %v", err)
	}
}

// TestGetNonExistent tests getting non-existent bridge
func TestGetNonExistent(t *testing.T) {
	_, err := Get("bridge999999")
	if err == nil {
		t.Error("Get() should fail for non-existent bridge")
	}
}

// TestAddMemberAfterFailedLookup guards against uninitialised memory in the
// ifdrv request: ifd_name was copied without a terminating NUL into malloc'd
// (not zeroed) memory, so a longer, freed name such as "bridge999999" left
// over from the previous call turned "bridge0" into "bridge0999999".
func TestAddMemberAfterFailedLookup(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	br, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	tap, err := cloneops.Create("tap")
	if err != nil {
		Destroy(br)
		t.Fatalf("create tap: %v", err)
	}
	t.Cleanup(func() {
		_ = DelMember(br, tap)
		cloneops.Destroy(tap)
		Destroy(br)
	})

	if _, err := Members("bridge999999"); err == nil {
		t.Fatal("Members() should fail for a non-existent bridge")
	}
	if err := AddMember(br, tap); err != nil {
		t.Fatalf("AddMember(%s, %s) after a failed lookup: %v", br, tap, err)
	}
	members, err := Members(br)
	if err != nil || len(members) != 1 || members[0] != tap {
		t.Errorf("Members(%s) = %v, %v; want [%s]", br, members, err, tap)
	}
}
