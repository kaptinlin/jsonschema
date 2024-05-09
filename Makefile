# Set up GOBIN so that our binaries are installed to ./bin instead of $GOPATH/bin.
PROJECT_ROOT = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
export GOBIN = $(PROJECT_ROOT)/bin

GOLANGCI_LINT_VERSION := $(shell $(GOBIN)/golangci-lint version --format short 2>/dev/null)
REQUIRED_GOLANGCI_LINT_VERSION := $(shell cat .golangci.version)

# Directories containing independent Go modules.
MODULE_DIRS = .

.PHONY: all
all: lint test

.PHONY: clean
clean:
	@rm -rf $(GOBIN)

.PHONY: test
test:
	@$(foreach mod,$(MODULE_DIRS),(cd $(mod) && go test -race ./...) &&) true

.PHONY: lint
lint: golangci-lint tidy-lint

# Install golangci-lint with the required version in GOBIN if it is not already installed.
.PHONY: install-golangci-lint
install-golangci-lint:
    ifneq ($(GOLANGCI_LINT_VERSION),$(REQUIRED_GOLANGCI_LINT_VERSION))
		@echo "[lint] installing golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) since current version is \"$(GOLANGCI_LINT_VERSION)\""
		@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v$(REQUIRED_GOLANGCI_LINT_VERSION)
    endif

.PHONY: golangci-lint
golangci-lint: install-golangci-lint
	@echo "[lint] $(shell $(GOBIN)/golangci-lint version)"
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] golangci-lint: $(mod)" && \
		$(GOBIN)/golangci-lint run --timeout=10m --path-prefix $(mod)) &&) true

.PHONY: tidy-lint
tidy-lint:
	@$(foreach mod,$(MODULE_DIRS), \
		(cd $(mod) && \
		echo "[lint] mod tidy: $(mod)" && \
		go mod tidy && \
		git diff --exit-code -- go.mod go.sum) &&) true
