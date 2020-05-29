# Prepare environment for CNAO testing and automation. This includes temporary Go paths and binaries.
#
# source automation/check-patch.setup.sh
# cd ${TMP_PROJECT_PATH}

echo 'Setup Go paths'
export GOROOT=/tmp/cluster-network-addons-operator/go/root
mkdir -p $GOROOT
export GOPATH=/tmp/cluster-network-addons-operator/go/path
mkdir -p $GOPATH
export PATH=${GOPATH}/bin:${GOROOT}/bin:${PATH}

echo 'Install Go 1.13'
export GIMME_GO_VERSION=1.13
GIMME=/tmp/cluster-network-addons-operator/go/gimme
mkdir -p $GIMME
curl -sL https://raw.githubusercontent.com/travis-ci/gimme/master/gimme | HOME=${GIMME} bash >> ${GIMME}/gimme.sh
source ${GIMME}/gimme.sh

echo 'Install operator repository under the temporary Go path'
TMP_PROJECT_PATH=${GOPATH}/src/github.com/kubevirt/cluster-network-addons-operator
rm -rf ${TMP_PROJECT_PATH}
mkdir -p ${TMP_PROJECT_PATH}
cp -rf $(pwd)/. ${TMP_PROJECT_PATH}

echo 'Exporting temporary project path'
export TMP_PROJECT_PATH
