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
    source hack/components/docker-utils.sh
    export OCI_BIN=${OCI_BIN:-$(docker-utils::determine_cri_bin)}
    cat "${QUAY_PASSWORD}" | ${OCI_BIN} login --username $(cat "${QUAY_USER}") --password-stdin=true quay.io

    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    make cluster-down
    make cluster-up
    trap teardown EXIT SIGINT SIGTERM

    # create and push bundle image and index image
    build_date="$(date +%Y%m%d)"
    export IMAGE_REGISTRY=quay.io/kubevirt
    export IMAGE_TAG="${build_date}_$(git show -s --format=%h)"
    export VERSION=${IMAGE_TAG}
    export DEPLOY_DIR=_out

     # build and push CNAO operator image
     make docker-build-operator docker-push-operator

    # generate manifests
    rm -rf ${DEPLOY_DIR}
    make gen-manifests

    # Create NS and deploy the CRD
    ./cluster/kubectl.sh create -f _out/cluster-network-addons/${VERSION}/namespace.yaml
    ./cluster/kubectl.sh create -f _out/cluster-network-addons/${VERSION}/network-addons-config.crd.yaml

    # Deploy the operator
    make cluster-operator-install

    # Simulate network restrictions on CNAO namespace
    ./hack/install-network-policy.sh

    # run functest
    make E2E_TEST_EXTRA_ARGS="-ginkgo.noColor --ginkgo.junit-report=$ARTIFACTS/junit.xml" test/e2e/workflow

    # functional test passed: publish latest nightly build
    cnao_bucket="kubevirt-prow/devel/nightly/release/kubevirt/cluster-network-addons-operator"

    # publish the nightly build manifests
    SRC_DIR="./_out/cluster-network-addons/${VERSION}"
    DEST="gs://${cnao_bucket}/${build_date}"
    gsutil cp "${SRC_DIR}/namespace.yaml" "${DEST}/namespace.yaml"
    gsutil cp "${SRC_DIR}/cluster-network-addons-operator.${VERSION}.clusterserviceversion.yaml" "${DEST}/cluster-network-addons-operator.${VERSION}.clusterserviceversion.yaml"
    gsutil cp "${SRC_DIR}/network-addons-config.crd.yaml" "${DEST}/network-addons-config.crd.yaml"
    gsutil cp "${SRC_DIR}/operator.yaml" "${DEST}/operator.yaml"
    gsutil cp "${SRC_DIR}/network-addons-config-example.cr.yaml" "${DEST}/network-addons-config-example.cr.yaml"

    git show -s --format=%H > ${SRC_DIR}/commit
    gsutil cp ${SRC_DIR}/commit "${DEST}/commit"

    echo "${build_date}" > "${SRC_DIR}/build-date"
    gsutil cp "${SRC_DIR}/build-date" gs://${cnao_bucket}/latest
}

main
