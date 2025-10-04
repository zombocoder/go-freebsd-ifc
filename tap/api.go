//go:build freebsd
// +build freebsd

package tap

import (
	"fmt"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
	"github.com/zombocoder/go-freebsd-ifc/internal/ifops"
)

// Create creates a new TAP interface.
//
// The kernel automatically assigns a name (e.g., "tap0", "tap1").
// Requires root privileges.
//
// Example:
//
//	name, err := tap.Create()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Created TAP interface: %s\n", name)
func Create() (string, error) {
	name, err := cloneops.Create("tap")
	if err != nil {
		return "", fmt.Errorf("create tap: %w", err)
	}
	return name, nil
}

// Destroy destroys a TAP interface.
//
// Requires root privileges.
//
// Example:
//
//	if err := tap.Destroy("tap0"); err != nil {
//		log.Fatal(err)
//	}
func Destroy(name string) error {
	if err := cloneops.Destroy(name); err != nil {
		return fmt.Errorf("destroy tap %s: %w", name, err)
	}
	return nil
}

// Up brings the TAP interface up or down.
//
// Requires root privileges.
//
// Example:
//
//	// Bring up
//	if err := tap.Up("tap0", true); err != nil {
//		log.Fatal(err)
//	}
//
//	// Bring down
//	if err := tap.Up("tap0", false); err != nil {
//		log.Fatal(err)
//	}
func Up(name string, up bool) error {
	if err := ifops.SetFlags(name, uint32(ifc.FlagUp), up); err != nil {
		if up {
			return fmt.Errorf("bring tap %s up: %w", name, err)
		}
		return fmt.Errorf("bring tap %s down: %w", name, err)
	}
	return nil
}
