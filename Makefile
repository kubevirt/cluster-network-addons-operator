all: fmt check

VERSION ?= 99.0.0
export VERSION := $(VERSION)
# Always keep the last released version here
VERSION_REPLACES ?= 0.93.0

DEPLOY_DIR ?= manifests

IMAGE_REGISTRY ?= quay.io/kubevirt
IMAGE_TAG ?= latest
OPERATOR_IMAGE ?= cluster-network-addons-operator
REGISTRY_IMAGE ?= cluster-network-addons-registry
export OCI_BIN ?= $(shell if podman ps >/dev/null 2>&1; then echo podman; elif docker ps >/dev/null 2>&1; then echo docker; fi)
TLS_SETTING := $(if $(filter $(OCI_BIN),podman),--tls-verify=false,)

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
GO_VERSION = $(shell hack/go-version.sh)

WHAT ?= ./pkg/... ./cmd/... ./tools/...

export E2E_TEST_TIMEOUT ?= 3h

E2E_TEST_EXTRA_ARGS ?=
E2E_TEST_ARGS ?= $(strip -test.v -test.timeout=$(E2E_TEST_TIMEOUT) -ginkgo.timeout=$(E2E_TEST_TIMEOUT) $(E2E_TEST_EXTRA_ARGS))
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

MONITORING_LINTER ?= $(BIN_DIR)/monitoringlinter

GO := $(GOBIN)/go

$(GO):
	hack/install-go.sh $(BIN_DIR)

$(OPERATOR_SDK): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/github.com/operator-framework/operator-sdk/cmd/operator-sdk

$(GITHUB_RELEASE): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/github.com/github-release/github-release

$(CONTROLLER_GEN): $(GO) go.mod
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install ./vendor/sigs.k8s.io/controller-tools/cmd/controller-gen

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

test/unit: $(GO)
	$(GO) test $(WHAT)

manager: $(GO)
	CGO_ENABLED=0 GOOS=linux $(GO) build -o $(BIN_DIR)/$@ ./cmd/...

manifest-templator: $(GO)
	CGO_ENABLED=0 GOOS=linux $(GO) build -o $(BIN_DIR)/$@ ./tools/manifest-templator/...

docker-build: docker-build-operator docker-build-registry

docker-build-operator: manager manifest-templator
	$(OCI_BIN) build -f build/operator/Dockerfile -t $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG) .

docker-build-registry:
	$(OCI_BIN) build -f build/registry/Dockerfile -t $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG) .

docker-push: docker-push-operator docker-push-registry

docker-push-operator:
	$(OCI_BIN) push ${TLS_SETTING} $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG)

docker-push-registry:
	$(OCI_BIN) push $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG)

prom-rules-verify:
	go run ./tools/prom-rule-ci $(OCI_BIN) ./tools/prom-rule-ci/tmp_prom_rules.yaml ./tools/prom-rule-ci/prom-rules-tests.yaml

cluster-up:
	./cluster/up.sh

cluster-down:
	./cluster/down.sh

cluster-sync: cluster-operator-push cluster-operator-install

cluster-operator-push:
	./cluster/operator-push.sh

cluster-operator-install:
	./cluster/operator-install.sh

$(E2E_SUITES): $(GO)
	GO=$(GO) TEST_SUITE=$@ TEST_ARGS="$(E2E_TEST_ARGS)" ./hack/functest.sh

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
	MULTUS_DYNAMIC_NETWORKS_CONTROLLER_IMAGE=$(MULTUS_DYNAMIC_NETWORKS_CONTROLLER_IMAGE) \
	KUBE_SECONDARY_DNS_IMAGE=$(KUBE_SECONDARY_DNS_IMAGE) \
	KUBEVIRT_IPAM_CONTROLLER_IMAGE=$(KUBEVIRT_IPAM_CONTROLLER_IMAGE) \
	CORE_DNS_IMAGE=$(CORE_DNS_IMAGE) \
	KUBE_RBAC_PROXY_IMAGE=$(KUBE_RBAC_PROXY_IMAGE) \
		./hack/generate-manifests.sh

gen-k8s: $(CONTROLLER_GEN) $(apis_sources)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
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

update-workflows-branches:
	./hack/update-workflows-branches.sh ${git_base_tag} ${branch_name}
statify-components:
	sed -i 's|update-policy: .*|update-policy: static|g' components.yaml

release-notes:
	hack/render-release-notes.sh $(WHAT)

release: $(GITHUB_RELEASE)
	GITHUB_RELEASE=$(GITHUB_RELEASE) \
	TAG=v$(shell hack/version.sh) \
	  hack/release.sh \
	    manifests/cluster-network-addons/cluster-network-addons.package.yaml \
	    $(shell find manifests/cluster-network-addons/$(shell hack/version.sh) -type f)

vendor: $(GO)
	$(GO) mod tidy -compat=$(GO_VERSION)
	$(GO) mod vendor

auto-bumper: $(GO)
	PUSH_IMAGES=true $(GO) run $(shell ls tools/bumper/*.go | grep -v test) ${ARGS}

bump-%:
	CNAO_VERSION=${VERSION} ./hack/components/bump-$*.sh
bump-all:
	set -e && for f in hack/components/bump*.*; do x=$${f%%.sh}; make $${x##*/}; done

generate-doc:
	go run ./tools/metricsdocs > docs/metrics.md

lint-metrics:
	./hack/prom_metric_linter.sh --operator-name="kubevirt" --sub-operator-name="cnao"

lint-monitoring:
	GOBIN=$$(pwd)/build/_output/bin/ $(GO) install -mod=mod github.com/kubevirt/monitoring/monitoringlinter/cmd/monitoringlinter@e2be790
	$(MONITORING_LINTER) ./...

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
	release \
	update-workflows-branches \
	statify-components \
	lint-monitoring

