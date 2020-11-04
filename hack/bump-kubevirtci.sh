#!/bin/bash
val=$(git ls-remote https://github.com/kubevirt/kubevirtci | grep HEAD | awk '{print $1}')
sed -i "/^[[:blank:]]*[KUBEVIRTCI_VERSION[:blank:]]*=/s/=.*/=\"${val}\"/" cluster/cluster.sh
git --no-pager diff cluster/cluster.sh | grep KUBEVIRTCI_VERSION
