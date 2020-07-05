#!/usr/bin/env bash

set -xeo pipefail

function git-utils::fetch_component() {
    local destination=$1
    local url=$2
    local commit=$3

    if [ ! -d ${destination} ]; then
        mkdir -p ${destination}
        git clone ${url} ${destination}
    fi

    (
        cd ${destination}
        git reset ${commit} --hard
    )
}

function git-utils::get_component_tag() {
    local component_dir=$1

    (
        cd ${component_dir}
        git describe --tags
    )
}

function git-utils::create_temp_path() {
	local component=$1
	mktemp -d -p /tmp -t deploy.${component}.XXXX
}
