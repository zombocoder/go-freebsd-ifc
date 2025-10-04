//go:build freebsd
// +build freebsd

package epair

import (
	"os"
	"strings"
	"testing"
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

// TestCreateDestroy tests epair creation and destruction
func TestCreateDestroy(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Create epair
	pair, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Verify names
	if pair.A == "" || pair.B == "" {
		t.Error("Create() returned empty interface names")
	}

	if !strings.HasSuffix(pair.A, "a") {
		t.Errorf("A side should end with 'a', got: %s", pair.A)
	}

	if !strings.HasSuffix(pair.B, "b") {
		t.Errorf("B side should end with 'b', got: %s", pair.B)
	}

	// Check that B is derived from A
	expectedB := strings.TrimSuffix(pair.A, "a") + "b"
	if pair.B != expectedB {
		t.Errorf("B side mismatch: got %s, expected %s", pair.B, expectedB)
	}

	// Destroy using A side
	if err := Destroy(pair.A); err != nil {
		t.Errorf("Destroy(%s) failed: %v", pair.A, err)
	}
}

// TestDestroyBothSides tests that destroying either side works
func TestDestroyBothSides(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Create first epair
	pair1, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Destroy using A side
	if err := Destroy(pair1.A); err != nil {
		t.Errorf("Destroy(%s) failed: %v", pair1.A, err)
	}

	// Create second epair
	pair2, err := Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Destroy using B side
	if err := Destroy(pair2.B); err != nil {
		t.Errorf("Destroy(%s) failed: %v", pair2.B, err)
	}
}

// TestPairStruct tests the Pair struct
func TestPairStruct(t *testing.T) {
	pair := Pair{A: "epair0a", B: "epair0b"}

	if pair.A != "epair0a" {
		t.Errorf("Expected A='epair0a', got '%s'", pair.A)
	}

	if pair.B != "epair0b" {
		t.Errorf("Expected B='epair0b', got '%s'", pair.B)
	}
}
