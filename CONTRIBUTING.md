# Contributing to go-freebsd-ifc

Thanks for your interest in improving the library! Bug reports, fixes, new
interface types and operations, tests, examples and documentation are all
welcome.

By participating you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
Security problems should be reported privately, see [SECURITY.md](SECURITY.md).

## Table of Contents

- [Getting Started](#getting-started)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Continuous Integration](#continuous-integration)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Architecture Overview](#architecture-overview)
- [Where Help Is Wanted](#where-help-is-wanted)

## Getting Started

### Prerequisites

- FreeBSD 14 or 15, amd64 or arm64
- Go 1.22+ (`pkg install go`)
- The base system C compiler (cgo); no extra headers are needed
- Root (or `doas`) for e2e tests and for examples that change interfaces

A FreeBSD virtual machine is the most comfortable setup: e2e tests create and
destroy interfaces, add addresses and routes, and create a throwaway vnet jail.

### Setup

```sh
git clone https://github.com/<your-username>/go-freebsd-ifc.git
cd go-freebsd-ifc
make build
make test
```

## Running Tests

| Command | Runs as | What it covers |
|---|---|---|
| `make test` | any user | Unit tests: validation, string mappings, struct layouts checked against the C headers. E2E tests skip themselves. |
| `doas make test-e2e` | root | Everything above plus tests that create, configure and destroy real interfaces (`IFCLIB_E2E=1`). |

Running a subset works the usual way:

```sh
doas env IFCLIB_E2E=1 go test -count=1 -run Lagg -v ./lagg
```

E2E tests only touch interfaces they create (kernel-assigned names such as
`epair3a`, `bridge1`), but they run as root and may load kernel modules
(`if_epair`, `if_bridge`, `if_lagg`, `if_vxlan`, ...). Prefer a VM or a test
machine over your workstation, and run packages one at a time (`go test -p 1`)
because some tests observe global state such as the routing table.

## Writing Tests

- **Test first.** Write a test that fails for the right reason, then the fix.
- **Compare with the CLI.** Where possible, check results with `ifconfig`,
  `route` or `netstat` rather than only with the library's own getters, so a
  wrong ioctl or struct layout cannot hide behind a matching wrong reader
  (see `lagg/api_test.go`).
- **Gate e2e tests** with `skipIfNotRoot(t)` and `skipIfNotE2E(t)`.
- **Never** hardcode or modify existing interfaces (`em0`, `vtnet0`, `lo0` as a
  target); create what you need and clean it up with `defer` or `t.Cleanup`,
  so it is removed even when the test fails.
- Struct layout tests (`*_abi_test.go`) compare Go mirrors of kernel structs
  with `sizeof`/`offsetof` from the headers; add one when you mirror a struct.

## Continuous Integration

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on every push to
`main` and every pull request:

- **Lint** (Linux): `gofmt` and `go mod tidy -diff`.
- **FreeBSD 14.5 and 15.1** (QEMU VMs via
  [vmactions/freebsd-vm](https://github.com/vmactions/freebsd-vm)): installs Go,
  runs `go vet`, builds all packages and examples, and runs the full test suite
  as root with `IFCLIB_E2E=1`.

## Code Style

### Go

- Standard Go conventions; `gofmt` is enforced by CI.
- Doc comments on all exported identifiers, with a short example for
  operations that change the system.
- Wrap errors with context (`fmt.Errorf("add port %s to lagg %s: %w", ...)`) and
  keep the typed errors (`ErrNotFound`, `ErrPermission`, ...) reachable with
  `errors.Is`.
- Validate input (names, tags, addresses) before issuing an ioctl.

### cgo

- All cgo lives in `internal/`; public packages are plain Go.
- **Memory**: every `C.CString`/`C.calloc` has a matching `C.free`.
- **Pointers**: never store a Go pointer inside a struct passed to the kernel
  (cgo pointer rules); allocate such buffers with `C.calloc`
  (see `internal/laggops.Get`, `internal/ifops.Rename`).
- **Request structs**: check in `sbin/ifconfig` how the base system issues the
  ioctl. Some take a `struct ifreq`, others (lagg, bridge `SIOCSDRVSPEC`) their
  own struct; getting this wrong usually shows up as `EINVAL` or
  `EPROTONOSUPPORT`.
- Take constants from the C headers (`C.LAGG_PROTO_FAILOVER`) instead of
  hardcoding numbers.

## Submitting Changes

1. Create a branch from `main`.
2. Make the change with tests; update doc comments, README and examples when
   the public API changes, and add an entry to [CHANGELOG.md](CHANGELOG.md)
   under "Unreleased".
3. Run `make check` and `doas make test-e2e` on FreeBSD.
4. Open a pull request; the template has a short checklist. CI must be green.

Commit messages: `type(scope): summary` (e.g. `fix(lagg): pass lagg_reqall
directly to SIOCSLAGG`), a blank line, then what changed and why. Common types
are `feat`, `fix`, `docs`, `test`, `refactor`, `ci`, `chore`.

### Scope

- The library is FreeBSD-specific by design; cross-platform layers are out of
  scope.
- It talks to the kernel directly. Shelling out to `ifconfig`/`route` is only
  acceptable in tests.
- Avoid breaking public API changes within 1.x. When one is unavoidable, call it
  out in the pull request and in the changelog.

## Architecture Overview

```
if/             Interfaces: list, get, MTU, flags, rename, promisc, stats, vnet
bridge/ epair/  Cloned interface types
vlan/ vxlan/
lagg/ tap/ tun/
ip/             IPv4/IPv6 addresses
route/          Routing table (routing socket)
internal/
  syscall/      Sockets, ioctl, kldload, error mapping
  constants/    ioctl numbers and flags from the C headers
  cloneops/     SIOCIFCREATE/SIOCIFDESTROY (+ module autoload)
  ifops/ ipaddr/ routing/ bridgeops/ vlanops/ vxlanops/ laggops/
examples/       Example programs
.github/        CI workflow, issue and PR templates
```

### Adding an operation

1. **Internal op** in the matching `internal/*ops` package: build the request
   struct as `ifconfig` does and issue the ioctl.
2. **Public API** in the top-level package: documented, validated, wrapped
   errors.
3. **Tests**: unit tests for validation, e2e test that checks the effect with
   the CLI, and an ABI test if you mirrored a kernel struct in Go.
4. **Docs**: README / FEATURES, an example if useful, and the changelog.

## Where Help Is Wanted

- **WireGuard** (`if_wg`) interface management
- **GIF/GRE** tunnels
- **CARP** configuration
- **Bridge** STP options and member flags
- **Interface capabilities** (TSO, LRO, checksum offload) get/set
- **Netlink** (FreeBSD 14+) as an alternative to routing sockets

Questions and design discussions are welcome in GitHub issues.
