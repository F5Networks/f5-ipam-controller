PACKAGE  := github.com/F5Networks/f5-ipam-controller

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

all: local-build

test: local-go-test

prod: verify prod-build

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

	docker build --build-arg RUN_TESTS=1 --build-arg BUILD_VERSION=$(BUILD_VERSION) --build-arg BUILD_INFO=$(BUILD_INFO) -t f5-ipam-controller:latest -f build-tools/Dockerfile.$(BASE_OS) .

prod-quick: prod-build-quick

prod-build-quick: pre-build
	@echo "Building without running tests..."
	docker build --build-arg RUN_TESTS=0 --build-arg BUILD_VERSION=$(BUILD_VERSION) --build-arg BUILD_INFO=$(BUILD_INFO) -t f5-ipam-controller:latest -f build-tools/Dockerfile.$(BASE_OS) .

debug: pre-build
	@echo "Building with debug support..."
	docker build  --build-arg RUN_TESTS=0 --build-arg BUILD_VERSION=$(BUILD_VERSION) --build-arg BUILD_INFO=$(BUILD_INFO) -t f5-ipam-controller:latest -f build-tools/Dockerfile.debug .

dev-license: pre-build
	@echo "Running with tests and licenses generated will be in all_attributions.txt..."
	docker build -t fic-attributions:latest -f build-tools/Dockerfile.attribution .

	$(eval id := $(shell docker create fic-attributions:latest))
	docker cp $(id):/opt/all_attributions.txt ./
	docker rm -v $(id)
	docker rmi -f fic-attributions:latest

fmt:
	@echo "Enforcing code formatting using 'go fmt'..."
	$(CURDIR)/build-tools/fmt.sh

vet:
	@echo "Running 'go vet'..."
	$(CURDIR)/build-tools/vet.sh

devel-image:
	docker build --build-arg RUN_TESTS=0 --build-arg BUILD_VERSION=$(BUILD_VERSION) --build-arg BUILD_INFO=$(BUILD_INFO) -t f5-ipam-controller-devel:latest -f build-tools/Dockerfile.$(BASE_OS) .

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

docker-tag:
ifdef tag
	docker tag f5-ipam-controller:latest $(tag)
	docker push $(tag)
else
	@echo "Define a tag to push. Eg: make docker-tag tag=username/f5-ipam-controller:dev"
endif

docker-devel-tag:
	docker push f5-ipam-controller-devel:latest

# one-time html build using a docker container
.PHONY: docker-html
docker-html:
	rm -rf docs/_build
	./build-tools/docker-docs.sh make -C docs/ html
