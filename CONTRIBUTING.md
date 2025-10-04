# Contributing to go-freebsd-ifc

Thank you for your interest in contributing to go-freebsd-ifc! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Feature Requests](#feature-requests)

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help create a positive community environment
- Follow the FreeBSD community standards

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/go-freebsd-ifc.git
   cd go-freebsd-ifc
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/zombocoder/go-freebsd-ifc.git
   ```

## Development Setup

### Requirements

- FreeBSD 12.x or later (tested on 14.x)
- Go 1.19 or later
- C compiler (comes with FreeBSD base system)
- Root/doas access for integration testing

### Build and Test

```bash
# Build all packages
make build

# Run unit tests (no root required)
make test

# Run integration tests (requires root)
doas make test-e2e

# Run all checks
make check

# Generate test coverage
make test-coverage
doas make test-coverage-e2e
```

## How to Contribute

### Types of Contributions

1. **Bug Fixes** - Fix issues in existing code
2. **New Features** - Add new network interface types or operations
3. **Documentation** - Improve docs, examples, or comments
4. **Tests** - Add or improve test coverage
5. **Performance** - Optimize existing code

### Contribution Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make your changes** following our coding standards

3. **Write tests** for your changes:
   - Unit tests for validation logic
   - Integration tests (E2E) for system calls

4. **Run checks**:
   ```bash
   make check           # fmt + vet
   make test            # unit tests
   doas make test-e2e   # integration tests
   ```

5. **Commit your changes** with a clear message

6. **Push to your fork**:
   ```bash
   git push origin feature/my-new-feature
   ```

7. **Open a Pull Request** on GitHub

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Run `go fmt` on all code (or use `make fmt`)
- Run `go vet` to catch common mistakes (or use `make vet`)
- Use `staticcheck` for linting (or use `make lint`)

### Code Organization

```
Public API (top-level packages)
  ↓
Internal Implementation (internal/)
  ↓
System Calls (cgo in internal/syscall)
```

**Rules:**
- Public APIs should be clean, documented, and idiomatic Go
- All cgo code must be in `internal/` packages
- No code duplication - use internal helpers
- Consistent error handling across packages

### Documentation

- **GoDoc comments** for all exported functions, types, and constants
- **Package-level documentation** at the top of main package files
- **Example code** for complex functions
- **Clear error messages** with context

Example:
```go
// SetMTU sets the Maximum Transmission Unit (MTU) for the specified interface.
// The MTU value must be between 68 and 65535 bytes.
//
// This operation requires root privileges.
//
// Example:
//
//	if err := ifc.SetMTU("em0", 9000); err != nil {
//	    log.Fatal(err)
//	}
func SetMTU(name string, mtu int) error {
    // implementation
}
```

### Error Handling

- Use typed errors from `internal/syscall`
- Provide context with `OperationError`
- Support `errors.Is()` and `errors.As()`
- Validate inputs early

Example:
```go
if name == "" {
    return isyscall.NewValidationError("interface name", "cannot be empty")
}

if err := someOperation(); err != nil {
    return isyscall.NewOperationError("set MTU", name, err)
}
```

## Testing Guidelines

### Unit Tests

- Test validation logic
- Test error conditions
- No root privileges required
- Fast and isolated

Example:
```go
func TestValidation(t *testing.T) {
    err := someFunc("")
    if !isyscall.IsValidation(err) {
        t.Errorf("expected validation error, got %v", err)
    }
}
```

### Integration Tests (E2E)

- Test actual system calls
- Requires root privileges
- Gated with `IFCLIB_E2E=1`
- Clean up resources in defer/cleanup

Example:
```go
func TestCreateBridge(t *testing.T) {
    if os.Getenv("IFCLIB_E2E") != "1" {
        t.Skip("Skipping E2E test (set IFCLIB_E2E=1 to run)")
    }
    if os.Geteuid() != 0 {
        t.Skip("Skipping test: requires root")
    }

    br, err := bridge.Create()
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }
    defer bridge.Destroy(br)

    // Test operations
}
```

### Test Coverage

- Aim for high coverage of public APIs
- Test both success and failure paths
- Include edge cases
- Document why tests are skipped

## Commit Messages

### Format

```
<type>: <short summary>

<optional detailed description>

<optional footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or fixes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build/tooling changes

### Examples

```
feat: add VXLAN interface support

Implements VXLAN (Virtual Extensible LAN) interface creation and
configuration. Adds vxlan package with Create, Destroy, and Configure
functions.

Closes #123
```

```
fix: handle EEXIST in bridge member addition

Bridge.AddMember now returns nil when member already exists,
making the operation idempotent as documented.

Fixes #456
```

## Pull Request Process

1. **Ensure all tests pass**:
   ```bash
   make test
   doas make test-e2e
   ```

2. **Update documentation** if needed:
   - README.md for API changes
   - FEATURES.md for new features
   - examples/ for new functionality
   - GoDoc comments

3. **Add entry to CHANGELOG.md** under "Unreleased"

4. **Fill out PR template** with:
   - Description of changes
   - Related issue numbers
   - Testing performed
   - FreeBSD version tested on

5. **Request review** from maintainers

6. **Address review feedback** promptly

7. **Squash commits** if requested

### PR Checklist

- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Commit messages are clear
- [ ] No unnecessary files included

## Reporting Bugs

### Before Reporting

- Check existing issues
- Verify on latest version
- Test on clean FreeBSD installation

### Bug Report Template

```markdown
**Environment:**
- FreeBSD version: (e.g., 14.1-RELEASE)
- Go version: (e.g., 1.21.0)
- Library version: (e.g., v1.0.0)

**Description:**
Clear description of the bug

**Steps to Reproduce:**
1. Step one
2. Step two
3. Step three

**Expected Behavior:**
What should happen

**Actual Behavior:**
What actually happens

**Code Sample:**
```go
// Minimal code to reproduce
```

**Error Output:**
```
Error messages or stack trace
```

**Additional Context:**
Any other relevant information
```

## Feature Requests

### Before Requesting

- Check if already requested
- Consider if it fits project scope
- Think about FreeBSD compatibility

### Feature Request Template

```markdown
**Feature Description:**
Clear description of the proposed feature

**Use Case:**
Why is this feature needed?

**Proposed API:**
```go
// Example API design
```

**Alternatives Considered:**
Other approaches you've thought about

**FreeBSD Compatibility:**
Which FreeBSD versions support this?

**Additional Context:**
Any other relevant information
```

## Development Tips

### Testing in Jails

To avoid breaking your host networking:

```bash
# Create test jail
doas jail -c name=ifctest path=/usr/jails/ifctest

# Run tests in jail
doas jexec ifctest make test-e2e
```

### Debugging

- Use `doas ktrace -i` for system call tracing
- Check `dmesg` for kernel messages
- Use `ifconfig -v` to verify interface state
- Test with `tcpdump` for network traffic

### Common Pitfalls

1. **Forgetting root checks** - Always check `os.Geteuid() != 0`
2. **Not cleaning up** - Use `defer` for resource cleanup
3. **Hardcoding interface names** - Use created/tested interfaces only
4. **Platform assumptions** - Remember this is FreeBSD-specific
5. **Skipping E2E tests** - Integration tests are critical

## Questions?

- Open a GitHub issue for questions
- Check existing documentation
- Review example programs in `examples/`

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to go-freebsd-ifc!**
