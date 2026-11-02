#!/usr/bin/env bash

set -eo pipefail

VERSION=${VERSION:-v1.0.18}

readonly URL="https://github.com/sebrandon1/tls-compliance-operator/releases/download/${VERSION}/install.yaml"
readonly NAMESPACE="tls-compliance-operator-system"
export DEPLOY_NAME="tls-compliance-operator-controller-manager"
readonly NP_NAME="tls-compliance-operator-controller-manager"
export NS_ARG="--include-namespaces=cluster-network-addons"

# Manager container args in the upstream install manifest.
readonly MANAGER_ARGS='
  select(.kind == "Deployment" and .metadata.name == env(DEPLOY_NAME))
  | .spec.template.spec.containers[]
  | select(.name == "manager")
  | .args
'

# The namespace limit must be in the manifest before the first start. A later
# Deployment patch does not stop scanAllEndpoints from rechecking reports the
# unscoped process already created.
echo "Installing TLS Compliance Operator ${VERSION} from: ${URL}.."
manifest="$(mktemp)"
trap 'rm -f "$manifest"' EXIT

curl -fsSL "${URL}" | yq "(${MANAGER_ARGS}) += [strenv(NS_ARG)]" > "${manifest}"

if ! yq -e "${MANAGER_ARGS}[] | select(. == strenv(NS_ARG))" "${manifest}" >/dev/null; then
    echo "failed to set ${NS_ARG} on the tls-compliance-operator manager container" >&2
    exit 1
fi

./cluster/kubectl.sh apply -f "${manifest}"

# By default the operator scan common TLS ports (443, 6443, 8443, 9443), and installs
# permissive network-policy for those ports only [1].
# To allow the operator reach and check all services the network-policy is deleted.
# [1] https://github.com/sebrandon1/tls-compliance-operator/blob/v1.0.18/docs/troubleshooting.md#:~:text=Check%20for%20NetworkPolicy%20restrictions
echo "Deleting the operator's network-policy to allow egress all services.."
./cluster/kubectl.sh -n $NAMESPACE delete networkpolicy $NP_NAME

echo "Patching the operator's Deployment for fine tuning reporting intervals.."
./cluster/kubectl.sh -n $NAMESPACE patch deployment $DEPLOY_NAME --type='json' -p='[
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--scan-interval=5s"},
  {"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--cleanup-interval=7s"}
]'

echo "Waiting for TLS Compliance Operator readiness.."
./cluster/kubectl.sh -n $NAMESPACE rollout status deployment/$DEPLOY_NAME --timeout=120s

echo "TLS Compliance Operator is ready"
