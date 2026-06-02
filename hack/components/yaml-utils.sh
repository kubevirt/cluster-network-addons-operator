#!/usr/bin/env bash

set -xeo pipefail

source hack/components/docker-utils.sh

export OCI_BIN=${OCI_BIN:-$(docker-utils::determine_cri_bin)}
export YQ_IMAGE=${YQ_IMAGE:-docker.io/mikefarah/yq:3.3.4}

# Pre-pull the yq image with retry logic to avoid failures during setup
function yaml-utils::prepull_yq_image() {
  local max_attempts=3
  local attempt=1
  local wait_time=5

  echo "Pre-pulling yq image: ${YQ_IMAGE}"

  while [ $attempt -le $max_attempts ]; do
    echo "Attempt $attempt of $max_attempts to pull ${YQ_IMAGE}"
    if ${OCI_BIN} pull ${YQ_IMAGE}; then
      echo "Successfully pulled ${YQ_IMAGE}"
      return 0
    else
      echo "Failed to pull ${YQ_IMAGE} on attempt $attempt"
      if [ $attempt -lt $max_attempts ]; then
        echo "Waiting ${wait_time} seconds before retry..."
        sleep $wait_time
        wait_time=$((wait_time * 2))  # Exponential backoff
      fi
      attempt=$((attempt + 1))
    fi
  done

  echo "ERROR: Failed to pull ${YQ_IMAGE} after $max_attempts attempts"
  return 1
}

function __yq() {
  ${OCI_BIN} run --pull=missing --rm -v ${PWD}:/workdir:Z ${YQ_IMAGE} yq "$@"
}

function yaml-utils::get_param() {
	local yaml_file=$1
	local arg=$2
	__yq r ${yaml_file} ${arg}
}

function yaml-utils::set_param() {
	local yaml_file=$1
	local path=$2
	local value="$3"

	__yq w -i ${yaml_file} ${path} "${value}"

	# yq write removes the heading --- from the yaml, so we re-add it.
	yaml-utils::append_delimiter ${yaml_file}
}

function yaml-utils::update_param() {
	local yaml_file=$1
	local path=$2
	local new_value="$3"

	local old_value=$(yaml-utils::get_param ${yaml_file} ${path})
	if [ ! -z "${old_value}" ]; then
		yaml-utils::set_param ${yaml_file} ${path} "${new_value}"
	else
		echo Error: ${path} is not found in ${yaml_file}
		exit 1
	fi
}

function yaml-utils::delete_param() {
	local yaml_file=$1
	local path=$2

	__yq d -i ${yaml_file} ${path} "${3}"

	# yq write removes the heading --- from the yaml, so we re-add it.
	yaml-utils::append_delimiter ${yaml_file}
}

function yaml-utils::get_component_url() {
	local component=$1
	arg=components.\"${component}\".url
	yaml-utils::get_param components.yaml ${arg}
}

function yaml-utils::get_component_commit() {
	local component=$1
	arg=components.\"${component}\".commit
	yaml-utils::get_param components.yaml ${arg}
}

function yaml-utils::get_component_repo() {
	local url=$1
	#remove the prefix.
	echo ${url} | sed 's#https://\(.*\)#\1#'
}

function yaml-utils::append_delimiter() {
		local yaml_file=$1

		if [ "$(head -n 1 ${yaml_file})" != "---" ]; then
			echo -e "---\n$(cat ${yaml_file})" > ${yaml_file}
		fi
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
		local kind=$(yaml-utils::get_param ${f} kind)
		local name=$(yaml-utils::get_param ${f} metadata.name)
		mv ${f} ${output_dir}/${kind}_${name}.yaml

	done
}

function yaml-utils::remove_single_quotes_from_yaml() {
	local yaml_file=$1

	sed -i "s/'//g" ${yaml_file}
}
