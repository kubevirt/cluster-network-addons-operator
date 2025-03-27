#! /bin/bash
set -exu

# Prepare environment for CNAO testing and automation.
# This includes temporary Go paths and binaries.
#
# This script exports:
# - TMP_PROJECT_PATH
#   CNAO project temporary directory
#
# - ARTIFACTS
#   Tests suite artifacts directory
#
# Example:
# source automation/check-patch.setup.sh
# cd ${TMP_PROJECT_PATH}
# go test --junit-output=$ARTIFACTS/junit.functest.xml

tmp_dir=/tmp/cnao/

rm -rf $tmp_dir
mkdir -p $tmp_dir

export TMP_PROJECT_PATH=$tmp_dir/cluster-network-addons-operator
export E2E_LOGS=${TMP_PROJECT_PATH}/_out/e2e
export ARTIFACTS=${ARTIFACTS-$tmp_dir/artifacts}
mkdir -p $ARTIFACTS

rsync -rt --links --filter=':- .gitignore' $(pwd)/ $TMP_PROJECT_PATH
