# go-freebsd-ifc Examples

This directory contains comprehensive examples demonstrating all features of the go-freebsd-ifc library.

## Quick Start

```bash
# List all available examples
make list-examples

# Build all examples
make examples
```

## Examples Overview

### Basic Examples (No Root Required)

#### 1. List Interfaces
Lists all network interfaces on the system.

```bash
go run examples/list/main.go
```

#### 2. List VLANs
Lists all VLAN interfaces with their configuration.

```bash
go run examples/list-vlans/main.go
```

#### 3. Interface Configuration Tool
Show detailed interface information including flags, MTU, and addresses.

```bash
# Show interface details
go run examples/iface-config/main.go show em0

# Show loopback interface
go run examples/iface-config/main.go show lo0
```

### Configuration Management Examples (Root Required)

#### 4. Interface Configuration Management
Complete interface configuration: MTU, up/down, promiscuous mode, rename.

```bash
# Set MTU (jumbo frames)
doas go run examples/iface-config/main.go mtu em0 9000

# Bring interface up/down
doas go run examples/iface-config/main.go up em0
doas go run examples/iface-config/main.go down em0

# Enable/disable promiscuous mode (packet capture)
doas go run examples/iface-config/main.go promisc em0 on
doas go run examples/iface-config/main.go promisc em0 off

# Rename interface
doas go run examples/iface-config/main.go rename em0 wan0
```

**Features demonstrated:**
- MTU configuration (standard 1500, jumbo 9000)
- Interface state management (up/down)
- Promiscuous mode for packet capture
- Interface renaming
- Detailed flag display

### Feature-Specific Examples (Root Required)

#### 5. VLAN Management
Complete VLAN interface management: create, configure, list, show, destroy.

```bash
# List all VLANs
go run examples/vlan-demo/main.go list

# Create VLAN with tag 100 on em0
doas go run examples/vlan-demo/main.go create 100 em0

# Show VLAN details
go run examples/vlan-demo/main.go show vlan0

# Destroy VLAN
doas go run examples/vlan-demo/main.go destroy vlan0
```

**Features demonstrated:**
- VLAN creation with automatic interface naming
- 802.1Q tag configuration (1-4094)
- Parent interface assignment
- VLAN status management

#### 4. TAP/TUN Management
Layer 2 (TAP) and Layer 3 (TUN) virtual interface management.

```bash
# List all TAP/TUN interfaces
go run examples/tap-tun-demo/main.go list

# Create TAP interface (Layer 2)
doas go run examples/tap-tun-demo/main.go create-tap

# Create TUN interface (Layer 3)
doas go run examples/tap-tun-demo/main.go create-tun

# Destroy interfaces
doas go run examples/tap-tun-demo/main.go destroy-tap tap0
doas go run examples/tap-tun-demo/main.go destroy-tun tun0
```

**Features demonstrated:**
- TAP (Ethernet/Layer 2) interface creation
- TUN (IP/Layer 3) interface creation
- Interface state management (up/down)
- Common VPN and virtualization use cases

#### 5. LAGG (Link Aggregation)
Complete link aggregation management with multiple protocols.

```bash
# List all LAGG interfaces
go run examples/lagg-demo/main.go list

# Create LAGG with LACP protocol
doas go run examples/lagg-demo/main.go create lacp

# Add ports to LAGG
doas go run examples/lagg-demo/main.go add-port lagg0 em0
doas go run examples/lagg-demo/main.go add-port lagg0 em1

# Show LAGG details
go run examples/lagg-demo/main.go show lagg0

# Remove port
doas go run examples/lagg-demo/main.go del-port lagg0 em1

# Destroy LAGG
doas go run examples/lagg-demo/main.go destroy lagg0
```

**Supported protocols:**
- `failover` - Active/backup failover
- `loadbalance` - Hash-based load balancing
- `lacp` - IEEE 802.3ad LACP
- `roundrobin` - Round-robin distribution
- `broadcast` - Broadcast to all ports

**Features demonstrated:**
- LAGG interface creation
- Protocol selection
- Port management (add/remove)
- Multi-port aggregation

#### 6. IPv6 Routing
IPv6 routing table management.

```bash
# Add IPv6 default route
doas go run examples/ipv6-routing/main.go add-default em0 fe80::1

# Add specific IPv6 route
doas go run examples/ipv6-routing/main.go add-route 2001:db8::/32 fe80::1 em0

# Delete routes
doas go run examples/ipv6-routing/main.go del-route 2001:db8::/32 fe80::1 em0
doas go run examples/ipv6-routing/main.go del-default em0 fe80::1
```

**Features demonstrated:**
- IPv6 default route management
- IPv6 static route management
- Link-local gateway support
- Route validation

#### 7. Bridge + Epair
Creates a bridge with an epair (virtual cable).

```bash
doas go run examples/net-bridge-up/main.go
```

**Features demonstrated:**
- Bridge creation
- Epair (virtual cable) creation
- Bridge membership management
- Interface state control

#### 8. IP Address Management
IPv4/IPv6 address assignment and management.

```bash
doas go run examples/ip-addr/main.go
```

**Features demonstrated:**
- IPv4 address assignment
- IPv6 address assignment
- Subnet mask/prefix length handling
- Idempotent operations

#### 9. IPv4 Routing
IPv4 routing table management.

```bash
doas go run examples/route-default/main.go
```

**Features demonstrated:**
- IPv4 default route management
- IPv4 static route management
- Gateway configuration
- Interface-based routing

### Comprehensive Demo

#### 10. All Features Demo
Demonstrates all library features in a single program.

```bash
doas go run examples/comprehensive-demo/main.go
```

This demo creates and configures:
- TAP interface (Layer 2 virtual)
- TUN interface (Layer 3 virtual)
- VLAN interface (802.1Q)
- Bridge with epair members
- LAGG interface (link aggregation)
- IPv4 and IPv6 addresses
- IPv4 and IPv6 routes

All interfaces are automatically cleaned up at the end.

## Building Examples

### Build all examples:
```bash
make examples
```

### Build specific example:
```bash
go build -o examples/vlan-demo/main examples/vlan-demo/main.go
```

## Common Patterns

### Error Handling
All examples demonstrate proper error handling:

```go
name, err := vlan.Create()
if err != nil {
    log.Fatalf("Failed to create VLAN: %v", err)
}
```

### Resource Cleanup
Examples use `defer` for automatic cleanup:

```go
name, _ := tap.Create()
defer tap.Destroy(name)
```

### Root Privilege Checking
Examples verify root access when required:

```go
if os.Geteuid() != 0 {
    log.Fatal("This operation requires root privileges")
}
```

## Testing Examples

Most examples can be run in a safe test environment:

### Using FreeBSD Jails
```bash
# Create test jail
doas jail -c name=test path=/jails/test persist

# Run example in jail
doas jexec test go run examples/vlan-demo/main.go list
```

### Using bhyve VM
Test destructive operations in a VM before production use.

## Example Output

### VLAN Demo
```
$ doas go run examples/vlan-demo/main.go create 100 em0
Creating VLAN with tag 100 on em0...
✓ Created VLAN interface: vlan0
Configuring VLAN tag 100 on parent em0...
✓ Configured VLAN
Bringing VLAN up...
✓ VLAN is up

VLAN Interface: vlan0
===================
  VLAN Tag:   100
  Parent:     em0
  State:      UP
  MTU:        1500
  Index:      5
  Addresses:  (none)
```

### LAGG Demo
```
$ doas go run examples/lagg-demo/main.go create lacp
Creating LAGG interface with protocol lacp...
✓ Created LAGG interface: lagg0
Setting protocol to lacp...
✓ Protocol set successfully
Bringing LAGG interface up...
✓ LAGG interface is up

LAGG Interface: lagg0
===================
  Protocol:   lacp
  State:      UP
  MTU:        1500
  Ports (0):
    (no ports)
```

## Documentation

For detailed API documentation, see:
- Library documentation: `make docs`
- Package documentation: `go doc ./vlan`
- Online documentation: https://pkg.go.dev/github.com/zombocoder/go-freebsd-ifc

## Contributing

When adding new examples:
1. Follow the existing pattern (CLI with subcommands)
2. Include comprehensive help text
3. Add error handling with descriptive messages
4. Use `defer` for cleanup
5. Check for root privileges when needed
6. Update Makefile EXAMPLES list
7. Update this README

## License

See main repository LICENSE file.
