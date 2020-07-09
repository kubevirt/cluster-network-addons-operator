#!/usr/bin/env bash

set -xeo pipefail

function __get_parameter_from_yaml() {
	local arg=$1
	cat components.yaml | docker run -i --rm evns/yq $arg | xargs
}

function yaml-utils::get_component_url() {
	local component=$1
	arg=.components.\"${component}\".url
	__get_parameter_from_yaml $arg
}

function yaml-utils::get_component_commit() {
	local component=$1
	arg=.components.\"${component}\".commit
	__get_parameter_from_yaml $arg
}

function yaml-utils::get_component_repo() {
	local url=$1
	#remove the prefix.
	echo ${url} | sed 's#https://\(.*\)#\1#'
}
