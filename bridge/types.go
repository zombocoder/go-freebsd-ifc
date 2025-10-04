//go:build freebsd
// +build freebsd

package bridge

// Info represents bridge interface information
type Info struct {
	Name    string
	Members []string
	MTU     int
	Up      bool
}
