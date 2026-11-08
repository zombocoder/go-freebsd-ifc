# Security Policy

## Supported versions

Security fixes are made on the `main` branch and released as a new patch or
minor version of the latest major release (currently 1.x).

## Reporting a vulnerability

Please **do not open a public issue** for security problems.

Report them privately through GitHub:
[Report a vulnerability](https://github.com/zombocoder/go-freebsd-ifc/security/advisories/new)
(Security tab → "Report a vulnerability").

Include what you can of:

- the affected function or package and the commit you tested;
- the FreeBSD version (`freebsd-version -ku`) and architecture;
- steps or a minimal program that reproduces the problem;
- the impact you expect (memory corruption, privilege issue, wrong interface
  or route modified, ...).

You should get a first response within a week. Once a fix is available, the
advisory is published with credit to the reporter unless you prefer otherwise.

## Scope

This library talks to the kernel through ioctls and routing sockets via cgo
and is typically used by privileged processes. Of particular interest are:

- memory-safety bugs in the cgo layer (`internal/`), such as out-of-bounds
  reads or writes of kernel request structures, wrong struct layouts, leaks,
  or Go pointers handed to the kernel;
- operations that act on a different interface, address, route or vnet than
  requested;
- input that reaches an ioctl without the validation callers would expect
  (interface names, addresses, VLAN/VXLAN identifiers).

Bugs in the FreeBSD network stack itself should be reported to the
[FreeBSD security team](https://www.freebsd.org/security/).
