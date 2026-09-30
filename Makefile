PLUGIN_ID := aq-codex-responses-lite
# Release version, always passed explicitly: "make package VERSION=0.3.0".
VERSION ?=
GOOS_NAME := $(shell go env GOOS)
GOARCH_NAME := $(shell go env GOARCH)

# Package target architecture: amd64 (default) or arm64. Cross builds are
# supported in both directions when the matching cross toolchain is installed.
TARGET_ARCH ?= amd64

ifeq ($(filter $(TARGET_ARCH),amd64 arm64),)
$(error unsupported TARGET_ARCH "$(TARGET_ARCH)"; expected amd64 or arm64)
endif

ifeq ($(TARGET_ARCH),$(GOARCH_NAME))
PACKAGE_ENV :=
else ifeq ($(TARGET_ARCH),arm64)
PACKAGE_ENV := GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc
else
PACKAGE_ENV := GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-gnu-gcc
endif

ifeq ($(GOOS_NAME),darwin)
LIB_EXT := dylib
else ifeq ($(GOOS_NAME),windows)
LIB_EXT := dll
else
LIB_EXT := so
endif

LIBRARY := dist/$(PLUGIN_ID).$(LIB_EXT)
RELEASE_ZIP := dist/$(PLUGIN_ID)_$(VERSION)_linux_$(TARGET_ARCH).zip

.PHONY: test vet check check-registry build package registry clean

test:
	go test ./... -count=1

vet:
	go vet ./...

check: test vet

# Validates the CPA Plugin Store manifest (registry.json) structure.
check-registry:
	python3 scripts/check-registry.py registry.json

build:
	mkdir -p dist
	$(BUILD_ENV) go build -buildmode=c-shared -trimpath -o $(LIBRARY) .

package: check
	@test -n "$(VERSION)" || (echo "VERSION is required, e.g. make package VERSION=0.3.0" >&2; exit 1)
	@test "$(GOOS_NAME)" = "linux" || (echo "package must run on Linux" >&2; exit 1)
	$(MAKE) build BUILD_ENV="$(PACKAGE_ENV)"
	cd dist && zip -j "$(notdir $(RELEASE_ZIP))" "$(PLUGIN_ID).so"
	cd dist && sha256sum "$(notdir $(RELEASE_ZIP))" > checksums.txt

# Points registry.json at a published release. Digests are read from that
# release's checksums.txt, or computed from local zips via ARTIFACTS_DIR.
# The release workflow runs this automatically after publishing.
registry:
	@test -n "$(VERSION)" || (echo "VERSION is required, e.g. make registry VERSION=0.3.0" >&2; exit 1)
	scripts/update-registry.sh "$(VERSION)" $(if $(ARTIFACTS_DIR),--artifacts-dir "$(ARTIFACTS_DIR)")

clean:
	rm -rf dist
