#!/usr/bin/env bash

set -xeo pipefail

CNAO_VERSION=${VERSION} # Exported from Makefile

echo 'Setup temporary Go path'
export GOPATH=${PWD}/_components/go
mkdir -p $GOPATH
export PATH=${GOPATH}/bin:${PATH}

echo 'kubemacpool'
source hack/components/bump-kubemacpool.sh

echo 'macvtap'
source hack/components/bump-macvtap.sh

echo 'kubernetes-nmstate'
source hack/components/bump-knmstate.sh
