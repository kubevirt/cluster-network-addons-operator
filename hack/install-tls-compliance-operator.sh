#!/usr/bin/env bash

set -euo pipefail

VERSION=${VERSION:-v1.0.18}

readonly URL="https://github.com/sebrandon1/tls-compliance-operator/releases/download/${VERSION}/install.yaml"
readonly NAMESPACE="tls-compliance-operator-system"
readonly DEPLOY_NAME="tls-compliance-operator-controller-manager"
readonly NP_NAME="tls-compliance-operator-controller-manager"

echo "Installing TLS Compliance Operator ${VERSION} from: ${URL}.."
./cluster/kubectl.sh create --dry-run=client -f "${URL}" -o json |
  jq -s --arg deployment "$DEPLOY_NAME" '
    {apiVersion: "v1", kind: "List", items: map(if .kind == "Deployment" and .metadata.name == $deployment then
      .spec.replicas = 0 |
      .spec.template.spec.containers[0].args += [
        "--include-namespaces=cluster-network-addons",
        "--scan-interval=5s", "--cleanup-interval=7s", "--max-retries=0"
      ]
    else . end)}' |
  ./cluster/kubectl.sh apply -f -

echo "Resetting reports for the dedicated CI scanner before startup.."
./cluster/kubectl.sh -n "$NAMESPACE" wait --for=delete pods --all --timeout=120s
./cluster/kubectl.sh wait --for=condition=Established crd/tlscompliancereports.security.telco.openshift.io --timeout=120s
./cluster/kubectl.sh delete tlscompliancereports --all

# By default the operator scan common TLS ports (443, 6443, 8443, 9443), and installs
# permissive network-policy for those ports only [1].
# To allow the operator reach and check all services the network-policy is deleted.
# [1] https://github.com/sebrandon1/tls-compliance-operator/blob/v1.0.18/docs/troubleshooting.md#:~:text=Check%20for%20NetworkPolicy%20restrictions
echo "Deleting the operator's network-policy to allow egress all services.."
./cluster/kubectl.sh -n $NAMESPACE delete networkpolicy $NP_NAME

echo "Starting the scoped CI scanner (failed checks retry on the next scan cycle).."
./cluster/kubectl.sh -n "$NAMESPACE" scale deployment "$DEPLOY_NAME" --replicas=1

echo "Waiting for TLS Compliance Operator readiness.."
./cluster/kubectl.sh -n $NAMESPACE rollout status deployment/$DEPLOY_NAME --timeout=120s

echo "TLS Compliance Operator is ready"
