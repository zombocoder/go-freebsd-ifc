//go:build freebsd
// +build freebsd

package epair

import (
	"strings"

	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
)

// Pair represents an epair - a pair of connected virtual Ethernet interfaces.
type Pair struct {
	A string // "A" side interface name (e.g., "epair0a")
	B string // "B" side interface name (e.g., "epair0b")
}

// Create creates a new epair.
//
// Returns both interface names. Destroying either side destroys both.
// Requires root privileges.
//
// Example:
//
//	pair, err := epair.Create()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Created: %s <-> %s\n", pair.A, pair.B)
//	defer epair.Destroy(pair.A)
func Create() (Pair, error) {
	nameA, err := cloneops.Create("epair")
	if err != nil {
		return Pair{}, err
	}

	// Derive the "b" side name from the "a" side
	nameB := strings.TrimSuffix(nameA, "a") + "b"

	return Pair{A: nameA, B: nameB}, nil
}

// Destroy destroys an epair.
//
// Either side (A or B) can be specified - both will be destroyed.
// Requires root privileges.
//
// Example:
//
//	if err := epair.Destroy("epair0a"); err != nil {
//		log.Fatal(err)
//	}
func Destroy(anySide string) error {
	return cloneops.Destroy(anySide)
}
