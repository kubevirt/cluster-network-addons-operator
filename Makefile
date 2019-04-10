all: fmt vet

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

manifests:
	CONTAINER_PREFIX=$(IMAGE_REGISTRY) CONTAINER_TAG=$(IMAGE_TAG) ./hack/build-manifests.sh

.PHONY:
	docker-build \
	docker-push \
	cluster-up \
	cluster-down \
	cluster-sync \
	cluster-clean \
	manifests
