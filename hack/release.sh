#!/bin/bash -e

make IMAGE_TAG=$TAG docker-build docker-push

git tag $TAG
git push https://github.com/kubevirt/cluster-network-addons-operator $TAG

$GITHUB_RELEASE release -u kubevirt -r cluster-network-addons-operator \
    --tag $TAG \
    --name $TAG \
    --description "$(cat $DESCRIPTION)"

for resource in "$@" ;do
    $GITHUB_RELEASE upload -u kubevirt -r cluster-network-addons-operator \
        --name $(basename $resource) \
        --tag $TAG \
        --file $resource
done
