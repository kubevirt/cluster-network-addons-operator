#!/usr/bin/env bash

set -xe

function fetch_component() {
    destination=$1
    url=$2
    commit=$3

    if [ ! -d ${destination} ]; then
        mkdir -p ${destination}
        git clone ${url} ${destination}
    fi

    (
        cd ${destination}
        git fetch origin
        git reset --hard
        git checkout ${commit}
    )
}

function get_component_tag() {
    component_dir=$1

    (
        cd ${component_dir}
        git describe --tags
    )
}