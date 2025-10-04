//go:build freebsd
// +build freebsd

package vlan

import (
	"fmt"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
	"github.com/zombocoder/go-freebsd-ifc/internal/ifops"
	"github.com/zombocoder/go-freebsd-ifc/internal/vlanops"
)

// Config represents VLAN interface configuration
type Config struct {
	Name   string
	Tag    uint16
	Parent string
	MTU    int
	Up     bool
}

// Create creates a new VLAN interface.
//
// The kernel automatically assigns a name (e.g., "vlan0", "vlan1").
// Requires root privileges.
func Create() (string, error) {
	name, err := cloneops.Create("vlan")
	if err != nil {
		return "", fmt.Errorf("create vlan: %w", err)
	}
	return name, nil
}

// Destroy destroys a VLAN interface.
//
// Requires root privileges.
func Destroy(name string) error {
	if err := cloneops.Destroy(name); err != nil {
		return fmt.Errorf("destroy vlan %s: %w", name, err)
	}
	return nil
}

// Configure sets the VLAN tag and parent interface.
//
// The VLAN tag must be between 1 and 4094 (802.1Q standard).
// Requires root privileges.
func Configure(name string, tag uint16, parent string) error {
	if err := vlanops.Configure(name, tag, parent); err != nil {
		return fmt.Errorf("configure vlan %s (tag=%d, parent=%s): %w", name, tag, parent, err)
	}
	return nil
}

// Get returns VLAN configuration.
//
// Returns Config with Tag, Parent, MTU, and Up status.
func Get(name string) (Config, error) {
	cfg, err := vlanops.Get(name)
	if err != nil {
		return Config{}, fmt.Errorf("get vlan %s config: %w", name, err)
	}

	iface, err := ifc.Get(name)
	if err != nil {
		return Config{}, fmt.Errorf("get vlan %s interface: %w", name, err)
	}

	return Config{
		Name:   name,
		Tag:    cfg.Tag,
		Parent: cfg.Parent,
		MTU:    iface.MTU,
		Up:     iface.Flags.IsUp(),
	}, nil
}

// Up brings the VLAN interface up or down.
//
// Requires root privileges.
func Up(name string, up bool) error {
	if err := ifops.SetFlags(name, uint32(ifc.FlagUp), up); err != nil {
		if up {
			return fmt.Errorf("bring vlan %s up: %w", name, err)
		}
		return fmt.Errorf("bring vlan %s down: %w", name, err)
	}
	return nil
}
