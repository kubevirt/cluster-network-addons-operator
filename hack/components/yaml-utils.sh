#!/usr/bin/env bash

set -xeo pipefail

function __yq() {
  docker run --rm -i -v ${PWD}:/workdir mikefarah/yq yq "$@"
}

function __get_parameter_from_yaml() {
	local yaml_file=$1
	local arg=$2
	__yq r ${yaml_file} ${arg} | xargs
}

function yaml-utils::set_param() {
	local yaml_file=$1
	local path=$2
	local value="$3"

	__yq w -i ${yaml_file} ${path} "${value}"
}

function yaml-utils::update_param() {
	local yaml_file=$1
	local path=$2
	local new_value="$3"

	local old_value=$(__get_parameter_from_yaml ${yaml_file} ${path})
	if [ ! -z "${old_value}" ]; then
		$(yaml-utils::set_param ${yaml_file} ${path} "${new_value}")
	else
		echo Error: ${path} is not found in ${yaml_file}
		exit 1
	fi
}

function yaml-utils::delete_param() {
	local yaml_file=$1
	local path=$2

	__yq d -i ${yaml_file} ${path} "${3}"
}

function yaml-utils::get_component_url() {
	local component=$1
	arg=components.\"${component}\".url
	__get_parameter_from_yaml components.yaml ${arg}
}

function yaml-utils::get_component_commit() {
	local component=$1
	arg=components.\"${component}\".commit
	__get_parameter_from_yaml components.yaml ${arg}
}

function yaml-utils::get_component_repo() {
	local url=$1
	#remove the prefix.
	echo ${url} | sed 's#https://\(.*\)#\1#'
}

# splits yaml to sub files by seperator '---'.
# files names are by line numbers
function yaml-utils::split_yaml_by_seperator() {
	local output_dir=$1
	local source_yaml=$2

	cd ${output_dir}
	awk '/\-\-\-/{f=NR".yaml"}; {print >f}' ${source_yaml}
}

# changes the yaml file names to be of format kind_names
function yaml-utils::rename_files_by_object() {
	local output_dir=$1

	for f in ${output_dir}/*; do
		local kind=$(__get_parameter_from_yaml ${f} kind)
		local name=$(__get_parameter_from_yaml ${f} metadata.name)
		mv ${f} ${output_dir}/${kind}_${name}.yaml

	done
}
