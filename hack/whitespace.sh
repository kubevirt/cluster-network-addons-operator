#!/usr/bin/env bash

set -e

function fix() {
    git ls-files -- ':!vendor/' | xargs sed --follow-symlinks -i 's/[[:space:]]*$//'
}

function check() {
    invalid_files=$(git ls-files -- ':!vendor/' | xargs egrep -Hn " +$" || true)
    if [[ $invalid_files ]]; then
        echo 'Found trailing whitespaces. Please remove trailing whitespaces using `make fmt`:'
        echo "$invalid_files"
        return 1
    fi
}

if [ "$1" == "--fix" ]; then
    fix
else
    check
fi
