//go:build freebsd
// +build freebsd

package ifc

import (
	"errors"
	"strings"
	"testing"
)

func TestNameValidation(t *testing.T) {
	long := strings.Repeat("x", 16)
	for _, name := range []string{"", long} {
		if err := SetUp(name, true); !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("SetUp(%q) = %v, want ErrInvalidArgument", name, err)
		}
		if err := SetMTU(name, 1500); !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("SetMTU(%q) = %v, want ErrInvalidArgument", name, err)
		}
		if err := SetPromisc(name, true); !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("SetPromisc(%q) = %v, want ErrInvalidArgument", name, err)
		}
		if err := Rename(name, "new0"); !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("Rename(%q, new0) = %v, want ErrInvalidArgument", name, err)
		}
		if err := SetDescription(name, "d"); !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("SetDescription(%q) = %v, want ErrInvalidArgument", name, err)
		}
		if _, err := Description(name); !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("Description(%q) = %v, want ErrInvalidArgument", name, err)
		}
	}
}
