all: fmt vet

# Always keep the future version here, so we won't overwrite latest released manifests
VERSION ?= 0.4.0

DEPLOY_DIR ?= manifests

IMAGE_REGISTRY ?= quay.io/kubevirt
IMAGE_TAG ?= latest
OPERATOR_IMAGE ?= cluster-network-addons-operator
REGISTRY_IMAGE ?= cluster-network-addons-registry

vet:
	go vet ./pkg/... ./cmd/...

fmt:
	go fmt ./pkg/... ./cmd/...

docker-build-operator:
	docker build -f build/operator/Dockerfile -t $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG) .

docker-build-registry:
	docker build -f build/registry/Dockerfile -t $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG) .

docker-push-operator:
	docker push $(IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(IMAGE_TAG)

docker-push-registry:
	docker push $(IMAGE_REGISTRY)/$(REGISTRY_IMAGE):$(IMAGE_TAG)

cluster-up:
	./cluster/up.sh

cluster-down:
	./cluster/down.sh

cluster-sync:
	VERSION=$(VERSION) ./cluster/sync.sh

cluster-clean:
	VERSION=$(VERSION) ./cluster/clean.sh

# Default images can be found in pkg/components/components.go
generate-manifests:
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
	docker-build-operator \
	docker-push-operator \
	docker-build-registry \
	docker-push-registry \
	cluster-up \
	cluster-down \
	cluster-sync \
	cluster-clean \
	generate-manifests
