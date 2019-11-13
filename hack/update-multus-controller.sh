#!/usr/bin/env bash
#
# This script builds and push k8s-net-attach-def-controller image to IMAGE_PATH
#
# Use it environment variable:
# IMAGE_PATH=quay.io/hlinacz/k8s-net-attach-def-controller ./hack/multus-controller-update.sh

set -e

mkdir temp
cd temp

git clone https://github.com/k8snetworkplumbingwg/k8s-net-attach-def-controller.git .
make image

docker tag k8s-net-attach-def-controller ${IMAGE_PATH}
docker push ${IMAGE_PATH}

cd ..
rm temp/ -rf