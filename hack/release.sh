#!/bin/bash -e

function eventually() {
    retries=2
    interval=2
    n=0
    until [ "$n" -ge $retries ]
    do
        $@ && break
        n=$((n+1))
        sleep $interval
    done
    [ "$n" -lt $retries ]
}

make IMAGE_TAG=$TAG docker-build docker-push

git tag $TAG
git push https://github.com/kubevirt/cluster-network-addons-operator $TAG

$GITHUB_RELEASE release -u kubevirt -r cluster-network-addons-operator \
    --tag $TAG \
    --name $TAG \
    --description "$(./hack/render-release-notes.sh $(./hack/versions.sh -2) $TAG)"

for resource in "$@" ;do
    eventually $GITHUB_RELEASE upload -u kubevirt -r cluster-network-addons-operator \
        --name $(basename $resource) \
        --tag $TAG \
        -R \
        --file $resource
done
