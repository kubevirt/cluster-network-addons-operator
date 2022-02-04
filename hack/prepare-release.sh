#!/usr/bin/env bash

set -ex

version_type=$1
previous_version=$(hack/version.sh)
released_version=$(hack/bump-version.sh $version_type)
prefixed_previous_version="v${previous_version}"
prefixed_released_version="v${released_version}"
commits=$(git log --pretty=format:"* %s" $prefixed_previous_version..HEAD)

echo 'Build manifests for the new release'
VERSION=${released_version} IMAGE_TAG=${prefixed_released_version} make gen-manifests
git add manifests/cluster-network-addons/${released_version}

echo 'Upgrade README.md with the released manifests'
sed -i "s/\(.*kubectl apply.*\)${prefixed_previous_version}\(.*\)/\1${prefixed_released_version}\2/g" README.md
sed -i "s/\(.*startingCSV.*\)${prefixed_previous_version}\(.*\)/\1${prefixed_released_version}\2/g" README.md

echo 'Generating new release for workflow e2e tests'
cp test/releases/99.0.0.go test/releases/${released_version}.go
git add test/releases/${released_version}.go
sed -i "s/Version: \"99.0.0\",/Version: \"${released_version}\",/" test/releases/${released_version}.go

echo 'Bump versions in Makefile'
sed -i "s/VERSION_REPLACES ?= .*/VERSION_REPLACES ?= ${released_version}/" Makefile
