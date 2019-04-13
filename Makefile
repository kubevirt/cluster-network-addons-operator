all: fmt vet

# Always keep the future version here, so we won't overwrite latest released manifests
VERSION ?= 0.4.0

DEPLOY_DIR ?= deploy

IMAGE_REGISTRY ?= quay.io/kubevirt
IMAGE_TAG ?= latest
OPERATOR_IMAGE ?= cluster-network-addons-operator

vet:
	go vet ./pkg/... ./cmd/...

fmt:
	go fmt ./pkg/... ./cmd/...

docker-build:
	docker build -f build/Dockerfile -t $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG) .

docker-push:
	docker push $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG)

cluster-up:
	./cluster/up.sh

cluster-down:
	./cluster/down.sh

cluster-sync:
	./cluster/sync.sh

cluster-clean:
	./cluster/clean.sh

# Default images can be found in pkg/components/components.go
manifests:
	VERSION=$(VERSION) \
	DEPLOY_DIR=$(DEPLOY_DIR) \
	CONTAINER_PREFIX=$(IMAGE_REGISTRY) \
	CONTAINER_TAG=$(IMAGE_TAG) \
	MULTUS_IMAGE=$(MULTUS_IMAGE) \
	LINUX_BRIDGE_CNI_IMAGE=$(LINUX_BRIDGE_CNI_IMAGE) \
	SRIOV_DP_IMAGE=$(SRIOV_DP_IMAGE) \
	SRIOV_CNI_IMAGE=$(SRIOV_CNI_IMAGE) \
	KUBEMACPOOL_IMAGE=$(KUBEMACPOOL_IMAGE) \
		./hack/build-manifests.sh

.PHONY:
	docker-build \
	docker-push \
	cluster-up \
	cluster-down \
	cluster-sync \
	cluster-clean \
	manifests
