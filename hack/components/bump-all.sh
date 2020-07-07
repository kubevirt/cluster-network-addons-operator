#!/usr/bin/env bash

set -xeo pipefail

source hack/components/common_functions.sh

CNAO_VERSION=${VERSION} # Exported from Makefile

echo 'Setup temporary Go path'
export GOPATH=${PWD}/_components/go
mkdir -p $GOPATH
export PATH=${GOPATH}/bin:${PATH}

echo 'kubemacpool'
source hack/components/bump_kubemacpool.sh

echo 'macvtap'
source hack/components/bump_macvtap.sh

echo 'kubernetes-nmstate'
source hack/components/bump_knmstate.sh
