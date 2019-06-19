#!/bin/bash -xe

main() {
    CLUSTER_PROVIDER="$0"
    CLUSTER_PROVIDER="${CLUSTER_PROVIDER#./}"
    CLUSTER_PROVIDER="${CLUSTER_PROVIDER#*.*.}"
    CLUSTER_PROVIDER="${CLUSTER_PROVIDER%.*}"
    echo "CLUSTER_PROVIDER=$CLUSTER_PROVIDER"
    export CLUSTER_PROVIDER

    TEST_SUITE="$0"
    TEST_SUITE="${TEST_SUITE#./}"
    TEST_SUITE="${TEST_SUITE#*.}"
    TEST_SUITE="${TEST_SUITE%%.*}"
    echo "TEST_SUITE=$TEST_SUITE"
    export TEST_SUITE

    echo "Setup Go paths"
    cd ..
    export GOROOT=/usr/local/go
    export GOPATH=$(pwd)/go
    export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
    mkdir -p $GOPATH

    echo "Install Go 1.11"
    export GIMME_GO_VERSION=1.11
    mkdir -p /gimme
    curl -sL https://raw.githubusercontent.com/travis-ci/gimme/master/gimme | HOME=/gimme bash >> /etc/profile.d/gimme.sh
    source /etc/profile.d/gimme.sh

    echo "Install operator repository to the right place"
    mkdir -p $GOPATH/src/github.com/kubevirt
    mkdir -p $GOPATH/pkg
    ln -s $(pwd)/cluster-network-addons-operator $GOPATH/src/github.com/kubevirt/
    cd $GOPATH/src/github.com/kubevirt/cluster-network-addons-operator

    echo "Run functional tests"
    exec automation/test.sh
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
