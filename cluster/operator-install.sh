#!/bin/bash
#
# Copyright 2018-2019 Red Hat, Inc.
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

set -ex

if ./cluster/kubectl.sh get crd prometheuses.monitoring.coreos.com; then
  # TODO: temporary hack for testing PR
   ./cluster/kubectl.sh  -n cluster-network-addons create rolebinding  cluster-network-addons-operator-monitoring-hack --role=cluster-network-addons-operator-monitoring --serviceaccount=monitoring:prometheus-k8s
  ./cluster/kubectl.sh patch prometheus k8s -n monitoring --type=json -p '[{"op": "replace", "path": "/spec/ruleSelector", "value":{}}, {"op": "replace", "path": "/spec/ruleNamespaceSelector", "value":{"matchLabels": {"prometheus.cnao.io": "true"}}}]'
  ./cluster/kubectl.sh patch prometheus k8s -n monitoring --type=json -p '[{"op": "replace", "path": "/spec/ruleSelector", "value":{}}, {"op": "replace", "path": "/spec/serviceMonitorNamespaceSelector", "value":{"matchLabels": {"prometheus.cnao.io": "true"}}}]'
fi

./cluster/kubectl.sh create -f _out/cluster-network-addons/${VERSION}/operator.yaml
./cluster/kubectl.sh -n cluster-network-addons wait deployment cluster-network-addons-operator --for condition=Available --timeout=600s
