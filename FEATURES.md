# go-freebsd-ifc Features

Complete list of features implemented in the FreeBSD Network Interface Control Library.

## Public API Packages

### 1. **if** - Interface Management
Core interface management functionality.

**Functions:**
- `List()` - List all network interfaces
- `Get(name)` - Get specific interface by name
- `SetUp(name, up)` - Bring interface up/down
- `SetMTU(name, mtu)` - Set interface MTU
- `Rename(oldName, newName)` - Rename interface
- `SetPromisc(name, enable)` - Enable/disable promiscuous mode
- `IsPromisc(name)` - Check if interface is in promiscuous mode
- `GetStats(name)` - Get interface statistics (packets, bytes, errors)

**Types:**
- `Interface` - Interface information (name, index, MTU, flags, addresses)
- `InterfaceFlags` - Interface flags with helper methods (IsUp, IsRunning, IsLoopback)
- `Stats` - Interface statistics (RX/TX packets, bytes, errors, drops, collisions)

**Example:**
```go
iface, _ := ifc.Get("em0")
fmt.Printf("MTU: %d, Up: %v\n", iface.MTU, iface.Flags.IsUp())
```

### 2. **bridge** - Bridge Management
Create and manage network bridges.

**Functions:**
- `Create()` - Create bridge interface
- `Destroy(name)` - Destroy bridge
- `Up(name, up)` - Bring bridge up/down
- `AddMember(bridge, member)` - Add interface to bridge
- `DelMember(bridge, member)` - Remove interface from bridge
- `Members(bridge)` - Get list of member interfaces
- `Get(bridge)` - Get bridge info (members, MTU, state)

**Example:**
```go
br, _ := bridge.Create()
bridge.AddMember(br, "em0")
bridge.Up(br, true)
```

### 3. **epair** - Epair Management
Create virtual Ethernet cable pairs.

**Functions:**
- `Create()` - Create epair (returns both ends)
- `Destroy(anySide)` - Destroy epair (either end)

**Types:**
- `Pair` - Epair with A and B sides

**Example:**
```go
pair, _ := epair.Create()  // pair.A, pair.B
defer epair.Destroy(pair.A)
```

### 4. **vlan** - VLAN Management
802.1Q VLAN interface management.

**Functions:**
- `Create()` - Create VLAN interface
- `Destroy(name)` - Destroy VLAN
- `Configure(name, tag, parent)` - Set VLAN tag and parent
- `Get(name)` - Get VLAN configuration
- `Up(name, up)` - Bring VLAN up/down

**Types:**
- `Config` - VLAN configuration (tag 1-4094, parent, MTU, state)

**Example:**
```go
vl, _ := vlan.Create()
vlan.Configure(vl, 100, "em0")
vlan.Up(vl, true)
```

### 5. **lagg** - Link Aggregation
Link aggregation (bonding) management.

**Functions:**
- `Create()` - Create LAGG interface
- `Destroy(name)` - Destroy LAGG
- `SetProto(name, proto)` - Set aggregation protocol
- `AddPort(lagg, port)` - Add port to LAGG
- `DelPort(lagg, port)` - Remove port from LAGG
- `Get(name)` - Get LAGG configuration
- `Up(name, up)` - Bring LAGG up/down

**Protocols:**
- `ProtoFailover` - Active/backup failover
- `ProtoLoadBalance` - Hash-based load balancing
- `ProtoLACP` - IEEE 802.3ad LACP
- `ProtoRoundRobin` - Round-robin distribution
- `ProtoBroadcast` - Broadcast to all ports

**Example:**
```go
lagg0, _ := lagg.Create()
lagg.SetProto(lagg0, lagg.ProtoLACP)
lagg.AddPort(lagg0, "em0")
lagg.AddPort(lagg0, "em1")
```

### 6. **tap** - TAP Interface Management
Layer 2 (Ethernet) virtual interfaces.

**Functions:**
- `Create()` - Create TAP interface
- `Destroy(name)` - Destroy TAP
- `Up(name, up)` - Bring TAP up/down

**Use cases:**
- VPN tunnels (OpenVPN, WireGuard)
- Virtual machine networking
- Container/jail networking

**Example:**
```go
tap0, _ := tap.Create()
tap.Up(tap0, true)
```

### 7. **tun** - TUN Interface Management
Layer 3 (IP) virtual interfaces.

**Functions:**
- `Create()` - Create TUN interface
- `Destroy(name)` - Destroy TUN
- `Up(name, up)` - Bring TUN up/down

**Use cases:**
- VPN tunnels
- IP tunneling (GRE, IPsec)
- Point-to-point links

**Example:**
```go
tun0, _ := tun.Create()
tun.Up(tun0, true)
```

### 8. **ip** - IP Address Management
IPv4 and IPv6 address configuration.

**Functions:**
- `Add4(iface, ip, mask)` - Add IPv4 address
- `Del4(iface, ip, mask)` - Delete IPv4 address
- `Add6(iface, ip, prefixLen)` - Add IPv6 address
- `Del6(iface, ip, prefixLen)` - Delete IPv6 address

**Example:**
```go
ip := net.ParseIP("192.168.1.100")
mask := net.IPv4Mask(255, 255, 255, 0)
ip.Add4("em0", ip, mask)
```

### 9. **route** - Routing Management
IPv4 and IPv6 routing table management.

**IPv4 Functions:**
- `AddDefault4(iface, gw)` - Add IPv4 default route
- `DelDefault4(iface, gw)` - Delete IPv4 default route
- `AddRoute4(dst, gw, iface)` - Add IPv4 route
- `DelRoute4(dst, gw, iface)` - Delete IPv4 route

**IPv6 Functions:**
- `AddDefault6(iface, gw)` - Add IPv6 default route
- `DelDefault6(iface, gw)` - Delete IPv6 default route
- `AddRoute6(dst, gw, iface)` - Add IPv6 route
- `DelRoute6(dst, gw, iface)` - Delete IPv6 route

**Example:**
```go
// IPv4
_, dst, _ := net.ParseCIDR("10.0.0.0/24")
gw := net.ParseIP("192.168.1.1")
route.AddRoute4(dst, gw, "em0")

// IPv6
_, dst6, _ := net.ParseCIDR("2001:db8::/32")
gw6 := net.ParseIP("fe80::1")
route.AddRoute6(dst6, gw6, "em0")
```

## Internal Packages

Implementation details hidden from users:

- **internal/syscall** - Socket and ioctl wrappers
- **internal/constants** - System constants (ioctl numbers, flags)
- **internal/ifops** - Interface operations
- **internal/bridgeops** - Bridge operations
- **internal/cloneops** - Clone interface operations
- **internal/vlanops** - VLAN operations
- **internal/laggops** - LAGG operations
- **internal/ipaddr** - IP address operations
- **internal/routing** - Routing operations

## Error Handling

Comprehensive error types:

- `ErrPermission` - Operation not permitted (need root)
- `ErrNotFound` - Resource not found
- `ErrExists` - Resource already exists
- `ErrInvalidArgument` - Invalid argument
- `ErrBusy` - Resource busy
- `ValidationError` - Input validation error
- `OperationError` - Operation-specific error

**Helper functions:**
- `IsNotFound(err)` - Check if error is not found
- `IsPermission(err)` - Check if error is permission denied
- `IsValidation(err)` - Check if error is validation error

## Example Programs

12 comprehensive example programs:

1. **list** - List all network interfaces
2. **list-vlans** - List VLAN interfaces
3. **iface-config** - Interface configuration tool (MTU, up/down, promisc, rename)
4. **ifstats** - Interface statistics viewer (show/list/watch with real-time updates)
5. **vlan-demo** - VLAN management CLI
6. **tap-tun-demo** - TAP/TUN management CLI
7. **lagg-demo** - LAGG management CLI
8. **ipv6-routing** - IPv6 routing CLI
9. **comprehensive-demo** - All features demonstration
10. **net-bridge-up** - Bridge + epair setup
11. **ip-addr** - IP address management
12. **route-default** - IPv4 routing management

## Key Features

### Architecture
- ✅ Clean separation: public APIs vs internal implementation
- ✅ All cgo code isolated in internal packages
- ✅ Consistent error handling across all packages
- ✅ Idempotent operations where appropriate

### Safety
- ✅ Root privilege checking
- ✅ Input validation
- ✅ Descriptive error messages
- ✅ Type-safe error checking with errors.Is/As

### Testing
- ✅ Unit tests for all public APIs
- ✅ Integration tests (E2E) with IFCLIB_E2E=1 gate
- ✅ Test coverage reporting
- ✅ Comprehensive test documentation

### Documentation
- ✅ GoDoc comments on all public APIs
- ✅ Package-level documentation
- ✅ Example code in documentation
- ✅ Comprehensive README files
- ✅ Error handling guide
- ✅ Testing guide

## Supported FreeBSD Versions

Tested on FreeBSD 14.x. Should work on FreeBSD 12.x and 13.x with minor adjustments.

## Performance Considerations

- Efficient: Uses native system calls, no external tools
- Low overhead: Direct ioctl calls, minimal allocations
- Thread-safe: No shared state, safe for concurrent use

## Common Use Cases

1. **Network Virtualization**: Create isolated network environments (bridges, VLANs, tap/tun)
2. **High Availability**: Link aggregation with LACP for redundancy
3. **Container Networking**: Create network namespaces for jails/containers
4. **VPN Servers**: Manage tap/tun interfaces for VPN endpoints
5. **Network Testing**: Create test topologies programmatically
6. **SDN/NFV**: Software-defined networking infrastructure
7. **Traffic Monitoring**: Enable promiscuous mode for packet capture
8. **IPv6 Migration**: Dual-stack configuration and routing

## Limitations

- FreeBSD-specific (uses FreeBSD-specific ioctls and system calls)
- Requires root privileges for most write operations
- Netgraph integration not implemented
- Some advanced bridge features (STP configuration) not implemented

## Future Enhancements

Potential additions for v1.1+:

### Network Virtualization & Advanced Topologies
- **Netgraph Integration** - FreeBSD's powerful kernel-level networking framework
  - Node creation and management
  - Hook connections and message passing
  - Support for netgraph node types (ng_bridge, ng_ether, ng_socket, etc.)
- **ng_one2many** - One-to-many and many-to-one packet multiplexing
  - Load balancing across multiple links
  - Redundancy and failover configurations
  - Packet duplication for monitoring
  - Dynamic link management (add/remove links at runtime)
  - Per-link statistics and monitoring
  - Use cases: WAN failover, multipath routing, packet replication

### Interface Features
- Full interface statistics via sysctl (extended if_data fields)
- Interface description management (set/get descriptions)
- Interface groups management
- Advanced MTU path discovery

### Advanced Bridge Features
- STP (Spanning Tree Protocol) configuration
- RSTP (Rapid Spanning Tree Protocol) support
- Bridge filtering and MAC learning controls
- VLAN filtering on bridge ports

### Additional Interface Types
- VXLAN support (Virtual Extensible LAN)
- GRE tunnel support (Generic Routing Encapsulation)
- GIF tunnel support (Generic Tunnel Interface)
- WireGuard interface integration

### Monitoring & Diagnostics
- Extended interface event monitoring
- Network topology discovery
- Real-time traffic analysis integration
- Performance metrics collection
