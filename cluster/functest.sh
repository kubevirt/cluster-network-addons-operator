#!/bin/bash -e

function new_test() {
    name=$1

    printf "%0.s=" {1..80}
    echo
    echo ${name}
}

function assert_condition() {
    condition=$1
    timeout=$2

    echo "Wait until NetworkAddonsConfig reports ${condition} condition"
    if ./cluster/kubectl.sh wait networkaddonsconfig cluster --for condition=${condition} --timeout=${timeout}; then
        echo 'OK'
    else
        echo "Status has not reached ${condition} condition within the timeout. Actual state:"
        ./cluster/kubectl.sh get networkaddonsconfig cluster -o yaml
        echo 'FAILED'
        exit 1
    fi
}

new_test 'Test invalid configuration'
cat <<EOF | ./cluster/kubectl.sh apply -f -
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  imagePullPolicy: Always
  kubeMacPool:
    rangeStart: this:aint:right
  linuxBridge: {}
  multus: {}
  sriov: {}
EOF
assert_condition Failing 60s

new_test 'Test valid configuration'
cat <<EOF | ./cluster/kubectl.sh apply -f -
apiVersion: networkaddonsoperator.network.kubevirt.io/v1alpha1
kind: NetworkAddonsConfig
metadata:
  name: cluster
spec:
  imagePullPolicy: Always
  kubeMacPool: {}
  linuxBridge: {}
  multus: {}
  sriov: {}
EOF
assert_condition Progressing 60s
assert_condition Ready 300s
