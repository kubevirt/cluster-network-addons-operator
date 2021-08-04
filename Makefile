all: fmt check

VERSION ?= 99.0.0
export VERSION := $(VERSION)
# Always keep the last released version here
VERSION_REPLACES ?= 0.57.0

DEPLOY_DIR ?= manifests

IMAGE_REGISTRY ?= quay.io/kubevirt
IMAGE_TAG ?= latest
OPERATOR_IMAGE ?= cluster-network-addons-operator
REGISTRY_IMAGE ?= cluster-network-addons-registry

TARGETS = \
	gen-k8s \
	gen-k8s-check \
	goimports \
	goimports-check \
	vet \
	whitespace \
	whitespace-check

export GOFLAGS=-mod=vendor
export GO111MODULE=on

WHAT ?= ./pkg ./cmd ./tools/

GINKGO_EXTRA_ARGS ?=
GINKGO_ARGS ?= --v -r --progress $(GINKGO_EXTRA_ARGS)
GINKGO ?= build/_output/bin/ginkgo

export E2E_TEST_TIMEOUT ?= 3h

E2E_TEST_EXTRA_ARGS ?=
E2E_TEST_ARGS ?= $(strip -test.v -test.timeout $(E2E_TEST_TIMEOUT) -ginkgo.v $(E2E_TEST_EXTRA_ARGS))
E2E_SUITES = \
	test/e2e/lifecycle \
	test/e2e/workflow

BIN_DIR = $(CURDIR)/build/_output/bin/
export GOROOT=$(BIN_DIR)/go/
export GOBIN = $(GOROOT)/bin/
export PATH := $(GOBIN):$(PATH)

OPERATOR_SDK ?= $(BIN_DIR)/operator-sdk

GITHUB_RELEASE ?= $(BIN_DIR)/github-release

GO := $(GOBIN)/go

$(GO):
	hack/install-go.sh $(BIN_DIR)

$(GINKGO): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/github.com/onsi/ginkgo/ginkgo

$(OPERATOR_SDK): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/github.com/operator-framework/operator-sdk/cmd/operator-sdk

$(GITHUB_RELEASE): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/github.com/github-release/github-release

# Make does not offer a recursive wildcard function, so here's one:
rwildcard=$(wildcard $1$2) $(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2))

# Gather needed source files and directories to create target dependencies
directories := $(filter-out ./ ./vendor/ ,$(sort $(dir $(wildcard ./*/))))
all_sources=$(call rwildcard,$(directories),*) $(filter-out $(TARGETS), $(wildcard *))
cmd_sources=$(call rwildcard,cmd/,*.go)
pkg_sources=$(call rwildcard,pkg/,*.go)
apis_sources=$(call rwildcard,pkg/apis,*.go)

fmt: whitespace goimports

goimports: $(cmd_sources) $(pkg_sources)
	$(GO) run ./vendor/golang.org/x/tools/cmd/goimports -w ./pkg ./cmd ./test/ ./tools/
	touch $@

whitespace: $(all_sources)
	./hack/whitespace.sh --fix
	touch $@

check: whitespace-check vet goimports-check gen-k8s-check test/unit prom-rules-verify
	./hack/check.sh

whitespace-check: $(all_sources)
	./hack/whitespace.sh
	touch $@

vet: $(GO) $(cmd_sources) $(pkg_sources)
	$(GO) vet ./pkg/... ./cmd/... ./test/... ./tools/...
	touch $@

goimports-check: $(GO) $(cmd_sources) $(pkg_sources)
	$(GO) run ./vendor/golang.org/x/tools/cmd/goimports -d ./pkg ./cmd
	touch $@

test/unit: $(GINKGO)
	$(GINKGO) $(GINKGO_ARGS) $(WHAT)

manager: $(GO)
	CGO_ENABLED=0 GOOS=linux $(GO) build -o $(BIN_DIR)/$@ ./cmd/...

manifest-templator: $(GO)
	CGO_ENABLED=0 GOOS=linux $(GO) build -o $(BIN_DIR)/$@ ./tools/manifest-templator/...

docker-build: docker-build-operator docker-build-registry

docker-build-operator: manager manifest-templator
	docker build -f build/operator/Dockerfile -t $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG) .

docker-build-registry:
	docker build -f build/registry/Dockerfile -t $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG) .

docker-push: docker-push-operator docker-push-registry

docker-push-operator:
	docker push $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG)

docker-push-registry:
	docker push $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG)

prom-rules-verify:
	hack/prom-rule-ci/verify-rules.sh \
	data/monitoring/prom-rule.yaml \
	hack/prom-rule-ci/prom-rules-tests.yaml

cluster-up:
	./cluster/up.sh

cluster-down:
	./cluster/down.sh

cluster-sync: cluster-operator-push cluster-operator-install

cluster-operator-push:
	./cluster/operator-push.sh

cluster-operator-install:
	./cluster/operator-install.sh

$(E2E_SUITES): $(OPERATOR_SDK)
	unset GOFLAGS && OPERATOR_SDK=$(OPERATOR_SDK) TEST_SUITE=$@ TEST_ARGS="$(E2E_TEST_ARGS)" ./hack/functest.sh

cluster-clean:
	./cluster/clean.sh

# Default images can be found in pkg/components/components.go
gen-manifests: manifest-templator
	VERSION_REPLACES=$(VERSION_REPLACES) \
	DEPLOY_DIR=$(DEPLOY_DIR) \
	CONTAINER_PREFIX=$(IMAGE_REGISTRY) \
	CONTAINER_TAG=$(IMAGE_TAG) \
	MULTUS_IMAGE=$(MULTUS_IMAGE) \
	LINUX_BRIDGE_CNI_IMAGE=$(LINUX_BRIDGE_CNI_IMAGE) \
	KUBEMACPOOL_IMAGE=$(KUBEMACPOOL_IMAGE) \
	MACVTAP_CNI_IMAGE=$(MACVTAP_CNI_IMAGE) \
		./hack/generate-manifests.sh

gen-k8s: $(OPERATOR_SDK) $(apis_sources)
	$(OPERATOR_SDK) generate k8s
	touch $@

gen-k8s-check: $(apis_sources)
	./hack/verify-codegen.sh
	touch $@

bump-kubevirtci:
	rm -rf _kubevirtci
	./hack/bump-kubevirtci.sh

prepare-patch:
	./hack/prepare-release.sh patch
prepare-minor:
	./hack/prepare-release.sh minor
prepare-major:
	./hack/prepare-release.sh major

release: $(GITHUB_RELEASE)
	DESCRIPTION=version/description \
	GITHUB_RELEASE=$(GITHUB_RELEASE) \
	TAG=v$(shell hack/version.sh) \
	  hack/release.sh \
	    manifests/cluster-network-addons/cluster-network-addons.package.yaml \
	    $(shell find manifests/cluster-network-addons/$(shell hack/version.sh) -type f)

vendor: $(GO)
	$(GO) mod tidy
	$(GO) mod vendor

auto-bumper: $(GO)
	PUSH_IMAGES=true $(GO) run $(shell ls tools/bumper/*.go | grep -v test) ${ARGS}

bump-%:
	CNAO_VERSION=${VERSION} ./hack/components/bump-$*.sh
bump-all: bump-nmstate bump-kubemacpool bump-macvtap-cni bump-linux-bridge bump-multus bump-ovs-cni bump-bridge-marker

.PHONY: \
	$(E2E_SUITES) \
	all \
	check \
	cluster-clean \
	cluster-down \
	cluster-operator-install \
	cluster-operator-push \
	cluster-sync \
	cluster-up \
	manager \
	manifests-templator \
	docker-build \
	docker-build-operator \
	docker-build-registry \
	docker-push \
	docker-push-operator \
	docker-push-registry \
	gen-manifests \
	bump-all \
	test/unit \
	bump-kubevirtci \
	prepare-patch \
	prepare-minor \
	prepare-major \
	vendor \
	auto-bumper \
	bump-% \
	bump-all \
	gen-k8s \
	prom-rules-verify \
	release
