.PHONY: all build test test-unit test-e2e clean examples install fmt vet lint check help docs

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build parameters
BUILD_FLAGS=-v
TEST_FLAGS=-v

# Example binaries (only those with main.go)
EXAMPLES=examples/list examples/list-vlans examples/vlan-demo examples/tap-tun-demo examples/lagg-demo examples/ipv6-routing examples/iface-config examples/ifstats examples/comprehensive-demo examples/net-bridge-up examples/ip-addr examples/route-default
EXAMPLE_BINS=$(EXAMPLES:%=%/main)


all: check build

##@ Build

build: ## Build all packages
	@echo "Building packages..."
	@$(GOBUILD) $(BUILD_FLAGS) ./...

examples: $(EXAMPLE_BINS) ## Build all example programs

examples/%/main: examples/%/main.go
	@echo "Building $@..."
	@$(GOBUILD) $(BUILD_FLAGS) -o $@ $<

install: ## Install the library
	@echo "Installing packages..."
	@$(GOCMD) install ./...

##@ Testing

test: test-unit ## Run all tests (alias for test-unit)

test-unit: ## Run unit tests (no root required)
	@echo "Running unit tests..."
	@$(GOTEST) $(TEST_FLAGS) ./...

test-e2e: ## Run integration tests (requires root)
	@echo "Running integration tests (requires root)..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "Error: Integration tests require root privileges"; \
		echo "Run: sudo make test-e2e"; \
		exit 1; \
	fi
	@IFCLIB_E2E=1 $(GOTEST) $(TEST_FLAGS) ./...

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-e2e: ## Generate E2E test coverage report (requires root)
	@echo "Generating E2E coverage report (requires root)..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "Error: E2E coverage requires root privileges"; \
		echo "Run: sudo make test-coverage-e2e"; \
		exit 1; \
	fi
	@IFCLIB_E2E=1 $(GOTEST) -coverprofile=coverage-e2e.out -covermode=atomic ./...
	@$(GOCMD) tool cover -html=coverage-e2e.out -o coverage-e2e.html
	@echo "E2E coverage report generated: coverage-e2e.html"

test-race: ## Run tests with race detector (may show ASLR/PIE warnings)
	@echo "Running tests with race detector..."
	@$(GOTEST) -race -v ./...

##@ Code Quality

fmt: ## Format all Go source files
	@echo "Formatting code..."
	@$(GOFMT) ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOVET) ./...

check: fmt vet ## Run fmt and vet

lint: ## Run static analysis (requires staticcheck)
	@echo "Running staticcheck..."
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not found. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

##@ Documentation

docs: ## Start local documentation server
	@echo "Starting godoc server on http://localhost:6060"
	@echo "Visit: http://localhost:6060/pkg/github.com/zombocoder/go-freebsd-ifc/"
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060; \
	else \
		echo "godoc not found. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

docs-view: ## View package documentation (using go doc)
	@$(GOCMD) doc -all ./if
	@echo "\nTo view other packages:"
	@echo "  go doc ./bridge"
	@echo "  go doc ./epair"
	@echo "  go doc ./vlan"
	@echo "  go doc ./ip"
	@echo "  go doc ./route"

##@ Maintenance

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(EXAMPLE_BINS)
	@rm -f coverage.out coverage.html coverage-e2e.out coverage-e2e.html
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOGET) -v -d ./...

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	@$(GOCMD) mod tidy

##@ Information

list-examples: ## List all example programs
	@echo "Available examples:"
	@echo "  examples/list               - List all network interfaces (no root)"
	@echo "  examples/list-vlans         - List all VLAN interfaces (no root)"
	@echo "  examples/iface-config       - Interface configuration tool (show/mtu/up/down/promisc)"
	@echo "  examples/ifstats            - Interface statistics viewer (show/list/watch)"
	@echo "  examples/vlan-demo          - VLAN management demo (requires root)"
	@echo "  examples/tap-tun-demo       - TAP/TUN management demo (requires root)"
	@echo "  examples/lagg-demo          - LAGG link aggregation demo (requires root)"
	@echo "  examples/ipv6-routing       - IPv6 routing management (requires root)"
	@echo "  examples/net-bridge-up      - Create bridge + epair (requires root)"
	@echo "  examples/ip-addr            - IP address management (requires root)"
	@echo "  examples/route-default      - IPv4 routing management (requires root)"
	@echo "  examples/comprehensive-demo - All features demo (requires root)"
	@echo ""
	@echo "To run an example:"
	@echo "  go run examples/list/main.go"
	@echo "  go run examples/iface-config/main.go show lo0"
	@echo "  go run examples/ifstats/main.go show em0"
	@echo "  go run examples/ifstats/main.go watch em0 2"
	@echo "  go run examples/list-vlans/main.go"
	@echo "  go run examples/vlan-demo/main.go list"
	@echo "  doas go run examples/vlan-demo/main.go create 100 em0"
	@echo "  go run examples/tap-tun-demo/main.go list"
	@echo "  doas go run examples/tap-tun-demo/main.go create-tap"
	@echo "  go run examples/lagg-demo/main.go list"
	@echo "  doas go run examples/lagg-demo/main.go create lacp"
	@echo "  doas go run examples/iface-config/main.go mtu em0 9000"
	@echo "  doas go run examples/ipv6-routing/main.go add-default em0 fe80::1"
	@echo "  doas go run examples/comprehensive-demo/main.go"
	@echo "  doas go run examples/net-bridge-up/main.go"

packages: ## List all packages
	@echo "Public packages:"
	@echo "  if      - Interface management"
	@echo "  bridge  - Bridge management"
	@echo "  epair   - Epair management"
	@echo "  vlan    - VLAN management"
	@echo "  lagg    - Link aggregation (LAGG) management"
	@echo "  tap     - TAP interface management"
	@echo "  tun     - TUN interface management"
	@echo "  ip      - IP address management"
	@echo "  route   - Routing management"
	@echo ""
	@echo "Internal packages (implementation):"
	@echo "  internal/syscall   - Socket & ioctl wrappers"
	@echo "  internal/constants - All magic numbers"
	@echo "  internal/ifops     - Interface operations"
	@echo "  internal/bridgeops - Bridge operations"
	@echo "  internal/cloneops  - Clone interface ops"
	@echo "  internal/vlanops   - VLAN operations"
	@echo "  internal/laggops   - LAGG operations"
	@echo "  internal/ipaddr    - IP address ops"
	@echo "  internal/routing   - Routing ops"

version: ## Show Go version and module info
	@echo "Go version:"
	@$(GOCMD) version
	@echo ""
	@echo "Module info:"
	@$(GOCMD) list -m

##@ Help

help: ## Display this help message
	@echo "================================================================================"
	@echo " go-freebsd-ifc - FreeBSD Network Interface Control Library"
	@echo "================================================================================"
	@echo ""
	@echo "Build"
	@echo "  build              Build all packages"
	@echo "  examples           Build all example programs"
	@echo "  install            Install the library"
	@echo ""
	@echo "Testing"
	@echo "  test               Run all tests (alias for test-unit)"
	@echo "  test-unit          Run unit tests (no root required)"
	@echo "  test-e2e           Run integration tests (requires root)"
	@echo "  test-coverage      Generate test coverage report"
	@echo "  test-coverage-e2e  Generate E2E test coverage report (requires root)"
	@echo "  test-race          Run tests with race detector (may show ASLR/PIE warnings)"
	@echo ""
	@echo "Code Quality"
	@echo "  fmt                Format all Go source files"
	@echo "  vet                Run go vet"
	@echo "  check              Run fmt and vet"
	@echo "  lint               Run static analysis (requires staticcheck)"
	@echo ""
	@echo "Documentation"
	@echo "  docs               Start local documentation server"
	@echo "  docs-view          View package documentation (using go doc)"
	@echo ""
	@echo "Maintenance"
	@echo "  clean              Clean build artifacts"
	@echo "  deps               Download dependencies"
	@echo "  tidy               Tidy go.mod"
	@echo ""
	@echo "Information"
	@echo "  list-examples      List all example programs"
	@echo "  packages           List all packages"
	@echo "  version            Show Go version and module info"
	@echo ""
	@echo "Help"
	@echo "  help               Display this help message"
	@echo ""
	@echo "================================================================================"
	@echo "Quick Start:"
	@echo "  make build           # Build all packages"
	@echo "  make test            # Run tests"
	@echo "  make docs            # Start documentation server"
	@echo "  make help            # Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make list-examples   # List all examples"
	@echo "  make examples        # Build example binaries"
	@echo ""
	@echo "Documentation:"
	@echo "  https://pkg.go.dev/github.com/zombocoder/go-freebsd-ifc"
	@echo "================================================================================"
