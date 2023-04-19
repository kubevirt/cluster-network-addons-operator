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

SCRIPTS_PATH="$(dirname "$(realpath "$0")")"
source ${SCRIPTS_PATH}/cluster.sh

export KUBEVIRT_DEPLOY_PROMETHEUS=true
export KUBEVIRT_DEPLOY_PROMETHEUS_ALERTMANAGER=true
export KUBEVIRT_DEPLOY_GRAFANA=true

cluster::install

$(cluster::path)/cluster-up/up.sh

if [[ "$KUBEVIRT_PROVIDER" =~ k8s- ]]; then
    echo 'Enable Open vSwitch'
    for node in $(./cluster/kubectl.sh get nodes --no-headers -o=custom-columns=NAME:.metadata.name); do
        ./cluster/cli.sh ssh ${node} -- sudo systemctl enable --now openvswitch
    done

    ./cluster/ssh.sh node01 "sudo curl -LO https://github.com/liggitt/audit2rbac/releases/download/v0.10.0/audit2rbac-linux-386.tar.gz"
    ./cluster/ssh.sh node01 "sudo tar -xvf audit2rbac-linux-386.tar.gz"
    ./cluster/ssh.sh node01 "sudo chmod 777 /etc/kubernetes/audit/adv-audit.yaml"
    ./cluster/ssh.sh node01 "sudo chmod 777 /etc/kubernetes/manifests/kube-apiserver.yaml"
    ./cluster/ssh.sh node01 "sudo chmod 777 /tmp"
    ./cluster/ssh.sh node01 "echo YXBpVmVyc2lvbjogYXVkaXQuazhzLmlvL3YxCmtpbmQ6IFBvbGljeQpydWxlczoKLSBsZXZlbDogTWV0YWRhdGEK | base64 -d  > /etc/kubernetes/audit/adv-audit.yaml"
    ./cluster/ssh.sh node01 "sudo mv /etc/kubernetes/manifests/kube-apiserver.yaml /tmp"
    sleep 3
    ./cluster/ssh.sh node01 "sudo mv /tmp/kube-apiserver.yaml /etc/kubernetes/manifests"
    until ./cluster/kubectl.sh get pods -A; do sleep 1; done
fi
