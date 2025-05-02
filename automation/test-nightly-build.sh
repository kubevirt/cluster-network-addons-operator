#!/bin/bash -xe

# This script should be able to execute workflow functional tests against
# Kubernetes cluster on any environment with basic dependencies listed in
# check-patch.packages installed and docker running.

teardown() {
    # Don't fail if there is no logs
    cp ${E2E_LOGS}/workflow/*.log ${ARTIFACTS} || true
    export REPORT_FLAG="${ARTIFACTS}/junit.xml"

    make cluster-down
}

main() {
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    make cluster-down
    make cluster-up
    trap teardown EXIT SIGINT SIGTERM

    # create and push bundle image and index image
    build_date="$(date +%Y%m%d)"
    export IMAGE_REGISTRY=quay.io/kubevirtci
    export IMAGE_TAG="nb_${build_date}_$(git show -s --format=%h)"

    # build CNAO operator
    make cluster-operator-push
    make cluster-operator-install

    # run functest
    make E2E_TEST_EXTRA_ARGS="-ginkgo.noColor --ginkgo.junit-report=$ARTIFACTS/junit.xml" test/e2e/workflow

    # functional test passed: publish latest nightly build
    cnao_bucket="kubevirt-prow/devel/nightly/release/kubevirt/cluster-network-addons-operator"

    # publish the nightly build manifests
    CNAO_VERSION=$(hack/version.sh)
    SRC_DIR="./_out/cluster-network-addons/${CNAO_VERSION}"
    DEST="gs://${cnao_bucket}/${build_date}"
    gsutil cp "${SRC_DIR}/namespace.yaml" "${DEST}/namespace.yaml"
    gsutil cp "${SRC_DIR}/cluster-network-addons-operator.${CNAO_VERSION}.clusterserviceversion.yaml" "${DEST}/cluster-network-addons-operator.${CNAO_VERSION}.clusterserviceversion.yaml"
    gsutil cp "${SRC_DIR}/network-addons-config.crd.yaml" "${DEST}/network-addons-config.crd.yaml"
    gsutil cp "${SRC_DIR}/operator.yaml" "${DEST}/operator.yaml"
    gsutil cp "${SRC_DIR}/network-addons-config-example.cr.yaml" "${DEST}/network-addons-config-example.cr.yaml"

    echo "${build_date}" > build-date
    gsutil cp ./build-date gs://${cnao_bucket}/latest
}

main
