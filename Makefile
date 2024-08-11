# Go paths
SERVER_SRC=cmd/server/main.go
CLIENT_SRC=cmd/client/main.go

# Binaries
SERVER_BIN=dyndns-server
CLIENT_BIN=dyndns-client

# Targets
TARGETS= \
    darwin-amd64 \
    darwin-arm64 \
    linux-386 \
    linux-amd64 \
    linux-arm64 \
    linux-riscv64 \
    linux-ppc64 \
    windows-amd64 \
    windows-386 \
    windows-arm \
    windows-arm64

# Output directory
DIST_DIR=dist

# Default target
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  all           - Build and package all targets"
	@echo "  build-all     - Build binaries for all targets"
	@echo "  pack-all      - Package binaries for all targets"
	@echo "  clean         - Remove the dist directory"
	@echo "  list-targets  - List all supported targets"
	@echo "  [target]      - Build and package a specific target (e.g., make linux-amd64)"
	@echo "  build-[target] - Build binaries for a specific target (e.g., make build-linux-amd64)"
	@echo "  pack-[target]  - Package binaries for a specific target (e.g., make pack-linux-amd64)"

# List all supported targets
.PHONY: list-targets
list-targets:
	@echo "Supported targets:"
	@for target in $(TARGETS); do \
		echo "  $$target"; \
	done

# Build and package all
.PHONY: all
all: build-all pack-all

# Build all binaries
.PHONY: build-all
build-all: $(addprefix build-,$(TARGETS))

# Package all binaries
.PHONY: pack-all
pack-all: $(addprefix pack-,$(TARGETS))

# Build and package each target
$(TARGETS):
	$(MAKE) build-$@
	$(MAKE) pack-$@

# Create dist directory
$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Build binaries for each target
build-%: $(DIST_DIR)
	GOARCH=$(shell echo $* | cut -d- -f2) \
	GOOS=$(shell echo $* | cut -d- -f1) \
	CGO_ENABLED=0 \
	go build -o $(DIST_DIR)/$(SERVER_BIN)-$*$(if $(findstring windows,$*),.exe) $(SERVER_SRC)
	CGO_ENABLED=0 \
	go build -o $(DIST_DIR)/$(CLIENT_BIN)-$*$(if $(findstring windows,$*),.exe) $(CLIENT_SRC)
	@echo "Built $(DIST_DIR)/$(SERVER_BIN)-$*$(if $(findstring windows,$*),.exe) and $(DIST_DIR)/$(CLIENT_BIN)-$*$(if $(findstring windows,$*),.exe) for $*."

# Package binaries for each target
pack-%: build-%
	tar -czvf $(DIST_DIR)/dyndns-$*.tar.gz -C $(DIST_DIR) $(SERVER_BIN)-$*$(if $(findstring windows,$*),.exe) $(CLIENT_BIN)-$*$(if $(findstring windows,$*),.exe)
	rm $(DIST_DIR)/$(SERVER_BIN)-$*$(if $(findstring windows,$*),.exe) $(DIST_DIR)/$(CLIENT_BIN)-$*$(if $(findstring windows,$*),.exe)
	@echo "Packaged $(DIST_DIR)/dyndns-$*.tar.gz."

# Clean: remove the dist directory
.PHONY: clean
clean:
	rm -rf $(DIST_DIR)
	@echo "Cleaned up $(DIST_DIR) directory."
