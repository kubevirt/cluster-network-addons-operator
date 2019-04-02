all: fmt vet

MANIFESTS_OUTPUT_DIR ?= deploy/manifests

OPERATOR_IMAGE_REGISTRY ?= quay.io/kubevirt
OPERATOR_IMAGE_TAG ?= latest
OPERATOR_IMAGE ?= cluster-network-addons-operator
IMAGE_PULL_POLICY ?= Always

MULTUS_IMAGE ?= docker.io/nfvpe/multus:latest
LINUX_BRIDGE_IMAGE ?= quay.io/kubevirt/cni-default-plugins
SRIOV_DP_IMAGE ?= quay.io/booxter/sriov-device-plugin:latest
SRIOV_CNI_IMAGE ?= docker.io/nfvpe/sriov-cni:latest
KUBEMACPOOL_IMAGE ?= quay.io/schseba/mac-controller:latest

vet:
	go vet ./pkg/... ./cmd/...

fmt:
	go fmt ./pkg/... ./cmd/...

generate-manifests:
	mkdir -p $(MANIFESTS_OUTPUT_DIR)
	for template in deploy/templates/*.in; do \
	    name=$$(basename $${template%.in}); \
	    echo $$name; \
	    sed \
	      -e "s#\$$OPERATOR_IMAGE_REGISTRY#$(OPERATOR_IMAGE_REGISTRY)#g" \
	      -e "s#\$$OPERATOR_IMAGE_TAG#$(OPERATOR_IMAGE_TAG)#g" \
	      -e "s#\$$OPERATOR_IMAGE#$(OPERATOR_IMAGE)#g" \
	      -e "s#\$$IMAGE_PULL_POLICY#$(IMAGE_PULL_POLICY)#g" \
	      -e "s#\$$MULTUS_IMAGE#$(MULTUS_IMAGE)#g" \
	      -e "s#\$$LINUX_BRIDGE_IMAGE#$(LINUX_BRIDGE_IMAGE)#g" \
	      -e "s#\$$SRIOV_DP_IMAGE#$(SRIOV_DP_IMAGE)#g" \
	      -e "s#\$$SRIOV_CNI_IMAGE#$(SRIOV_CNI_IMAGE)#g" \
	      -e "s#\$$KUBEMACPOOL_IMAGE#$(KUBEMACPOOL_IMAGE)#g" \
	    $${template} > $(MANIFESTS_OUTPUT_DIR)/$${name}; \
	done

docker-build:
	docker build -f build/Dockerfile -t $(OPERATOR_IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(OPERATOR_IMAGE_TAG) .

docker-push:
	docker push $(OPERATOR_IMAGE_REGISTRY)/$(OPERATOR_IMAGE):$(OPERATOR_IMAGE_TAG)

cluster-up:
	./cluster/up.sh

cluster-down:
	./cluster/down.sh

cluster-sync:
	./cluster/sync.sh

cluster-clean:
	./cluster/clean.sh

.PHONY:
	generate-manifests \
	docker-build \
	docker-push \
	cluster-up \
	cluster-down \
	cluster-sync \
	cluster-clean
