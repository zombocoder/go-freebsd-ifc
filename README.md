| `AddRoute4(dst *net.IPNet, gw net.IP, iface string) error` | Add route            | Yes           |
| `DelRoute4(dst *net.IPNet, gw net.IP, iface string) error` | Delete route         | Yes           |
| `List() ([]Route, error)`                                  | Read the whole table | No            |
| `List4() ([]Route, error)`                                 | Read the IPv4 table  | No            |
| `List6() ([]Route, error)`                                 | Read the IPv6 table  | No            |
| `List() ([]Route, error)`                                  | Read the whole table | No            |
| `List4() ([]Route, error)`                                 | Read the IPv4 table  | No            |
| `List6() ([]Route, error)`                                 | Read the IPv6 table  | No            |
# go-freebsd-ifc

[![Go Reference](https://pkg.go.dev/badge/github.com/zombocoder/go-freebsd-ifc.svg)](https://pkg.go.dev/github.com/zombocoder/go-freebsd-ifc)
[![FreeBSD](https://img.shields.io/badge/platform-FreeBSD-red.svg)](https://www.freebsd.org/)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

FreeBSD network interface control library for Go with native cgo bindings.

**Latest Release:** v1.0.0

## Features

- 🔧 **Interface Management** - List, query, configure, and monitor network interfaces
  - Get interface details (name, index, MTU, flags, addresses)
  - Set MTU, bring up/down, rename interfaces
  - Enable/disable promiscuous mode
  - Get interface statistics (packets, bytes, errors, drops)
- **Bridge Support** - Create and manage bridge(4) interfaces with member management
- **Epair Support** - Create paired virtual Ethernet interfaces for jails/VMs
- **VLAN Support** - Configure 802.1Q VLAN interfaces (tags 1-4094)
- **VXLAN Support** - RFC 7348 overlays: VNI, endpoints, ports, and static peers via the forwarding table
- **LAGG Support** - Link aggregation with LACP, failover, loadbalance, roundrobin, broadcast
- **TAP/TUN Support** - Layer 2 (TAP) and Layer 3 (TUN) virtual interfaces for VPNs
- **IP Management** - Add/remove IPv4 and IPv6 addresses with full dual-stack support
- **Routing** - Manage IPv4 and IPv6 routing table entries (default routes, static routes) and read the table back
- **Idempotent** - Safe to call operations multiple times
- **Type-Safe** - Clean Go API with proper error handling
- **Statistics** - Real-time interface statistics and monitoring
- **Well Documented** - Comprehensive GoDoc, examples, and guides

## Installation

```bash
go get github.com/zombocoder/go-freebsd-ifc
```

## Requirements

- **OS**: FreeBSD 12.x or later (tested on 14.x)
- **Go**: 1.19 or later
- **Privileges**: Root required for most write operations
- **Build**: C compiler for cgo (comes with FreeBSD base system)

## Quick Start

### List Interfaces

```go
package main

import (
    "fmt"
    "log"

    ifc "github.com/zombocoder/go-freebsd-ifc/if"
)

func main() {
    ifaces, err := ifc.List()
    if err != nil {
        log.Fatal(err)
    }

    for _, iface := range ifaces {
        fmt.Printf("%s: MTU=%d, Up=%v\n",
            iface.Name, iface.MTU, iface.Flags.IsUp())
        for _, addr := range iface.Addrs {
            fmt.Printf("  %s\n", addr.String())
        }
    }
}
```

### Create a Network Bridge

```go
package main

import (
    "log"

    "github.com/zombocoder/go-freebsd-ifc/bridge"
    "github.com/zombocoder/go-freebsd-ifc/epair"
)

func main() {
    // Create bridge
    br, err := bridge.Create()
    if err != nil {
        log.Fatal(err)
    }
    defer bridge.Destroy(br)

    // Create epair
    pair, err := epair.Create()
    if err != nil {
        log.Fatal(err)
    }
    defer epair.Destroy(pair.A)

    // Add epair to bridge
    if err := bridge.AddMember(br, pair.B); err != nil {
        log.Fatal(err)
    }

    // Bring bridge up
    if err := bridge.Up(br, true); err != nil {
        log.Fatal(err)
    }
}
```

## API Documentation

### Package: `if` - Interface Management

```go
import ifc "github.com/zombocoder/go-freebsd-ifc/if"
```

| Function                               | Description             | Root Required |
| -------------------------------------- | ----------------------- | ------------- |
| `List() ([]Interface, error)`          | List all interfaces     | No            |
| `Get(name string) (*Interface, error)` | Get specific interface  | No            |
| `SetUp(name string, up bool) error`    | Bring interface up/down | Yes           |
| `SetMTU(name string, mtu int) error`   | Set interface MTU       | Yes           |
| `Rename(old, new string) error`        | Rename interface        | Yes           |

**Example:**

```go
// Get interface details
iface, err := ifc.Get("em0")
if err != nil {
    log.Fatal(err)
}

// Configure interface
ifc.SetUp("em0", true)
ifc.SetMTU("em0", 9000)
```

### Package: `bridge` - Bridge Management

```go
import "github.com/zombocoder/go-freebsd-ifc/bridge"
```

| Function                                   | Description             | Root Required |
| ------------------------------------------ | ----------------------- | ------------- |
| `Create() (string, error)`                 | Create new bridge       | Yes           |
| `Destroy(name string) error`               | Destroy bridge          | Yes           |
| `Up(name string, up bool) error`           | Bring bridge up/down    | Yes           |
| `AddMember(bridge, member string) error`   | Add member interface    | Yes           |
| `DelMember(bridge, member string) error`   | Remove member interface | Yes           |
| `Members(bridge string) ([]string, error)` | List members            | No            |
| `Get(bridge string) (Info, error)`         | Get bridge info         | No            |

**Example:**

```go
br, _ := bridge.Create()
bridge.AddMember(br, "em0")
bridge.AddMember(br, "em1")
bridge.Up(br, true)

info, _ := bridge.Get(br)
fmt.Printf("Members: %v\n", info.Members)
```

### Package: `epair` - Paired Virtual Interfaces

```go
import "github.com/zombocoder/go-freebsd-ifc/epair"
```

| Function                     | Description                 | Root Required |
| ---------------------------- | --------------------------- | ------------- |
| `Create() (Pair, error)`     | Create epair                | Yes           |
| `Destroy(name string) error` | Destroy epair (either side) | Yes           |

**Example:**

```go
pair, _ := epair.Create()
fmt.Printf("Created: %s <-> %s\n", pair.A, pair.B)

// Typically: pair.B goes into jail, pair.A stays on host
bridge.AddMember("bridge0", pair.B)
```

### Package: `vlan` - VLAN Management

```go
import "github.com/zombocoder/go-freebsd-ifc/vlan"
```

| Function                                                  | Description            | Root Required |
| --------------------------------------------------------- | ---------------------- | ------------- |
| `Create() (string, error)`                                | Create VLAN interface  | Yes           |
| `Destroy(name string) error`                              | Destroy VLAN interface | Yes           |
| `Configure(name string, tag uint16, parent string) error` | Configure VLAN         | Yes           |
| `Get(name string) (Config, error)`                        | Get VLAN config        | No            |
| `Up(name string, up bool) error`                          | Bring VLAN up/down     | Yes           |

**Example:**

```go
vlan, _ := vlan.Create()
vlan.Configure(vlan, 100, "em0")  // Tag 100 on em0
vlan.Up(vlan, true)

cfg, _ := vlan.Get(vlan)
fmt.Printf("VLAN %d on %s\n", cfg.Tag, cfg.Parent)
```

### Package: `vxlan` - VXLAN Overlay Management

```go
import "github.com/zombocoder/go-freebsd-ifc/vxlan"
```

Requires the vxlan(4) driver: `doas kldload if_vxlan`.

| Function                                                                     | Description                    | Root Required |
| ---------------------------------------------------------------------------- | ------------------------------ | ------------- |
| `Create() (string, error)`                                                   | Create VXLAN interface         | Yes           |
| `Destroy(name string) error`                                                 | Destroy VXLAN interface        | Yes           |
| `Configure(name string, p Params) error`                                     | Apply VNI, endpoints, ports... | Yes           |
| `Get(name string) (Config, error)`                                           | Get kernel state               | No            |
| `Up(name string, up bool) error`                                             | Bring interface up/down        | Yes           |
| `PeerAdd(name string, mac net.HardwareAddr, remote net.IP, port uint16) error` | Add a static peer              | Yes           |
| `PeerDel(name string, mac net.HardwareAddr) error`                           | Remove a static peer           | Yes           |
| `Peers(name string) ([]Peer, error)`                                         | List the forwarding table      | No            |
| `FlushPeers(name string, includeStatic bool) error`                          | Flush forwarding entries       | Yes           |
| `Overhead(name string) (int, error)`                                         | Encapsulation overhead         | No            |

**Example:**

```go
name, _ := vxlan.Create()
defer vxlan.Destroy(name)

// Configure while the interface is still down.
vxlan.Configure(name, vxlan.Params{
    VNI:    100,
    Local:  net.ParseIP("192.0.2.1"),
    Remote: net.ParseIP("192.0.2.2"),
})
vxlan.Up(name, true)

cfg, _ := vxlan.Get(name)
fmt.Printf("VNI %d, %s -> %s, MTU %d, running %v\n",
    cfg.VNI, cfg.Local, cfg.Remote, cfg.MTU, cfg.Running)

// Peers are forwarding table entries: a remote MAC and the VTEP that owns it.
mac, _ := net.ParseMAC("02:11:22:33:44:55")
vxlan.PeerAdd(name, mac, net.ParseIP("192.0.2.3"), 0)
peers, _ := vxlan.Peers(name)
```

**Peers.** FreeBSD has no VXLAN peer list; it has a per-interface forwarding
table mapping a remote MAC address to its VTEP, the same structure MAC
learning fills in. `PeerAdd`/`PeerDel` issue `VXLAN_CMD_FTABLE_ENTRY_ADD` and
`_REM`. `ifconfig(8)` implements neither, so these are reachable only through
the ioctl.

**Ordering.** Configure before `Up`: the kernel returns `EBUSY` for every
`VXLAN_CMD_SET_*` once the interface is running. `Up` does not report a
rejected configuration, because `vxlan_init()` in the kernel returns `void`;
the interface is left `UP` but not `RUNNING`. Check `Get().Running`.

**MTU.** VXLAN adds 50 bytes over IPv4 (14 outer Ethernet + 20 IP + 8 UDP +
8 VXLAN) and 70 over IPv6, so the kernel's default MTU is 1450 over IPv4.
`Overhead()` reads that number back from the live interface
(`if_data.ifi_hdrlen`) rather than recomputing it.

### Package: `ip` - IP Address Management

```go
import "github.com/zombocoder/go-freebsd-ifc/ip"
```

| Function                                               | Description         | Root Required |
| ------------------------------------------------------ | ------------------- | ------------- |
| `Add4(iface string, ip net.IP, mask net.IPMask) error` | Add IPv4 address    | Yes           |
| `Del4(iface string, ip net.IP, mask net.IPMask) error` | Delete IPv4 address | Yes           |
| `Add6(iface string, ip net.IP, prefixLen int) error`   | Add IPv6 address    | Yes           |
| `Del6(iface string, ip net.IP, prefixLen int) error`   | Delete IPv6 address | Yes           |

**Example:**

```go
// IPv4
ip4 := net.ParseIP("192.168.1.10")
mask := net.CIDRMask(24, 32)
ip.Add4("em0", ip4, mask)

// IPv6
ip6 := net.ParseIP("fd00::1")
ip.Add6("em0", ip6, 64)
```

### Package: `route` - Routing Management

```go
import "github.com/zombocoder/go-freebsd-ifc/route"
```

| Function                                                   | Description          | Root Required |
| ---------------------------------------------------------- | -------------------- | ------------- |
| `AddDefault4(iface string, gw net.IP) error`               | Add default route    | Yes           |
| `DelDefault4(iface string, gw net.IP) error`               | Delete default route | Yes           |
| `AddRoute4(dst *net.IPNet, gw net.IP, iface string) error` | Add route            | Yes           |
| `DelRoute4(dst *net.IPNet, gw net.IP, iface string) error` | Delete route         | Yes           |
| `List() ([]Route, error)`                                  | Read the whole table | No            |
| `List4() ([]Route, error)`                                 | Read the IPv4 table  | No            |
| `List6() ([]Route, error)`                                 | Read the IPv6 table  | No            |

**Example:**

```go
// Add default route
gw := net.ParseIP("192.168.1.1")
route.AddDefault4("em0", gw)

// Add specific route
_, dst, _ := net.ParseCIDR("10.0.0.0/24")
route.AddRoute4(dst, gw, "em0")

// Read the table back, e.g. to reconcile against it
routes, _ := route.List()
for _, r := range routes {
    fmt.Println(r) // 0.0.0.0/0 via 192.168.1.1 dev em0 flags 0x803
}
```

`Route.Family` (`FamilyIPv4` / `FamilyIPv6`) comes from the destination
sockaddr the kernel returns, not from the shape of `Route.Dst`. Go's `net.IP`
conflates the families for IPv4-mapped addresses, and FreeBSD's IPv6 table
really does contain a `::ffff:0.0.0.0/96` route, which `net.IPNet.String()`
would print as `0.0.0.0/0`. Always test `Family`, never `Dst.IP.To4()`.

## Error Handling

The library provides comprehensive error handling with typed errors and context:

```go
import (
    "errors"
    "github.com/zombocoder/go-freebsd-ifc/bridge"
    isyscall "github.com/zombocoder/go-freebsd-ifc/internal/syscall"
)

err := bridge.AddMember("bridge0", "em0")
if errors.Is(err, isyscall.ErrNotFound) {
    fmt.Println("Bridge or interface not found")
} else if errors.Is(err, isyscall.ErrPermission) {
    fmt.Println("Need root privileges")
} else if isyscall.IsValidation(err) {
    fmt.Println("Invalid parameter")
} else if err != nil {
    fmt.Printf("Error: %v\n", err)
}
```

### Error Types

- `ErrPermission` - Operation requires root privileges
- `ErrNotFound` - Interface/resource not found
- `ErrExists` - Resource already exists
- `ErrInvalidArgument` - Invalid parameter provided
- `ErrBusy` - Resource is in use
- `ErrAddressNotAvailable` - Address is not assigned to the interface (EADDRNOTAVAIL)
- `ErrNotSupported` - Operation not supported
- `ValidationError` - Input validation failed (includes field details)
- `OperationError` - Wraps errors with operation context

All errors include context and support `errors.Is()` and `errors.As()`.

## Idempotency

Most operations are idempotent for safety:

```go
// Adding an existing member returns nil (no error)
bridge.AddMember("bridge0", "em0")
bridge.AddMember("bridge0", "em0") // Returns nil

// Deleting non-existent member returns nil
bridge.DelMember("bridge0", "em1") // Returns nil

// Same for IP addresses and routes
ip.Add4("em0", ipAddr, mask)
ip.Add4("em0", ipAddr, mask) // Returns nil

// Deleting an address that is not assigned returns nil. The kernel reports
// EADDRNOTAVAIL for this; ip.Del4 and ip.Del6 treat it as success.
ip.Del4("em0", ipAddr, mask)
ip.Del4("em0", ipAddr, mask) // Returns nil
```

VXLAN peers are the deliberate exception. `vxlan.PeerDel` reports a missing
entry rather than swallowing it: unlike an address or a route, removing a
forwarding table entry that is not there usually means the caller's idea of
the table is stale.

## Examples

Complete examples are in the `examples/` directory:

```bash
# List interfaces (no root)
go run examples/list/main.go

# Bridge + epair demo (requires root)
sudo go run examples/net-bridge-up/main.go

# IP address management (requires root)
sudo go run examples/ip-addr/main.go

# Route management (requires root)
sudo go run examples/route-default/main.go
```

## Building

```bash
# Build all packages
make build

# Build examples
make examples

# Run tests (unit tests, no root)
make test

# Run integration tests (requires root)
sudo make test-e2e
```

## Architecture

### Clean Separation of Concerns

```
Public API (if/, bridge/, epair/, etc.)
    ↓ (thin wrappers)
Internal Implementation (internal/)
    ├── syscall/     - Socket & ioctl wrappers
    ├── constants/   - All magic numbers
    ├── ifops/       - Interface operations
    ├── bridgeops/   - Bridge operations
    ├── cloneops/    - Clone interface ops
    ├── ipaddr/      - IP address ops
    └── routing/     - Routing ops
```

- **Public packages**: Clean, documented APIs for developers
- **Internal packages**: Implementation details, cannot be imported externally
- **No code duplication**: Shared logic in internal helpers

## Safety & Security

⚠️ **Important Notes:**

1. **Root Required**: Mutation operations require root privileges
2. **System Impact**: Operations modify live network configuration
3. **Testing**: Use jails or VMs for testing to avoid breaking host networking
4. **Idempotency**: Most operations are safe to retry
5. **Error Handling**: Always check errors, especially for routing changes

## Example Programs

The library includes 13 comprehensive example programs demonstrating all features:

```bash
# View all available examples
make list-examples

# Run examples
go run examples/list/main.go                    # List interfaces
go run examples/ifstats/main.go show em0        # Interface statistics
go run examples/ifstats/main.go watch em0 2     # Real-time monitoring
doas go run examples/vlan-demo/main.go create 100 em0
doas go run examples/vxlan-demo/main.go create 100 192.0.2.1 192.0.2.2
doas go run examples/lagg-demo/main.go create lacp
doas go run examples/ipv6-routing/main.go add-default em0 fe80::1
doas go run examples/comprehensive-demo/main.go # All features demo
```

See [examples/README.md](examples/README.md) for detailed documentation.

## Use Cases

- **Jail Networking**: Create bridges and epairs for FreeBSD jails
- **VM Networking**: Configure network for bhyve VMs
- **High Availability**: LACP link aggregation for redundancy
- **VPN Servers**: Manage TAP/TUN interfaces for VPN endpoints
- **SDN/NFV**: Software-defined networking infrastructure
- **Network Testing**: Programmatically create test topologies
- **Traffic Monitoring**: Enable promiscuous mode, collect statistics
- **IPv6 Migration**: Dual-stack network configuration and routing

## Contributing

Contributions welcome! Please:

1. Follow Go best practices
2. Add tests for new functionality
3. Update documentation
4. Ensure FreeBSD compatibility (13/14/15)

## Documentation

- **[FEATURES.md](FEATURES.md)** - Comprehensive feature list
- **[examples/README.md](examples/README.md)** - Examples documentation

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

Developed by zombocoder for FreeBSD network automation.

**Version:** 1.0.0 | **Status:** Stable | **Tested:** FreeBSD 14.x

## See Also

- [FreeBSD Handbook - Networking](https://docs.freebsd.org/en/books/handbook/network/)
- `ifconfig(8)`, `bridge(4)`, `epair(4)`, `vlan(4)`, `vxlan(4)`, `route(8)`
- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/zombocoder/go-freebsd-ifc)
