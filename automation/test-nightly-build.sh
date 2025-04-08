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

    # build succeeded: publish the nightly build manifests
    cnao_bucket="kubevirt-prow/devel/nightly/release/kubevirt/cluster-network-addons-operator"
    gsutil cp ./_out/cluster-network-addons/99.0.0/namespace.yaml "gs://${cnao_bucket}/${build_date}"
    gsutil cp ./_out/cluster-network-addons/99.0.0/cluster-network-addons-operator.99.0.0.clusterserviceversion.yaml "gs://${cnao_bucket}/${build_date}"
    gsutil cp ./_out/cluster-network-addons/99.0.0/network-addons-config.crd.yaml "gs://${cnao_bucket}/${build_date}"
    gsutil cp ./_out/cluster-network-addons/99.0.0/operator.yaml "gs://${cnao_bucket}/${build_date}"
    gsutil cp ./_out/cluster-network-addons/99.0.0/network-addons-config-example.cr.yaml "gs://${cnao_bucket}/${build_date}"

    # run functest
    make E2E_TEST_EXTRA_ARGS="-ginkgo.noColor --ginkgo.junit-report=$ARTIFACTS/junit.xml" test/e2e/workflow

    # functional test passed: publish latest nightly build
    echo "${build_date}" > build-date
    gsutil cp ./build-date gs://${cnao_bucket}/latest
}

main
