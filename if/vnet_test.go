//go:build freebsd
// +build freebsd

package ifc

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/zombocoder/go-freebsd-ifc/internal/cloneops"
)

// createVNETJail starts a persistent vnet jail with jail(8) and returns its JID.
func createVNETJail(t *testing.T) int {
	t.Helper()
	name := fmt.Sprintf("ifctest_vnet_%d", os.Getpid())
	out, err := exec.Command("jail", "-i", "-c", "name="+name, "path=/", "vnet", "persist").CombinedOutput()
	if err != nil {
		t.Skipf("cannot create vnet jail (VIMAGE kernel required): %v: %s", err, out)
	}
	t.Cleanup(func() {
		if out, err := exec.Command("jail", "-r", name).CombinedOutput(); err != nil {
			t.Errorf("cleanup: remove jail %s: %v: %s", name, err, out)
		}
	})
	jid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		t.Fatalf("parse jid from %q: %v", out, err)
	}
	return jid
}

// TestMoveToVNETAndReclaim moves a tap interface into a vnet jail and back.
func TestMoveToVNETAndReclaim(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	jid := createVNETJail(t)

	name, err := cloneops.Create("tap")
	if err != nil {
		t.Fatalf("create tap: %v", err)
	}
	t.Cleanup(func() { _ = cloneops.Destroy(name) })

	if err := MoveToVNET(name, jid); err != nil {
		t.Fatalf("MoveToVNET(%s, %d): %v", name, jid, err)
	}
	if _, err := Get(name); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after MoveToVNET, Get(%s) = %v, want ErrNotFound", name, err)
	}

	out, err := exec.Command("jexec", strconv.Itoa(jid), "ifconfig", name).CombinedOutput()
	if err != nil {
		t.Fatalf("interface %s not visible inside jail %d: %v: %s", name, jid, err, out)
	}

	if err := ReclaimFromVNET(name, jid); err != nil {
		t.Fatalf("ReclaimFromVNET(%s, %d): %v", name, jid, err)
	}
	if _, err := Get(name); err != nil {
		t.Fatalf("after ReclaimFromVNET, Get(%s): %v", name, err)
	}
}

func TestMoveToVNETErrors(t *testing.T) {
	if err := MoveToVNET("", 1); err == nil {
		t.Error("MoveToVNET with empty name: want error")
	}
	if err := MoveToVNET("tap0", 0); err == nil {
		t.Error("MoveToVNET with jid 0: want error")
	}
	if err := ReclaimFromVNET("tap0", -1); err == nil {
		t.Error("ReclaimFromVNET with negative jid: want error")
	}
}
