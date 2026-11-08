//go:build freebsd
// +build freebsd

package clone

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
	"github.com/zombocoder/go-freebsd-ifc/internal/constants"
	isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

// Create creates a cloned interface and returns its actual name.
//
// name is either a full name with unit ("wg0") or a bare driver name ("wg"),
// in which case the kernel picks the unit. If the driver's cloner is not
// registered, the if_<driver> kernel module is loaded and the create retried.
// The name must start with a letter and contain only letters, digits, '_',
// '.' and '-'.
//
// If the interface already exists, errors.Is(err, syscall.EEXIST) is true.
// Requires root privileges.
//
// Example:
//
//	name, err := clone.Create("wg0")
//	if err != nil {
//		log.Fatal(err)
//	}
func Create(name string) (string, error) {
	if err := checkCreateName(name); err != nil {
		return "", fmt.Errorf("clone create %q: %w", name, err)
	}
	actual, err := cloneops.Create(name)
	if err != nil {
		return "", fmt.Errorf("clone create %s: %w", name, withErrno(err))
	}
	return actual, nil
}

// Destroy destroys a cloned interface by name.
//
// Requires root privileges.
func Destroy(name string) error {
	if err := checkName(name); err != nil {
		return fmt.Errorf("clone destroy %q: %w", name, err)
	}
	if err := cloneops.Destroy(name); err != nil {
		return fmt.Errorf("clone destroy %s: %w", name, withErrno(err))
	}
	return nil
}

func checkName(name string) error {
	if name == "" {
		return errors.New("empty interface name")
	}
	if len(name) >= constants.IFNAMSIZ {
		return fmt.Errorf("interface name too long (max %d characters)", constants.IFNAMSIZ-1)
	}
	return nil
}

// checkCreateName additionally restricts the characters of a name passed to
// Create. The driver part of the name can end up in kldload(2) as
// "if_<driver>", and kldload treats a name containing '/' as a file path, so
// only letters, digits, '_', '.' and '-' are accepted, starting with a letter.
func checkCreateName(name string) error {
	if err := checkName(name); err != nil {
		return err
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		letter := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		if i == 0 && !letter {
			return fmt.Errorf("invalid interface name %q: must start with a letter", name)
		}
		if !letter && !(c >= '0' && c <= '9') && c != '_' && c != '.' && c != '-' {
			return fmt.Errorf("invalid interface name %q: invalid character %q", name, c)
		}
	}
	return nil
}

// withErrno makes errors.Is(err, syscall.EEXIST) work for an existing name;
// internal/syscall maps EEXIST to its own sentinel.
func withErrno(err error) error {
	if errors.Is(err, isyscall.ErrExists) {
		return fmt.Errorf("%w: %w", syscall.EEXIST, err)
	}
	return err
}
