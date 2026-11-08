# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `clone.Create(name)` / `clone.Destroy(name)`: create and destroy any cloned
  interface by exact name (e.g. `wg0`), loading `if_<driver>` when needed.
  `clone.Create` accepts only names that start with a letter and contain
  letters, digits, `_`, `.` and `-`.
- `ifc.ErrInvalidArgument`, `ifc.ErrBusy`, `ifc.ErrNotSupported`,
  `ifc.ErrNetworkDown`, `ifc.ErrAddressInUse` and `ifc.ErrAddressNotAvailable`
  are re-exported from the internal error package (they were already
  documented as part of the error API but could not be referenced).

### Fixed

- Module autoload on concurrent first-time creates: when another caller loaded
  `if_<driver>` between this caller's failed create and its kldload(2), the
  kldload reported EEXIST and the create failed with the original EINVAL
  instead of being retried. The create is now retried whenever the module is
  loaded or already present.

### Security

- Module autoload never passes a path to kldload(2): the driver name must be
  `[a-z][a-z0-9_]*`, and the internal module loader refuses names containing
  `/` or NUL. kldload treats a name with `/` as a file path, so an
  attacker-chosen interface name like `../x` could otherwise load a module
  from an arbitrary relative path in a privileged process.

## [1.1.0] - 2026-11-08

### Added

- `ifc.MoveToVNET(name, jid)` and `ifc.ReclaimFromVNET(name, jid)` move an
  interface into and out of a vnet jail (`SIOCSIFVNET` / `SIOCSIFRVNET`, the
  ioctls behind `ifconfig <if> vnet <jail>` and `-vnet`).
- Creating a cloned interface (`epair`, `bridge`, `lagg`, `vxlan`, ...) loads
  the matching `if_<driver>` kernel module when it is not loaded yet and
  retries, like `ifconfig` does. Previously this failed with
  "invalid argument" on a system where the module was not loaded.
- GitHub Actions CI: gofmt and `go mod tidy` checks, and the full e2e test
  suite as root in FreeBSD 14.5 and 15.1 VMs under QEMU.
- Community files: code of conduct, security policy, changelog, issue and PR
  templates, Dependabot configuration.

### Fixed

- `lagg`: `SetProto`, `AddPort`, `DelPort` and `Get` always failed. The lagg
  ioctls take `struct lagg_reqall` / `struct lagg_reqport` directly, not
  wrapped in a `struct ifreq`. `Get` also no longer passes a Go pointer
  inside a struct to the kernel.
- `lagg`: the `Proto*` constants did not match the kernel's
  `enum lagg_proto` (e.g. `ProtoFailover` selected round-robin). They now
  take their values from `net/if_lagg.h`. Code that uses the named constants
  needs no change.
- `vlan` tests used `lo0` as the parent, which does not support VLANs; they
  now use an epair.

## [1.0.2] - 2026-08-29

### Added

- `vxlan` package: create VXLAN interfaces, configure VNI, endpoints and
  ports, and manage static peers via the forwarding table.
- IPv6 routing functions in `route`.

### Fixed

- `rt_msghdr` layout used by the routing socket code.
- `ifc.Rename` (`SIOCSIFNAME` takes a pointer to the new name, not the name
  inline).
- IP address deletion.

## [1.0.1] - 2025-10-06

Same code as 1.0.0 (the tag points to the same commit).

## [1.0.0] - 2025-10-06

### Added

- Initial release: interface listing and configuration (MTU, flags, rename,
  promiscuous mode, statistics), `bridge`, `epair`, `vlan`, `lagg`, `tap`,
  `tun`, IPv4/IPv6 address management and routing.

[Unreleased]: https://github.com/zombocoder/go-freebsd-ifc/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/zombocoder/go-freebsd-ifc/compare/v1.0.2...v1.1.0
[1.0.2]: https://github.com/zombocoder/go-freebsd-ifc/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/zombocoder/go-freebsd-ifc/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/zombocoder/go-freebsd-ifc/releases/tag/v1.0.0
