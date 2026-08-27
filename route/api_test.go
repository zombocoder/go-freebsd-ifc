//go:build freebsd
// +build freebsd

package route

import (
	"net"
	"os"
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

// TestAddDelRoute4 tests IPv4 route addition and deletion
func TestAddDelRoute4(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Use lo0 for testing
	iface := "lo0"
	gw := net.ParseIP("127.0.0.1")
	_, dst, _ := net.ParseCIDR("198.51.100.0/24")

	// Add route
	if err := AddRoute4(dst, gw, iface); err != nil {
		t.Fatalf("AddRoute4() failed: %v", err)
	}

	// Add again (should be idempotent)
	if err := AddRoute4(dst, gw, iface); err != nil {
		t.Errorf("AddRoute4() should be idempotent, got error: %v", err)
	}

	// Delete route
	if err := DelRoute4(dst, gw, iface); err != nil {
		t.Errorf("DelRoute4() failed: %v", err)
	}

	// Delete again (should be idempotent)
	if err := DelRoute4(dst, gw, iface); err != nil {
		t.Errorf("DelRoute4() should be idempotent, got error: %v", err)
	}
}

// TestInvalidGateway tests error handling for invalid gateway
func TestInvalidGateway(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	_, dst, _ := net.ParseCIDR("192.0.2.0/24")

	// IPv6 gateway for IPv4 route (should fail)
	gw := net.ParseIP("::1")

	err := AddRoute4(dst, gw, "lo0")
	if err == nil {
		t.Error("AddRoute4() should fail with IPv6 gateway")
	}
}

// TestInvalidDestination tests error handling for invalid destination
func TestInvalidDestination(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	gw := net.ParseIP("127.0.0.1")

	// Nil destination
	err := AddRoute4(nil, gw, "lo0")
	if err == nil {
		t.Error("AddRoute4() should fail with nil destination")
	}
}

// TestInvalidInterface tests operations on non-existent interface
func TestInvalidInterface(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	gw := net.ParseIP("127.0.0.1")
	_, dst, _ := net.ParseCIDR("203.0.113.0/24")

	err := AddRoute4(dst, gw, "nonexistent999")
	if err == nil {
		t.Error("AddRoute4() should fail with non-existent interface")
	}
}

// TestAddDelRoute6 tests IPv6 route addition and deletion
func TestAddDelRoute6(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	// Use lo0 for testing
	iface := "lo0"
	gw := net.ParseIP("::1")
	_, dst, _ := net.ParseCIDR("2001:db8::/32")

	// Add route
	if err := AddRoute6(dst, gw, iface); err != nil {
		t.Fatalf("AddRoute6() failed: %v", err)
	}

	// Add again (should be idempotent)
	if err := AddRoute6(dst, gw, iface); err != nil {
		t.Errorf("AddRoute6() should be idempotent, got error: %v", err)
	}

	// Delete route
	if err := DelRoute6(dst, gw, iface); err != nil {
		t.Errorf("DelRoute6() failed: %v", err)
	}

	// Delete again (should be idempotent)
	if err := DelRoute6(dst, gw, iface); err != nil {
		t.Errorf("DelRoute6() should be idempotent, got error: %v", err)
	}
}

// TestInvalidGateway6 tests error handling for invalid IPv6 gateway
func TestInvalidGateway6(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	_, dst, _ := net.ParseCIDR("2001:db8:1::/48")

	// IPv4 gateway for IPv6 route (should fail)
	gw := net.ParseIP("127.0.0.1")

	err := AddRoute6(dst, gw, "lo0")
	if err == nil {
		t.Error("AddRoute6() should fail with IPv4 gateway")
	}
}

// TestInvalidDestination6 tests error handling for invalid IPv6 destination
func TestInvalidDestination6(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	gw := net.ParseIP("::1")

	// IPv4 destination for IPv6 route (should fail)
	_, dst, _ := net.ParseCIDR("192.0.2.0/24")

	err := AddRoute6(dst, gw, "lo0")
	if err == nil {
		t.Error("AddRoute6() should fail with IPv4 destination")
	}
}

// hasRoute reports whether the table contains a route to dst via gw.
func hasRoute(t *testing.T, routes []Route, dst *net.IPNet, gw net.IP) bool {
	t.Helper()

	for _, r := range routes {
		if r.Dst == nil || r.Dst.String() != dst.String() {
			continue
		}
		if gw == nil {
			return true
		}
		if r.Gateway != nil && r.Gateway.Equal(gw) {
			return true
		}
	}
	return false
}

func countRoutes(t *testing.T, routes []Route, dst *net.IPNet) int {
	t.Helper()

	n := 0
	for _, r := range routes {
		if r.Dst != nil && r.Dst.String() == dst.String() {
			n++
		}
	}
	return n
}

// TestList reads the routing table without privileges.
//
// This is the read side the routing package was missing; a route library that
// cannot read the table back cannot reconcile against it.
func TestList(t *testing.T) {
	routes, err := List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(routes) == 0 {
		t.Fatal("List() returned no routes, expected at least the loopback route")
	}

	// 127.0.0.1 always has a host route through lo0.
	found := false
	for _, r := range routes {
		if r.Dst != nil && r.Dst.IP.Equal(net.ParseIP("127.0.0.1")) && r.Iface == "lo0" {
			found = true
			if !r.Flags.IsHost() {
				t.Error("the 127.0.0.1 route is not flagged as a host route")
			}
			if !r.Flags.IsUp() {
				t.Error("the 127.0.0.1 route is not flagged up")
			}
		}
	}
	if !found {
		t.Errorf("no host route for 127.0.0.1 via lo0 in %d routes", len(routes))
	}
}

// TestList4List6 checks the family filters.
//
// The family is checked with Route.Family, which comes from the destination
// sockaddr's sa_family, not by looking at the shape of Dst. FreeBSD's IPv6
// table contains a real ::ffff:0.0.0.0/96 route, and for that value
// net.IP.To4() returns non-nil, so inferring the family from the address would
// misclassify a genuine IPv6 route as IPv4.
func TestList4List6(t *testing.T) {
	v4, err := List4()
	if err != nil {
		t.Fatalf("List4() failed: %v", err)
	}
	for _, r := range v4 {
		if !r.Is4() {
			t.Errorf("List4() returned a %s route: %s", r.Family, r)
		}
		if len(r.Dst.IP) != net.IPv4len {
			t.Errorf("List4() route %s has a %d-byte destination, want 4", r, len(r.Dst.IP))
		}
	}

	v6, err := List6()
	if err != nil {
		t.Fatalf("List6() failed: %v", err)
	}
	for _, r := range v6 {
		if !r.Is6() {
			t.Errorf("List6() returned a %s route: %s", r.Family, r)
		}
		if len(r.Dst.IP) != net.IPv6len {
			t.Errorf("List6() route %s has a %d-byte destination, want 16", r, len(r.Dst.IP))
		}
	}

	all, err := List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(all) < len(v4) {
		t.Errorf("List() returned %d routes, fewer than List4()'s %d", len(all), len(v4))
	}
}

// TestAddListDelRoute4 proves the round trip against a live kernel: the route
// really lands in the table, and really leaves it again.
//
// Note that net.route.multipath may be 1 on the host, in which case a
// conflicting RTM_ADD can build an ECMP group rather than returning EEXIST. The
// test therefore counts matching entries instead of assuming an error, so a
// duplicate that silently became a second nexthop would be caught.
func TestAddListDelRoute4(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	iface := "lo0"
	gw := net.ParseIP("127.0.0.1")
	_, dst, _ := net.ParseCIDR("198.51.100.0/24")

	before, err := List4()
	if err != nil {
		t.Fatalf("List4() failed: %v", err)
	}
	if hasRoute(t, before, dst, nil) {
		t.Fatalf("%s is already in the routing table, refusing to touch it", dst)
	}

	if err := AddRoute4(dst, gw, iface); err != nil {
		t.Fatalf("AddRoute4() failed: %v", err)
	}
	// Remove the route however the test ends.
	t.Cleanup(func() {
		for i := 0; i < 4; i++ {
			after, err := List4()
			if err != nil || !hasRoute(t, after, dst, nil) {
				return
			}
			if err := DelRoute4(dst, gw, iface); err != nil {
				t.Errorf("cleanup: DelRoute4() failed: %v", err)
				return
			}
		}
		t.Errorf("cleanup: %s is still in the routing table", dst)
	})

	added, err := List4()
	if err != nil {
		t.Fatalf("List4() failed: %v", err)
	}
	if !hasRoute(t, added, dst, gw) {
		t.Fatalf("%s via %s is not in the table after AddRoute4()", dst, gw)
	}
	if n := countRoutes(t, added, dst); n != 1 {
		t.Fatalf("%s appears %d times after one AddRoute4(), want 1", dst, n)
	}

	// A second identical add must converge, not accumulate a second
	// nexthop, whatever net.route.multipath is set to.
	if err := AddRoute4(dst, gw, iface); err != nil {
		t.Errorf("AddRoute4() should be idempotent, got error: %v", err)
	}
	dup, err := List4()
	if err != nil {
		t.Fatalf("List4() failed: %v", err)
	}
	if n := countRoutes(t, dup, dst); n != 1 {
		t.Errorf("%s appears %d times after a duplicate AddRoute4(), want 1", dst, n)
	}

	if err := DelRoute4(dst, gw, iface); err != nil {
		t.Fatalf("DelRoute4() failed: %v", err)
	}
	gone, err := List4()
	if err != nil {
		t.Fatalf("List4() failed: %v", err)
	}
	if hasRoute(t, gone, dst, nil) {
		t.Fatalf("%s is still in the table after DelRoute4()", dst)
	}

	if err := DelRoute4(dst, gw, iface); err != nil {
		t.Errorf("DelRoute4() should be idempotent, got error: %v", err)
	}

	// The table must be back where it started.
	if len(gone) != len(before) {
		t.Errorf("routing table has %d IPv4 routes, started with %d", len(gone), len(before))
	}
}

// TestAddListDelRoute6 is TestAddListDelRoute4 for IPv6.
func TestAddListDelRoute6(t *testing.T) {
	skipIfNotRoot(t)
	skipIfNotE2E(t)

	iface := "lo0"
	gw := net.ParseIP("::1")
	_, dst, _ := net.ParseCIDR("2001:db8:5555::/48")

	before, err := List6()
	if err != nil {
		t.Fatalf("List6() failed: %v", err)
	}
	if hasRoute(t, before, dst, nil) {
		t.Fatalf("%s is already in the routing table, refusing to touch it", dst)
	}

	if err := AddRoute6(dst, gw, iface); err != nil {
		t.Fatalf("AddRoute6() failed: %v", err)
	}
	t.Cleanup(func() {
		for i := 0; i < 4; i++ {
			after, err := List6()
			if err != nil || !hasRoute(t, after, dst, nil) {
				return
			}
			if err := DelRoute6(dst, gw, iface); err != nil {
				t.Errorf("cleanup: DelRoute6() failed: %v", err)
				return
			}
		}
		t.Errorf("cleanup: %s is still in the routing table", dst)
	})

	added, err := List6()
	if err != nil {
		t.Fatalf("List6() failed: %v", err)
	}
	if !hasRoute(t, added, dst, gw) {
		t.Fatalf("%s via %s is not in the table after AddRoute6()", dst, gw)
	}
	if n := countRoutes(t, added, dst); n != 1 {
		t.Fatalf("%s appears %d times after one AddRoute6(), want 1", dst, n)
	}

	if err := DelRoute6(dst, gw, iface); err != nil {
		t.Fatalf("DelRoute6() failed: %v", err)
	}
	gone, err := List6()
	if err != nil {
		t.Fatalf("List6() failed: %v", err)
	}
	if hasRoute(t, gone, dst, nil) {
		t.Fatalf("%s is still in the table after DelRoute6()", dst)
	}
	if len(gone) != len(before) {
		t.Errorf("routing table has %d IPv6 routes, started with %d", len(gone), len(before))
	}
}

// TestMappedIPv6RouteKeepsItsFamily is the regression test for IPv4-mapped
// IPv6 destinations.
//
// FreeBSD's IPv6 table carries a ::ffff:0.0.0.0/96 route. It is a genuine IPv6
// route and List6 must return it as one, but Go's net.IP conflates the
// families for such values: net.IP.To4() returns non-nil, and
// net.IPNet.String() collapses ::ffff:0.0.0.0/96 to "0.0.0.0/0" because
// networkNumberAndMask truncates both the address and the mask to four bytes.
// Classifying by the value's shape therefore reports a real IPv6 route as
// IPv4, and printing it loses both the family and the prefix length.
//
// The fix carries sa_family from the kernel's sockaddr, so this checks the
// route is present, is classified IPv6, keeps its 16-byte width, and prints in
// IPv6 notation. It does not assert the route exists, because it is a system
// route this suite must never create: if the host does not have it the checks
// are skipped rather than faked.
func TestMappedIPv6RouteKeepsItsFamily(t *testing.T) {
	v6, err := List6()
	if err != nil {
		t.Fatalf("List6() failed: %v", err)
	}

	mapped := net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0, 0, 0, 0}

	var found *Route
	for i := range v6 {
		if len(v6[i].Dst.IP) == net.IPv6len && v6[i].Dst.IP.Equal(mapped) {
			found = &v6[i]
			break
		}
	}
	if found == nil {
		t.Skip("this host has no ::ffff:0.0.0.0/96 route")
	}

	// The premise: Go really does treat this address as IPv4.
	if found.Dst.IP.To4() == nil {
		t.Fatal("premise broken: To4() should be non-nil for an IPv4-mapped address")
	}
	if got := found.Dst.String(); got != "0.0.0.0/0" {
		t.Logf("note: net.IPNet.String() gave %q, expected the collapsed %q", got, "0.0.0.0/0")
	}

	// What the fix guarantees.
	if !found.Is6() {
		t.Errorf("::ffff:0.0.0.0/96 classified as %s, want inet6", found.Family)
	}
	if found.Is4() {
		t.Error("::ffff:0.0.0.0/96 classified as IPv4")
	}
	if len(found.Dst.IP) != net.IPv6len {
		t.Errorf("destination is %d bytes, want 16", len(found.Dst.IP))
	}
	if ones, bits := found.Dst.Mask.Size(); ones != 96 || bits != 128 {
		t.Errorf("mask is /%d of %d bits, want /96 of 128", ones, bits)
	}
	if got, want := found.prefixString(), "::ffff:0.0.0.0/96"; got != want {
		t.Errorf("Route prefix printed as %q, want %q", got, want)
	}

	// And it must not also show up in the IPv4 table.
	v4, err := List4()
	if err != nil {
		t.Fatalf("List4() failed: %v", err)
	}
	for _, r := range v4 {
		if len(r.Dst.IP) == net.IPv6len {
			t.Errorf("List4() returned a 16-byte destination: %s", r)
		}
	}
}
