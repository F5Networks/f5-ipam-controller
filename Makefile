PACKAGE  := github.com/f5devcentral/f5-ipam-controller

BASE     := $(GOPATH)/src/$(PACKAGE)
GOOS     = $(shell go env GOOS)
GOARCH   = $(shell go env GOARCH)
GOBIN    = $(GOPATH)/bin/$(GOOS)-$(GOARCH)

NEXT_VERSION := $(shell ./build-tools/version-tool.py version)
export BUILD_VERSION := $(if $(BUILD_VERSION),$(BUILD_VERSION),$(NEXT_VERSION))
export BUILD_INFO := $(shell ./build-tools/version-tool.py build-info)

GO_BUILD_FLAGS=-v -ldflags "-extldflags \"-static\" -X main.version=$(BUILD_VERSION) -X main.buildInfo=$(BUILD_INFO)"

# Allow users to pass in BASE_OS build options (debian or rhel)
BASE_OS ?= debian

# This is for generating licences for vendor packages. Set the environment variable to true to generate the all_attributions.txt
LICENSE ?= false

all: local-build

test: local-go-test

prod: prod-build

verify: fmt vet

docs: _docs

clean:
	docker rmi f5-ipam-controller-devel || true
	docker rmi f5-ipam-controller-debug || true
	@echo "Did not clean local go workspace"

info:
	env


############################################################################
# NOTE:
#   The following targets are supporting targets for the publicly maintained
#   targets above. Publicly maintained targets above are always provided.
############################################################################

# Depend on always-build when inputs aren't known
.PHONY: always-build

# Disable builtin implicit rules
.SUFFIXES:

local-go-test: local-build check-gopath
	ginkgo ./pkg/... ./cmd/...

local-build: check-gopath
	GOBIN=$(GOBIN) go install $(GO_BUILD_FLAGS) ./pkg/... ./cmd/...

check-gopath:
	@if [ "$(BASE)" != "$(CURDIR)" ]; then \
	  echo "Source directory must be in valid GO workspace."; \
	  echo "Check GOPATH?"; \
	  false; \
	fi

pre-build:
	git status
	git describe --all --long --always

prod-build: pre-build
	@echo "Building with minimal instrumentation..."
	LICENSE=$(LICENSE) RUN_TESTS=1 BASE_OS=$(BASE_OS) BASE_OS=$(BASE_OS) $(CURDIR)/build-tools/build-image.sh

prod-quick: prod-build-quick

prod-build-quick: pre-build
	@echo "Building without running tests..."
	LICENSE=$(LICENSE) RUN_TESTS=0 BASE_OS=$(BASE_OS) $(CURDIR)/build-tools/build-image.sh

debug: pre-build
	@echo "Building with debug support..."
	LICENSE=$(LICENSE) DEBUG=0 RUN_TESTS=0 BASE_OS=$(BASE_OS) $(CURDIR)/build-tools/build-image.sh


dev-licences: pre-build
	@echo "Building without running tests..."
	LICENSE=true RUN_TESTS=0 BASE_OS=$(BASE_OS) $(CURDIR)/build-tools/build-image.sh

fmt:
	@echo "Enforcing code formatting using 'go fmt'..."
	$(CURDIR)/build-tools/fmt.sh

vet:
	@echo "Running 'go vet'..."
	$(CURDIR)/build-tools/vet.sh

devel-image:
	TARGET=builder BASE_OS=$(BASE_OS) ./build-tools/build-image.sh

# Enable certain funtionalities only on a developer build
dev-patch:
	# Place Holder

reset-dev-patch:
	# Place Holder

# Build devloper image
dev: dev-patch prod-quick reset-dev-patch

# Docs
#
doc-preview:
	rm -rf docs/_build
	DOCKER_RUN_ARGS="-p 127.0.0.1:8000:8000" \
	  ./build-tools/docker-docs.sh make -C docs preview

_docs: always-build
	./build-tools/docker-docs.sh ./build-tools/make-docs.sh

docker-test:
	rm -rf docs/_build
	./build-tools/docker-docs.sh ./build-tools/make-docs.sh

# one-time html build using a docker container
.PHONY: docker-html
docker-html:
	rm -rf docs/_build
	./build-tools/docker-docs.sh make -C docs/ html
