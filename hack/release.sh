#!/usr/bin/env bash

set -e

previous_version=$1
released_version=$2
future_version=$3
origin_remote=$4
fork_remote=$5

echo 'Fetching origin'
git fetch ${origin_remote}

echo 'Checkout origin/master into release branch'
git checkout ${origin_remote}/master -b release_${released_version}

echo 'Build manifests for the new release'
VERSION=${released_version} IMAGE_TAG=${released_version} make gen-manifests

echo 'Add new manifests to the source tree'
git add manifests/

echo 'Upgrade README.md with the released manifests'
sed -i "s/\(.*kubectl apply.*\)${previous_version}\(.*\)/\1${released_version}\2/g" README.md
sed -i "s/\(.*startingCSV.*\)${previous_version}\(.*\)/\1${released_version}\2/g" README.md

echo 'Add updated readme to the source tree'
git add README.md

echo 'Generating new release for workflow e2e tests'
cp tests/releases/${released_version}.go tests/releases/${future_version}.go
sed -i "s/Version: \"${released_version}\",/Version: \"${future_version}\",/" tests/releases/${future_version}.go

echo 'Committing new master release module'
git add tests/releases/${future_version}.go

echo 'Commit new manifests, this commit is to be tagged with the new release'
git commit -s -m "release ${released_version} - update manifests"

echo 'Bump versions in Makefile'
sed -i "s/VERSION ?= .*/VERSION ?= ${future_version}/" Makefile
sed -i "s/VERSION_REPLACES ?= .*/VERSION_REPLACES ?= ${released_version}/" Makefile

echo 'Commit Makefile with bumped versions'
git add Makefile

echo 'Commit updated Makefile'
git commit -s -m "release ${released_version} - bump versions in Makefile"

echo 'Push changes to forked repo'
git push ${fork_remote}

echo 'The rest is on you, open a new PR. Once the PR is merged, do not forget to tag the first of these commits, include change log and upload released manifests'
