#!/bin/bash

# This script pulls release manifests needed for lifecycle tests.
#
# Usage example:
# export RELEASES_SELECTOR="{0.89.3,0.91.1,0.93.0,0.95.0,99.0.0}"
# ./hack/update_releases.sh

set -ue
RELEASES=${RELEASES_SELECTOR//[\{\}]}

IFS=',' read -ra RELEASES_ARRAY <<< "$RELEASES"

for ((i = 0; i < ${#RELEASES_ARRAY[@]} - 1; i++)); do
    VERSION=${RELEASES_ARRAY[i]}
    if [ -d "manifests/cluster-network-addons/${VERSION}" ]; then
      continue
    fi
    echo "Processing release: ${VERSION}"

    mkdir -p manifests/cluster-network-addons/${VERSION}
    pushd manifests/cluster-network-addons/${VERSION} > /dev/null
    curl -sLO https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v${VERSION}/cluster-network-addons-operator.${VERSION}.clusterserviceversion.yaml
    curl -sLO https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v${VERSION}/namespace.yaml
    curl -sLO https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v${VERSION}/network-addons-config-example.cr.yaml
    curl -sLO https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v${VERSION}/network-addons-config.crd.yaml
    curl -sLO https://github.com/kubevirt/cluster-network-addons-operator/releases/download/v${VERSION}/operator.yaml
    popd > /dev/null

    pushd test/releases > /dev/null
    curl -sLO https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator/release-${VERSION%.*}/test/releases/${VERSION}.go
    popd > /dev/null
done

