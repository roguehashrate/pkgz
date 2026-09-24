VERSION    := 1.3.0
BINARY     := pkgz
OUTPUT_DIR := build
LDFLAGS    := -s -w

GO        := go
GOCMD     := $(GO) build -ldflags="$(LDFLAGS)"

HOST_OS   := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_RAW  := $(shell uname -m)

ifeq ($(patsubst x86_64%,x86_64,$(HOST_RAW)),x86_64)
HOST_ARCH := amd64
else ifeq ($(patsubst amd64%,amd64,$(HOST_RAW)),amd64)
HOST_ARCH := amd64
else ifeq ($(patsubst aarch64%,aarch64,$(HOST_RAW)),aarch64)
HOST_ARCH := arm64
else ifeq ($(patsubst arm64%,arm64,$(HOST_RAW)),arm64)
HOST_ARCH := arm64
else ifeq ($(patsubst i386%,i386,$(HOST_RAW)),i386)
HOST_ARCH := 386
else ifeq ($(patsubst i686%,i686,$(HOST_RAW)),i686)
HOST_ARCH := 386
else ifeq ($(patsubst arm%,arm,$(HOST_RAW)),arm)
HOST_ARCH := arm
else
HOST_ARCH := amd64
endif

.PHONY: all build dev dist test test-race vet fmt lint run install clean help

all: build

# Host build only (no packaging)
build:
	$(GOCMD) -o $(BINARY) .

# Host build + .tar.gz and .deb packages (equivalent to build.sh --dev)
dev: dist

dist:
	$(MAKE) build-host
	$(MAKE) tarball
	$(MAKE) deb

build-host:
	$(MAKE) build-$(HOST_OS)-$(HOST_ARCH)

# Target a specific OS/ARCH, e.g. make build-linux-amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOCMD) -o $(OUTPUT_DIR)/linux/amd64/$(BINARY)
build-linux-386:
	GOOS=linux GOARCH=386   $(GOCMD) -o $(OUTPUT_DIR)/linux/386/$(BINARY)
build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GOCMD) -o $(OUTPUT_DIR)/linux/arm64/$(BINARY)
build-linux-arm:
	GOOS=linux GOARCH=arm   $(GOCMD) -o $(OUTPUT_DIR)/linux/arm/$(BINARY)

tarball:
	@rm -rf $(OUTPUT_DIR)/pkgz-dist
	@mkdir -p $(OUTPUT_DIR)/pkgz-dist
	@cp $(OUTPUT_DIR)/$(HOST_OS)/$(HOST_ARCH)/$(BINARY) $(OUTPUT_DIR)/pkgz-dist/
	@cp LICENSE README.md $(OUTPUT_DIR)/pkgz-dist/ 2>/dev/null || true
	@tar -czf $(OUTPUT_DIR)/pkgz-v$(VERSION)-$(HOST_OS)-$(HOST_ARCH).tar.gz -C $(OUTPUT_DIR) pkgz-dist
	@rm -rf $(OUTPUT_DIR)/pkgz-dist
	@echo "Created: $(OUTPUT_DIR)/pkgz-v$(VERSION)-$(HOST_OS)-$(HOST_ARCH).tar.gz"

deb:
	@arch="$(HOST_ARCH)"; \
	case "$$arch" in \
		amd64) deb_arch="amd64" ;; \
		386|i386|i686) deb_arch="i386" ;; \
		arm64) deb_arch="arm64" ;; \
		arm*) deb_arch="armhf" ;; \
		*) echo "Unknown arch: $$arch"; exit 1 ;; \
	esac; \
	stage="$(OUTPUT_DIR)/deb-stage"; \
	rm -rf "$$stage"; \
	mkdir -p "$$stage/DEBIAN" "$$stage/usr/local/bin"; \
	cp $(OUTPUT_DIR)/$(HOST_OS)/$(HOST_ARCH)/$(BINARY) "$$stage/usr/local/bin/$(BINARY)"; \
	printf 'Package: pkgz\nVersion: $(VERSION)\nSection: utils\nPriority: optional\nArchitecture: %s\nMaintainer: roguehashrate <https://github.com/roguehashrate>\nDescription: Fast, extensible CLI tool for managing multiple package types on Linux.\n' "$$deb_arch" > "$$stage/DEBIAN/control"; \
	dpkg-deb --root-owner-group --build "$$stage" "$(OUTPUT_DIR)/pkgz_$(VERSION)_$$deb_arch.deb" >/dev/null; \
	rm -rf "$$stage"; \
	echo "Created: $(OUTPUT_DIR)/pkgz_$(VERSION)_$$deb_arch.deb"

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w main.go main_test.go pkg/

lint:
	command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed; skipping"

run:
	$(GO) run .

install: build
	install -m 0755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

PREFIX ?= /usr/local

clean:
	rm -rf $(BINARY) $(OUTPUT_DIR)

help:
	@echo "pkgz Makefile (v$(VERSION))"
	@echo ""
	@echo "Targets:"
	@echo "  all          host build (default)"
	@echo "  build        build $(BINARY) for the host platform"
	@echo "  dev          host build + .tar.gz + .deb  (like build.sh --dev)"
	@echo "  dist         alias for dev"
	@echo "  build-linux-{amd64,386,arm64,arm}  cross-compile a specific target"
	@echo "  tarball      package current host build as a .tar.gz"
	@echo "  deb          package current host build as a .deb"
	@echo "  test         run tests"
	@echo "  test-race    run tests with the race detector"
	@echo "  vet          run go vet"
	@echo "  fmt          format all Go sources"
	@echo "  lint         run golangci-lint if available"
	@echo "  run          run the program"
	@echo "  install      install to PREFIX/bin (default $(PREFIX))"
	@echo "  clean        remove the binary and $(OUTPUT_DIR)"