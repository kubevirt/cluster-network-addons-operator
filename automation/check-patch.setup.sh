# Prepare environment for CNAO testing and automation. This includes temporary Go paths and binaries.
#
# source automation/check-patch.setup.sh
# cd ${TMP_PROJECT_PATH}

tmp_dir=/tmp/cnao/

rm -rf $tmp_dir
mkdir -p $tmp_dir

export TMP_PROJECT_PATH=$tmp_dir/cluster-network-addons-operator
export ARTIFACTS=${ARTIFACTS-$tmp_dir/artifacts}
mkdir -p $ARTIFACTS

rsync -rt --links --filter=':- .gitignore' $(pwd)/ $TMP_PROJECT_PATH

