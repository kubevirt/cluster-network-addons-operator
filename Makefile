all: fmt vet

vet:
	go vet ./pkg/... ./cmd/...

fmt:
	go fmt ./pkg/... ./cmd/...

cluster-up:
	./cluster/up.sh

cluster-down:
	./cluster/down.sh

cluster-sync:
	./cluster/sync.sh

cluster-clean:
	./cluster/clean.sh

.PHONY: cluster-up \
	cluster-down \
	cluster-sync \
	cluster-clean
