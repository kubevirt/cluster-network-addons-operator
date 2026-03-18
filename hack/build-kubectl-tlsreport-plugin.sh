#!/usr/bin/env bash
#
# Builds a container image containing the kubectl-tlsreport plugin using podman.
# See https://github.com/sebrandon1/tls-compliance-operator/blob/main/docs/exporting-reports.md
#
# Usage:
#   ./hack/build-kubectl-tlsreport-plugin.sh
#
# Examples:
#   ./hack/build-kubectl-tlsreport-plugin.sh
#   IMAGE_NAME=quay.io/myuser/kubectl-tlsreport:latest ./hack/build-kubectl-tlsreport-plugin.sh
#   VERSION=v1.0.0 ./hack/build-kubectl-tlsreport-plugin.sh
#
# How to use:
#   podman run --rm -v $KUBECONFIG:/root/c -e KUBECONFIG=/root/c kubectl-tlsreport:latest \
#      kubectl tlsreport summary

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

OCI_BIN="${OCI_BIN:-podman}"

REPO_URL="${REPO_URL:-https://github.com/sebrandon1/tls-compliance-operator.git}"
VERSION="${VERSION:-v0.0.10}"

IMAGE_NAME="${IMAGE_NAME:-kubectl-tlsreport:latest}"

WORK_DIR="${TMPDIR:-/tmp}/kubectl-tlsreport-build-$$"

cleanup() {
  rm -rf "${WORK_DIR}"
}
trap cleanup EXIT

echo "Building kubectl-tlsreport image from ${REPO_URL} (${VERSION})"
mkdir -p "${WORK_DIR}"
cd "${WORK_DIR}"

cat > Dockerfile <<EOF
# Build stage: compile kubectl-tlsreport
FROM golang:1.26 AS builder
RUN apt-get update && apt-get install -y --no-install-recommends \
  git \
  ca-certificates \
  curl \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /build

RUN git clone --depth 1 --branch ${VERSION} ${REPO_URL} . || \
    (git clone --depth 1 ${REPO_URL} . && git checkout ${VERSION})
RUN go build -o kubectl-tlsreport ./cmd/kubectl-tlsreport/

RUN curl -sSL -o kubectl \
  "https://dl.k8s.io/release/\$(curl -sL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && chmod +x kubectl

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /build/kubectl-tlsreport /usr/local/bin/kubectl-tlsreport
COPY --from=builder /build/kubectl           /usr/local/bin/kubectl
CMD ["/bin/sh"]
EOF

echo "Building image: ${IMAGE_NAME}"
podman build -t "${IMAGE_NAME}" .

echo "Built: ${IMAGE_NAME}"
echo "How to use:"
echo "Run:"
echo "  podman run --rm -v $KUBECONFIG:/root/c -e KUBECONFIG=/root/c ${IMAGE_NAME} kubectl tlsreport summary"
echo "  podman run --rm -v $KUBECONFIG:/root/c -e KUBECONFIG=/root/c ${IMAGE_NAME} kubectl tlsreport junit"
