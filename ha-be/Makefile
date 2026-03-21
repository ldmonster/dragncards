LOCALBIN ?= $(abspath bin)

.PHONY: lint test golangci-lint
lint: ## Run golangci-lint for ha-be module
	@GOLANGCI=$$(command -v golangci-lint 2>/dev/null || echo $(GOLANGCI_LINT)); \
	if [ -z "$$GOLANGCI" ]; then echo "golangci-lint not found"; exit 1; fi; \
	$$GOLANGCI run ./... --fix

test: go-check gotestsum ## Run go tests with race detector and coverage using gotestsum
	$(GOTESTSUM) --format standard-verbose -- -race -cover ./...

.PHONY: go-check
go-check:
	$(call error-if-empty,$(shell command -v go 2>/dev/null),go)


$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# binary files
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GOTESTSUM ?= $(LOCALBIN)/gotestsum

# component versions
GOLANGCI_LINT_VERSION ?= v2.18.0
GOTESTSUM_VERSION ?= v1.13.0

golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

gotestsum: $(GOTESTSUM)
$(GOTESTSUM): $(LOCALBIN)
	$(call go-install-tool,$(GOTESTSUM),gotest.tools/gotestsum,$(GOTESTSUM_VERSION))


# go-install-tool will 'go install' any package with custom target and name of binary.
# $1 - target path (binary path)
# $2 - package url (module path)
# $3 - version
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
	set -e; \
	package=$(2)@$(3); \
	echo "Downloading $${package}"; \
	rm -f $(1) || true; \
	GOBIN=$(LOCALBIN) go install $${package}; \
	mv $(1) $(1)-$(3); \
}; \
ln -sf $(1)-$(3) $(1)
endef

# error-if-empty fails the make target if a required command is missing.
# $1 - evaluated value
# $2 - tool name to show in the error message
define error-if-empty
	@if [ -z "$(1)" ]; then \
		echo "$(2) not installed"; \
		false; \
	fi
endef
