#!/usr/bin/env bash

set -e

VERSION=${VERSION:-v0.0.10}

URL="https://github.com/sebrandon1/tls-compliance-operator/releases/download/${VERSION}/install.yaml"

echo "Installing TLS Compliance Operator ${VERSION} from: ${URL}"
./cluster/kubectl.sh apply -f "${URL}"

echo "Waiting for TLS Compliance Operator readiness.."
./cluster/kubectl.sh rollout status deployment/tls-compliance-operator-controller-manager -n tls-compliance-operator-system --timeout=120s

echo "TLS Compliance Operator is ready"
