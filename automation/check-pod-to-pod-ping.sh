#!/usr/bin/env bash
set -euo pipefail

# Cross-node pod-to-pod ICMP connectivity check (no virt-handler dependency).
#
# Creates two short-lived busybox pods pinned to two nodes and pings PodIP<->PodIP.
#
# Usage:
#   export KUBECONFIG=...
#   ./automation/check-pod-to-pod-ping.sh                # defaults: node02<->node03
#   NODE1=node01 NODE2=node02 ./automation/check-pod-to-pod-ping.sh
#   ./automation/check-pod-to-pod-ping.sh node01 node03  # args override env
#
# Notes:
# - ping needs CAP_NET_RAW. We run in a dedicated namespace labeled PodSecurity=privileged.
# - Requires cluster-admin (to create/label namespace).

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing required binary: $1" >&2; exit 1; }; }
need kubectl

if [[ -z "${KUBECONFIG:-}" ]]; then
  echo "KUBECONFIG is required" >&2
  exit 1
fi

NODE1="${NODE1:-node02}"
NODE2="${NODE2:-node03}"
if [[ $# -ge 1 ]]; then NODE1="$1"; fi
if [[ $# -ge 2 ]]; then NODE2="$2"; fi

NS="${NS:-pod-connectivity-check}"
IMG="${IMG:-busybox:1.36.1}"
PING_COUNT="${PING_COUNT:-2}"
PING_TIMEOUT_SEC="${PING_TIMEOUT_SEC:-1}"
POD_READY_TIMEOUT="${POD_READY_TIMEOUT:-300s}"
KEEP="${KEEP:-false}"

RUN_ID="$(date +%s)-$$"
POD1="p2p-ping-${RUN_ID}-n1"
POD2="p2p-ping-${RUN_ID}-n2"

cleanup() {
  if [[ "${KEEP}" == "true" ]]; then
    echo "KEEP=true: not cleaning up pods/namespace"
    return 0
  fi
  set +e
  kubectl -n "${NS}" delete pod "${POD1}" "${POD2}" --ignore-not-found >/dev/null 2>&1 || true
  #kubectl delete ns "${NS}" --ignore-not-found >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "## kubeconfig"
echo "KUBECONFIG=${KUBECONFIG}"
echo

echo "## target nodes"
echo "node1=${NODE1}"
echo "node2=${NODE2}"
echo

echo "## ensure privileged namespace for ping: ${NS}"
kubectl get ns "${NS}" >/dev/null 2>&1 || kubectl create ns "${NS}" >/dev/null
kubectl label ns "${NS}" \
  pod-security.kubernetes.io/enforce=privileged \
  pod-security.kubernetes.io/audit=privileged \
  pod-security.kubernetes.io/warn=privileged \
  --overwrite >/dev/null

create_pod() {
  local pod="$1" node="$2"
  cat <<EOF | kubectl -n "${NS}" apply -f - >/dev/null
apiVersion: v1
kind: Pod
metadata:
  name: ${pod}
  labels:
    app: p2p-ping
spec:
  nodeName: ${node}
  restartPolicy: Never
  containers:
  - name: busybox
    image: ${IMG}
    command: ["sh","-lc","sleep 3600"]
    imagePullPolicy: IfNotPresent
    securityContext:
      allowPrivilegeEscalation: false
      runAsUser: 0
      runAsNonRoot: false
      capabilities:
        drop: ["ALL"]
        add: ["NET_RAW"]
      seccompProfile:
        type: RuntimeDefault
EOF
}

echo "## create ping pods"
kubectl -n "${NS}" delete pod "${POD1}" "${POD2}" --ignore-not-found >/dev/null 2>&1 || true
create_pod "${POD1}" "${NODE1}"
create_pod "${POD2}" "${NODE2}"

echo "## wait ready (timeout=${POD_READY_TIMEOUT})"
kubectl -n "${NS}" wait --for=condition=Ready "pod/${POD1}" --timeout="${POD_READY_TIMEOUT}" >/dev/null
kubectl -n "${NS}" wait --for=condition=Ready "pod/${POD2}" --timeout="${POD_READY_TIMEOUT}" >/dev/null

echo "## pods"
kubectl -n "${NS}" get pod "${POD1}" "${POD2}" -o wide

IP1="$(kubectl -n "${NS}" get pod "${POD1}" -o jsonpath='{.status.podIP}')"
IP2="$(kubectl -n "${NS}" get pod "${POD2}" -o jsonpath='{.status.podIP}')"
echo
echo "## pod IPs"
echo "${POD1} (${NODE1}) = ${IP1}"
echo "${POD2} (${NODE2}) = ${IP2}"
echo

ping_failed=false

run_ping() {
  local src_pod="$1" src_node="$2" dst_pod="$3" dst_node="$4" dst_ip="$5"

  echo "## ping: ${src_node} -> ${dst_node}"
  set +e
  local out
  out="$(kubectl -n "${NS}" exec "${src_pod}" -- ping -c "${PING_COUNT}" -W "${PING_TIMEOUT_SEC}" "${dst_ip}" 2>&1)"
  local rc=$?
  set -e

  if [[ $rc -ne 0 ]]; then
    echo "ERROR: ping failed: ${src_pod} (${src_node}) -> ${dst_pod} (${dst_node}) [${dst_ip}]"
    echo "---- ping output ----"
    echo "${out}"
    echo "---------------------"
    ping_failed=true
    return 0
  fi

  echo "${out}"
  return 0
}

run_ping "${POD1}" "${NODE1}" "${POD2}" "${NODE2}" "${IP2}"
echo

echo "## result"
if [[ "${ping_failed}" == "true" ]]; then
  echo "FAILED: pod-to-pod ping failed (see errors above)"
  exit 1
fi
echo "OK: pod-to-pod ping succeeded both directions"
