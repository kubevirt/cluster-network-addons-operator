#!/usr/bin/env bash

set -xeo pipefail

echo 'Setup temporary Go path'
export GOPATH=${PWD}/_components/go
mkdir -p $GOPATH
export PATH=${GOPATH}/bin:${PATH}

echo 'kubemacpool'
CNAO_VERSION=${VERSION} ./hack/components/bump-kubemacpool.sh

echo 'macvtap'
CNAO_VERSION=${VERSION} ./hack/components/bump-macvtap.sh

echo 'kubernetes-nmstate'
CNAO_VERSION=${VERSION} ./hack/components/bump-knmstate.sh
