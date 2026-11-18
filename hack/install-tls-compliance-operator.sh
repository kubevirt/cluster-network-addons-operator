#!/usr/bin/env bash

set -eo pipefail

VERSION=${VERSION:-v1.2.0}

readonly URL="https://github.com/sebrandon1/tls-compliance-operator/releases/download/${VERSION}/install.yaml"
readonly NAMESPACE="tls-compliance-operator-system"
readonly DEPLOY_NAME="tls-compliance-operator-controller-manager"
readonly NP_NAME="tls-compliance-operator-controller-manager"

echo "Installing TLS Compliance Operator ${VERSION} from: ${URL}.."
./cluster/kubectl.sh apply -f "${URL}"

# By default the operator scan common TLS ports (443, 6443, 8443, 9443), and installs
# permissive network-policy for those ports only [1].
# To allow the operator reach and check all services the network-policy is deleted.
# [1] https://github.com/sebrandon1/tls-compliance-operator/blob/v1.2.0/docs/troubleshooting.md#:~:text=Check%20for%20NetworkPolicy%20restrictions
echo "Deleting the operator's network-policy to allow egress all services.."
./cluster/kubectl.sh -n $NAMESPACE delete networkpolicy $NP_NAME

# The e2e only checks CNAO reports, so limit scans to the CNAO namespace.
# Since v1.1.21 the namespace filter also applies to reports the first pod
# already created, and cleanup deletes them, so a plain Deployment patch is
# enough and the manifest no longer has to be rewritten with yq before the
# first start.
#
# The periodic scan probes each endpoint once per pass instead of sleeping in
# retry backoff, so --max-retries=0 is no longer needed.
echo "Patching the operator's Deployment for fine tuning reporting intervals and namespace scope.."
./cluster/kubectl.sh -n $NAMESPACE patch deployment $DEPLOY_NAME --type='json' -p='[
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--scan-interval=5s"},
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--cleanup-interval=7s"},
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--include-namespaces=cluster-network-addons"}}
]'

echo "Waiting for TLS Compliance Operator readiness.."
./cluster/kubectl.sh -n $NAMESPACE rollout status deployment/$DEPLOY_NAME --timeout=120s

echo "TLS Compliance Operator is ready"
