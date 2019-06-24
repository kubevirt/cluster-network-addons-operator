#!/bin/bash -ex

kubectl() { cluster/kubectl.sh "$@"; }

# Make sure that the VM is properly shut down on exit
trap '{ make cluster-down; }' EXIT SIGINT SIGTERM SIGSTOP

export E2E_TEST_EXTRA_ARGS="-ginkgo.noColor"

make cluster-down
make cluster-up
if [ "${TEST_SUITE}" == "workflow" ]; then
    make cluster-operator-push
    make cluster-operator-install
    make test/e2e/workflow
elif [ "${TEST_SUITE}" == "lifecycle" ]; then
    make cluster-operator-push
    make test/e2e/lifecycle
else
    echo "Unknown test suite ${TEST_SUITE}"
fi
