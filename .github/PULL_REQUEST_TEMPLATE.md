## What and why

<!-- What does this change do, and what problem does it solve? Link issues with "Fixes #123". -->

## How it was tested

<!-- CI runs everything on FreeBSD 14.x and 15.x VMs. If you also tested locally, say how, e.g.:
     doas env IFCLIB_E2E=1 go test -p 1 ./... on FreeBSD 15.1 -->

## Checklist

- [ ] New behavior is covered by tests (a failing test first, then the fix)
- [ ] Tests that create interfaces are gated by `IFCLIB_E2E=1` + root, use kernel-assigned names and clean up with `defer`/`t.Cleanup`
- [ ] Where possible, tests check the result with `ifconfig`/`route` rather than only with the library itself
- [ ] cgo code frees every C allocation and passes no Go pointers inside structs handed to the kernel
- [ ] Public API changes are documented (doc comments, README, examples) and listed in `CHANGELOG.md`
- [ ] `gofmt -l .` prints nothing
