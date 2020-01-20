all: fmt check

# Always keep the future version here, so we won't overwrite latest released manifests
VERSION ?= 0.26.0
export VERSION := $(VERSION)
# Always keep the last released version here
VERSION_REPLACES ?= 0.25.0

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

GINKGO_EXTRA_ARGS ?=
GINKGO_ARGS ?= --v -r --progress $(GINKGO_EXTRA_ARGS)
GINKGO ?= build/_output/bin/ginkgo

E2E_TEST_EXTRA_ARGS ?=
E2E_TEST_ARGS ?= $(strip -test.v -test.timeout 3h -ginkgo.v $(E2E_TEST_EXTRA_ARGS))
E2E_SUITES = \
	test/e2e/lifecycle \
	test/e2e/workflow

OPERATOR_SDK ?= build/_output/bin/operator-sdk

$(GINKGO): Gopkg.toml
	GOBIN=$$(pwd)/build/_output/bin/ go install ./vendor/github.com/onsi/ginkgo/ginkgo

$(OPERATOR_SDK): Gopkg.toml
	GOBIN=$$(pwd)/build/_output/bin/ go install ./vendor/github.com/operator-framework/operator-sdk/cmd/operator-sdk

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
	go run ./vendor/golang.org/x/tools/cmd/goimports -w ./pkg ./cmd ./test/
	touch $@

whitespace: $(all_sources)
	./hack/whitespace.sh --fix
	touch $@

check: whitespace-check vet goimports-check gen-k8s-check test/unit

whitespace-check: $(all_sources)
	./hack/whitespace.sh
	touch $@

vet: $(cmd_sources) $(pkg_sources)
	go vet ./pkg/... ./cmd/... ./test/...
	touch $@

goimports-check: $(cmd_sources) $(pkg_sources)
	go run ./vendor/golang.org/x/tools/cmd/goimports -d ./pkg ./cmd
	touch $@

test/unit: $(GINKGO)
	$(GINKGO) $(GINKGO_ARGS) ./pkg/ ./cmd/

docker-build: docker-build-operator docker-build-registry

docker-build-operator:
	docker build -f build/operator/Dockerfile -t $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG) .

docker-build-registry:
	docker build -f build/registry/Dockerfile -t $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG) .

docker-push: docker-push-operator docker-push-registry

docker-push-operator:
	docker push $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG)

docker-push-registry:
	docker push $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG)

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
	OPERATOR_SDK=$(OPERATOR_SDK) TEST_SUITE=$@ TEST_ARGS="$(E2E_TEST_ARGS)" ./hack/functest.sh

cluster-clean:
	./cluster/clean.sh

# Default images can be found in pkg/components/components.go
gen-manifests:
	VERSION_REPLACES=$(VERSION_REPLACES) \
	DEPLOY_DIR=$(DEPLOY_DIR) \
	CONTAINER_PREFIX=$(IMAGE_REGISTRY) \
	CONTAINER_TAG=$(IMAGE_TAG) \
	MULTUS_IMAGE=$(MULTUS_IMAGE) \
	LINUX_BRIDGE_CNI_IMAGE=$(LINUX_BRIDGE_CNI_IMAGE) \
	KUBEMACPOOL_IMAGE=$(KUBEMACPOOL_IMAGE) \
		./hack/generate-manifests.sh

gen-k8s: $(OPERATOR_SDK) $(apis_sources)
	$(OPERATOR_SDK) generate k8s
	touch $@

gen-k8s-check: $(apis_sources)
	./hack/verify-codegen.sh
	touch $@

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
	docker-build \
	docker-build-operator \
	docker-build-registry \
	docker-push \
	docker-push-operator \
	docker-push-registry \
	gen-manifests \
	test/unit
