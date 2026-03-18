#!/usr/bin/env bash

set -e

VERSION=${VERSION:-v0.0.10}

URL="https://github.com/sebrandon1/tls-compliance-operator/releases/download/${VERSION}/install.yaml"

NAMESPACE="tls-compliance-operator-system"
DEPLOY_NAME="tls-compliance-operator-controller-manager"
NP_NAME="tls-compliance-operator-controller-manager"

echo "Installing TLS Compliance Operator ${VERSION} from: ${URL}"
./cluster/kubectl.sh apply -f "${URL}"

# By default the operator scan common TLS ports (443, 6443, 8443, 9443), and installed with network-policy
# allow egress for common ports only [1].
# Remove the operator's network-policy to allow egress to all services including Kubemacpool's.
# [1] github.com/sebrandon1/tls-compliance-operator/blob/main/docs/troubleshooting.md#:~:text=Check for NetworkPolicy restrictions
echo "Delete the operator's network-policy to allow egress all services"
./cluster/kubectl.sh -n $NAMESPACE delete networkpolicy $NP_NAME 

echo "Patching deployment - fine tune reporting intervals"
./cluster/kubectl.sh -n $NAMESPACE patch deployment $DEPLOY_NAME --type='json' -p='[
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--scan-interval=30s"},
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--cleanup-interval=15s"},
]'

sleep 5 

echo "Waiting for TLS Compliance Operator readiness.."
./cluster/kubectl.sh -n $NAMESPACE rollout status deployment/$DEPLOY_NAME --timeout=120s

echo "TLS Compliance Operator is ready"
