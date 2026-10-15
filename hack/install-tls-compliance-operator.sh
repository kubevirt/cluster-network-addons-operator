#!/usr/bin/env bash

set -euo pipefail

VERSION=${VERSION:-v1.0.18}

readonly URL="https://github.com/sebrandon1/tls-compliance-operator/releases/download/${VERSION}/install.yaml"
readonly NAMESPACE="tls-compliance-operator-system"
readonly DEPLOY_NAME="tls-compliance-operator-controller-manager"
readonly NP_NAME="tls-compliance-operator-controller-manager"

echo "Installing TLS Compliance Operator ${VERSION} from: ${URL}.."
manifest=$(mktemp)
trap 'rm -f "$manifest"' EXIT
curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' --max-time 60 "$URL" -o "$manifest"
./cluster/kubectl.sh create --dry-run=client -f "$manifest" -o json --request-timeout=15s |
  jq -s --arg deployment "$DEPLOY_NAME" '
    {apiVersion: "v1", kind: "List", items: map(if .kind == "Deployment" and .metadata.name == $deployment then
      .spec.replicas = 0 |
      .spec.template.spec.containers[0].args += [
        "--include-namespaces=cluster-network-addons",
        "--scan-interval=5s", "--cleanup-interval=7s", "--max-retries=0"
      ]
    else . end)}' |
  ./cluster/kubectl.sh apply -f - --request-timeout=15s

echo "Checking the dedicated CI scanner's report scope before startup.."
./cluster/kubectl.sh -n "$NAMESPACE" rollout status deployment/"$DEPLOY_NAME" --timeout=120s --request-timeout=15s
./cluster/kubectl.sh -n "$NAMESPACE" wait --for=delete pods -l app.kubernetes.io/name=tls-compliance-operator,control-plane=controller-manager --timeout=120s --request-timeout=15s
./cluster/kubectl.sh wait --for=condition=Established crd/tlscompliancereports.security.telco.openshift.io --timeout=120s --request-timeout=15s
if ! ./cluster/kubectl.sh get tlscompliancereports -o json --request-timeout=15s |
  jq -e 'all(.items[]; .spec.sourceNamespace == "cluster-network-addons")' > /dev/null; then
  echo "TLS scanner requires an isolated CI cluster: existing reports outside cluster-network-addons are not namespace-filtered during scans; no reports were deleted." >&2
  exit 1
fi

# By default the operator scan common TLS ports (443, 6443, 8443, 9443), and installs
# permissive network-policy for those ports only [1].
# To allow the operator reach and check all services the network-policy is deleted.
# [1] https://github.com/sebrandon1/tls-compliance-operator/blob/v1.0.18/docs/troubleshooting.md#:~:text=Check%20for%20NetworkPolicy%20restrictions
echo "Deleting the operator's network-policy to allow egress all services.."
./cluster/kubectl.sh -n "$NAMESPACE" delete networkpolicy "$NP_NAME" --timeout=30s --request-timeout=15s

echo "Starting the scoped CI scanner (failed checks retry on the next scan cycle).."
./cluster/kubectl.sh -n "$NAMESPACE" scale deployment "$DEPLOY_NAME" --replicas=1 --request-timeout=15s

echo "Waiting for TLS Compliance Operator readiness.."
./cluster/kubectl.sh -n "$NAMESPACE" rollout status deployment/"$DEPLOY_NAME" --timeout=120s --request-timeout=15s

echo "TLS Compliance Operator is ready"
