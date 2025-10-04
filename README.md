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
- **LAGG Support** - Link aggregation with LACP, failover, loadbalance, roundrobin, broadcast
- **TAP/TUN Support** - Layer 2 (TAP) and Layer 3 (TUN) virtual interfaces for VPNs
- **IP Management** - Add/remove IPv4 and IPv6 addresses with full dual-stack support
- **Routing** - Manage IPv4 and IPv6 routing table entries (default routes, static routes)
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

**Example:**

```go
// Add default route
gw := net.ParseIP("192.168.1.1")
route.AddDefault4("em0", gw)

// Add specific route
_, dst, _ := net.ParseCIDR("10.0.0.0/24")
route.AddRoute4(dst, gw, "em0")
```

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
```

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

The library includes 12 comprehensive example programs demonstrating all features:

```bash
# View all available examples
make list-examples

# Run examples
go run examples/list/main.go                    # List interfaces
go run examples/ifstats/main.go show em0        # Interface statistics
go run examples/ifstats/main.go watch em0 2     # Real-time monitoring
doas go run examples/vlan-demo/main.go create 100 em0
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
- `ifconfig(8)`, `bridge(4)`, `epair(4)`, `vlan(4)`, `route(8)`
- [pkg.go.dev Documentation](https://pkg.go.dev/github.com/zombocoder/go-freebsd-ifc)
