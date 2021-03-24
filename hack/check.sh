#!/usr/bin/env bash

set -xe

if [[ -n "$(git status --porcelain)" ]] ; then
    echo "You have Uncommitted changes. Please commit the changes1"
    git status --porcelain
    git diff
    exit 1
fi
