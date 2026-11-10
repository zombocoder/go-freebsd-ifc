//go:build freebsd
// +build freebsd

package syscall

import (
	"testing"
	"unsafe"
)

func TestCopyStringEmptyAndTruncated(t *testing.T) {
	buf := [8]byte{'x', 'x', 'x', 'x', 'x', 'x', 'x', 'x'}
	// An empty source used to index src[0] and panic.
	CopyString(unsafe.Pointer(&buf[0]), "", 8)
	if buf != [8]byte{'x', 'x', 'x', 'x', 'x', 'x', 'x', 'x'} {
		t.Errorf("empty copy modified the buffer: %v", buf)
	}
	CopyString(unsafe.Pointer(&buf[0]), "abcdefghij", 4)
	if string(buf[:4]) != "abcd" || buf[4] != 'x' {
		t.Errorf("truncated copy wrong: %q", buf[:])
	}
}
