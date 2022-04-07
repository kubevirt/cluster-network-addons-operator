all: fmt check

VERSION ?= 99.0.0
export VERSION := $(VERSION)
# Always keep the last released version here
VERSION_REPLACES ?= 0.74.0

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
	test/e2e/workflow \
	test/e2e/monitoring

BIN_DIR = $(CURDIR)/build/_output/bin/
export GOROOT=$(BIN_DIR)/go/
export GOBIN = $(GOROOT)/bin/
export PATH := $(GOBIN):$(PATH)

OPERATOR_SDK ?= $(BIN_DIR)/operator-sdk

GITHUB_RELEASE ?= $(BIN_DIR)/github-release

CONTROLLER_GEN ?= $(BIN_DIR)/controller-gen

GO := $(GOBIN)/go

$(GO):
	hack/install-go.sh $(BIN_DIR)

$(GINKGO): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/github.com/onsi/ginkgo/ginkgo

$(OPERATOR_SDK): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/github.com/operator-framework/operator-sdk/cmd/operator-sdk

$(GITHUB_RELEASE): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/github.com/github-release/github-release

$(CONTROLLER_GEN): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen

lint: $(GO) 
	GOFLAGS=-mod=mod $(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2 run --fix

.PHONY: check
check: lint gen-k8s-check test/unit prom-rules-verify
	./hack/check.sh

.PHONY: test/unit
test/unit: $(GINKGO)
	$(GINKGO) $(GINKGO_ARGS) $(WHAT)

.PHONY: manager
manager: $(GO)
	CGO_ENABLED=0 GOOS=linux $(GO) build -o $(BIN_DIR)/$@ ./cmd/...

.PHONY: manifests-templator
manifest-templator: $(GO)
	CGO_ENABLED=0 GOOS=linux $(GO) build -o $(BIN_DIR)/$@ ./tools/manifest-templator/...

.PHONY: docker-build
docker-build: docker-build-operator docker-build-registry

.PHONY: docker-build-operator
docker-build-operator: manager manifest-templator
	docker build -f build/operator/Dockerfile -t $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG) .

.PHONY: docker-build-registry
docker-build-registry:
	docker build -f build/registry/Dockerfile -t $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG) .

.PHONY: docker-push
docker-push: docker-push-operator docker-push-registry

.PHONY: docker-push-operator
docker-push-operator:
	docker push $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG)

.PHONY: docker-push-registry
docker-push-registry:
	docker push $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG)

.PHONY: prom-rules-verify
prom-rules-verify:
	hack/prom-rule-ci/verify-rules.sh \
	data/monitoring/prom-rule.yaml \
	hack/prom-rule-ci/prom-rules-tests.yaml

.PHONY: cluster-up
cluster-up:
	./cluster/up.sh

.PHONY: cluster-down
cluster-down:
	./cluster/down.sh

.PHONY: cluster-sync
cluster-sync: cluster-operator-push cluster-operator-install

.PHONY: cluster-operator-push
cluster-operator-push:
	./cluster/operator-push.sh

.PHONY: cluster-operator-install
cluster-operator-install:
	./cluster/operator-install.sh

.PHONY: $(E2E_SUITES)
$(E2E_SUITES): $(GINKGO)
	GINKGO=$(GINKGO) GO=$(GO) TEST_SUITE=$@ TEST_ARGS="$(E2E_TEST_ARGS)" ./hack/functest.sh

.PHONY: cluster-clean
cluster-clean:
	./cluster/clean.sh

# Default images can be found in pkg/components/components.go
.PHONY: gen-manifests
gen-manifests: manifest-templator
	VERSION_REPLACES=$(VERSION_REPLACES) \
	DEPLOY_DIR=$(DEPLOY_DIR) \
	CONTAINER_PREFIX=$(IMAGE_REGISTRY) \
	CONTAINER_TAG=$(IMAGE_TAG) \
	MULTUS_IMAGE=$(MULTUS_IMAGE) \
	LINUX_BRIDGE_CNI_IMAGE=$(LINUX_BRIDGE_CNI_IMAGE) \
	KUBEMACPOOL_IMAGE=$(KUBEMACPOOL_IMAGE) \
	MACVTAP_CNI_IMAGE=$(MACVTAP_CNI_IMAGE) \
	KUBE_RBAC_PROXY_IMAGE=$(KUBE_RBAC_PROXY_IMAGE) \
		./hack/generate-manifests.sh

.PHONY: gen-k8s
gen-k8s: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: gen-k8s-check
gen-k8s-check:
	./hack/verify-codegen.sh

.PHONY: bump-kubevirtci
bump-kubevirtci:
	rm -rf _kubevirtci
	./hack/bump-kubevirtci.sh

.PHONY: prepare-patch
prepare-patch:
	./hack/prepare-release.sh patch
.PHONY: prepare-minor
prepare-minor:
	./hack/prepare-release.sh minor
.PHONY: prepare-major
prepare-major:
	./hack/prepare-release.sh major

.PHONY: release-notes
release-notes:
	hack/render-release-notes.sh $(WHAT)

.PHONY: release
release: $(GITHUB_RELEASE)
	GITHUB_RELEASE=$(GITHUB_RELEASE) \
	TAG=v$(shell hack/version.sh) \
	  hack/release.sh \
	    manifests/cluster-network-addons/cluster-network-addons.package.yaml \
	    $(shell find manifests/cluster-network-addons/$(shell hack/version.sh) -type f)

.PHONY: vendor
vendor: $(GO)
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: auto-bumper
auto-bumper: $(GO)
	PUSH_IMAGES=true $(GO) run $(shell ls tools/bumper/*.go | grep -v test) ${ARGS}

.PHONY: bump-%
bump-%:
	CNAO_VERSION=${VERSION} ./hack/components/bump-$*.sh
.PHONY: bump-all
bump-all: bump-kubemacpool bump-macvtap-cni bump-linux-bridge bump-multus bump-ovs-cni bump-bridge-marker

.PHONY: generate-doc
generate-doc:
	go run ./tools/metricsdocs > docs/metrics.md
