#!/bin/bash -xe

# This script should be able to execute KubeMacPool
# functional tests against Kubernetes cluster with
# CNAO built with latest changes, on any
# environment with basic dependencies listed in
# check-patch.packages installed and docker running.
#
# yum -y install automation/check-patch.packages
# automation/check-patch.e2e-kubemacpool-functests.sh

function __get_skipped_tests() {
	local  __resultvar=$1
	local skipped_regex=""

	# We can't test all KMP opt-mode CNAO context, as the operator will reconcile
	# back to the configured opt-mode when the test tries to change it.
	# So we check KMP webhook's opt-mode and skip tests accordingly
	if grep 'default/mutatevirtualmachines_opt_out_patch.yaml' hack/components/bump-kubemacpool.sh; then
		echo "KMP VM webhook is set to opt-out mode. Skipping opt-in Context"
		skipped_regex="${skipped_regex} \(opt-in\smode\)"
	elif grep 'default/mutatevirtualmachines_opt_in_patch.yaml' hack/components/bump-kubemacpool.sh; then
		echo "KMP VM webhook is set to opt-in mode. Skipping opt-out Context"
		skipped_regex="${skipped_regex} \(opt-out\smode\)"
	fi

	eval $__resultvar="'$skipped_regex'"
}

teardown() {
    rm -rf "${TMP_COMPONENT_PATH}"
    cd ${TMP_PROJECT_PATH}
    make cluster-down
}

main() {
    # Setup CNAO and artifacts temp directory
    source automation/check-patch.setup.sh
    cd ${TMP_PROJECT_PATH}

    # Spin-up ephemeral cluster with latest CNAO
    # this script also exports KUBECONFIG, and fetch $COMPONENT repository
    COMPONENT="kubemacpool" source automation/components-functests.setup.sh

    trap teardown EXIT

    echo "Deploy KubeVirt latest stable release"
    ./hack/deploy-kubevirt.sh

    echo "Get skip tests regex"
    __get_skipped_tests SKIPPED_TESTS

    # Run KubeMacPool functional tests
    cd ${TMP_COMPONENT_PATH}

    export CLUSTER_ROOT_DIRECTORY=${TMP_PROJECT_PATH}
    KUBECONFIG=${KUBECONFIG} E2E_TEST_ARGS="--junit-output=$ARTIFACTS/junit.functest.xml --ginkgo.skip $SKIPPED_TESTS" make functest
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
