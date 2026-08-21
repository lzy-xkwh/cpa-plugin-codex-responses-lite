PLUGIN_ID := aq-codex-responses-lite
VERSION ?= 0.1.0
GOOS_NAME := $(shell go env GOOS)
GOARCH_NAME := $(shell go env GOARCH)

ifeq ($(GOOS_NAME),darwin)
LIB_EXT := dylib
else ifeq ($(GOOS_NAME),windows)
LIB_EXT := dll
else
LIB_EXT := so
endif

LIBRARY := dist/$(PLUGIN_ID).$(LIB_EXT)
RELEASE_ZIP := dist/$(PLUGIN_ID)_$(VERSION)_linux_amd64.zip

.PHONY: test vet check build package clean

test:
	go test ./... -count=1

vet:
	go vet ./...

check: test vet

build:
	mkdir -p dist
	go build -buildmode=c-shared -trimpath -o $(LIBRARY) .

package: check
	@test "$(GOOS_NAME)" = "linux" || (echo "package must run on Linux" >&2; exit 1)
	@test "$(GOARCH_NAME)" = "amd64" || (echo "package must run on amd64" >&2; exit 1)
	$(MAKE) build
	cd dist && zip -j "$(notdir $(RELEASE_ZIP))" "$(PLUGIN_ID).so"
	cd dist && sha256sum "$(notdir $(RELEASE_ZIP))" > checksums.txt

clean:
	rm -rf dist
