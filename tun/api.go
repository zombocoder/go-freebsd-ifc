//go:build freebsd
// +build freebsd

package tun

import (
	"fmt"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
	"github.com/zombocoder/go-freebsd-ifc/internal/ifops"
)

// Create creates a new TUN interface.
//
// The kernel automatically assigns a name (e.g., "tun0", "tun1").
// Requires root privileges.
//
// Example:
//
//	name, err := tun.Create()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Created TUN interface: %s\n", name)
func Create() (string, error) {
	name, err := cloneops.Create("tun")
	if err != nil {
		return "", fmt.Errorf("create tun: %w", err)
	}
	return name, nil
}

// Destroy destroys a TUN interface.
//
// Requires root privileges.
//
// Example:
//
//	if err := tun.Destroy("tun0"); err != nil {
//		log.Fatal(err)
//	}
func Destroy(name string) error {
	if err := cloneops.Destroy(name); err != nil {
		return fmt.Errorf("destroy tun %s: %w", name, err)
	}
	return nil
}

// Up brings the TUN interface up or down.
//
// Requires root privileges.
//
// Example:
//
//	// Bring up
//	if err := tun.Up("tun0", true); err != nil {
//		log.Fatal(err)
//	}
//
//	// Bring down
//	if err := tun.Up("tun0", false); err != nil {
//		log.Fatal(err)
//	}
func Up(name string, up bool) error {
	if err := ifops.SetFlags(name, uint32(ifc.FlagUp), up); err != nil {
		if up {
			return fmt.Errorf("bring tun %s up: %w", name, err)
		}
		return fmt.Errorf("bring tun %s down: %w", name, err)
	}
	return nil
}
