#!/bin/bash -ex
#
# Copyright 2025 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# This script install NetworkPolicy that affects CNAO namespace.
# The network-policy blocks egress/ingress traffic in CNAO namespace with the following exceptions:
# 1. Allow egress to cluster API and DNS for pods who labeled with
#    "hco.kubevirt.io/allow-access-cluster-services"
# 2. Allow ingress to metrics endpoint to pods who labeled with
#    "hco.kubevirt.io/allow-prometheus-access"

readonly ns="$(./cluster/kubectl.sh get pod -l name=cluster-network-addons-operator -A -o=custom-columns=NS:.metadata.namespace --no-headers | head -1)"
[[ -z "${ns}" ]] && echo "FATAL: CNAO pods not found. Make sure its installed" && exit 1

cat <<EOF | ./cluster/kubectl.sh -n "${ns}" apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress: []
  egress: []
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-egress-to-cluster-dns
spec:
  podSelector:
    matchExpressions:
    - key: hco.kubevirt.io/allow-access-cluster-services
      operator: Exists
  policyTypes:
  - Egress
  egress:
    - to:
      - namespaceSelector:
          matchLabels:
            kubernetes.io/metadata.name: kube-system
        podSelector:
          matchLabels:
            k8s-app: kube-dns
      ports:
        - protocol: TCP
          port: 53
        - protocol: UDP
          port: 53
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-egress-to-cluster-api
spec:
  podSelector:
    matchExpressions:
    - key: hco.kubevirt.io/allow-access-cluster-services
      operator: Exists
  policyTypes:
  - Egress
  egress:
  - ports:
    - protocol: TCP
      port: 6443
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-ingress-to-metrics-endpoint
spec:
  podSelector:
    matchExpressions:
    - key: hco.kubevirt.io/allow-prometheus-access
      operator: Exists
  policyTypes:
  - Ingress
  ingress:
  - ports:
    - protocol: TCP
      port: 8443
EOF
