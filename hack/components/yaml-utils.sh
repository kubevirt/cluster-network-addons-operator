#!/usr/bin/env bash

set -xeo pipefail

source hack/components/docker-utils.sh

export OCI_BIN=${OCI_BIN:-$(docker-utils::determine_cri_bin)}

function __yq() {
  yq "$@"
}

function yaml-utils::get_param() {
	local yaml_file=$1
	local arg=$2
	# Add leading dot if not present (unless it starts with '(' for filter expressions)
	local yq_path="${arg}"
	if [[ "${yq_path}" != .* ]] && [[ "${yq_path}" != \(* ]]; then
		yq_path=".${yq_path}"
	fi
	__yq eval "${yq_path}" "${yaml_file}"
}

function yaml-utils::set_param() {
	local yaml_file=$1
	local path=$2
	local value="$3"

	# Add leading dot if not present (unless it starts with '(' for filter expressions)
	local yq_path="${path}"
	if [[ "${yq_path}" != .* ]] && [[ "${yq_path}" != \(* ]]; then
		yq_path=".${yq_path}"
	fi

	# Check if value is empty object or empty array
	if [[ "$value" == "{}" ]] || [[ "$value" == "[]" ]]; then
		# Empty objects/arrays should not use strenv
		__yq eval "${yq_path} = ${value}" -i "${yaml_file}"
	elif [[ "$value" == "true" ]] || [[ "$value" == "false" ]]; then
		# Boolean values - use directly without strenv to preserve type
		__yq eval "${yq_path} = ${value}" -i "${yaml_file}"
	elif [[ "$value" =~ ^[0-9]+$ ]]; then
		# Numeric values - use directly without strenv to preserve type
		__yq eval "${yq_path} = ${value}" -i "${yaml_file}"
	elif { [[ "$value" == "["* ]] || [[ "$value" == "{"* ]]; } && [[ "$value" != *"{{"* ]]; then
		# Non-empty JSON array or object without Go templates - use from_json to parse it
		export YQ_VALUE="${value}"
		__yq eval "${yq_path} = (strenv(YQ_VALUE) | from_json)" -i "${yaml_file}"
		unset YQ_VALUE
	else
		# Regular value (including JSON with Go templates) - use strenv
		export YQ_VALUE="${value}"
		__yq eval "${yq_path} = strenv(YQ_VALUE)" -i "${yaml_file}"
		unset YQ_VALUE
	fi

	# yq write removes the heading --- from the yaml, so we re-add it.
	yaml-utils::append_delimiter "${yaml_file}"
}

function yaml-utils::update_param() {
	local yaml_file=$1
	local path=$2
	local new_value="$3"

	local old_value=$(yaml-utils::get_param "${yaml_file}" "${path}")
	if [ ! -z "${old_value}" ]; then
		yaml-utils::set_param "${yaml_file}" "${path}" "${new_value}"
	else
		echo Error: ${path} is not found in ${yaml_file}
		exit 1
	fi
}

function yaml-utils::delete_param() {
	local yaml_file=$1
	local path=$2

	# Add leading dot if not present (unless it starts with '(' for filter expressions)
	local yq_path="${path}"
	if [[ "${yq_path}" != .* ]] && [[ "${yq_path}" != \(* ]]; then
		yq_path=".${yq_path}"
	fi
	__yq eval "del(${yq_path})" -i "${yaml_file}"

	# yq write removes the heading --- from the yaml, so we re-add it.
	yaml-utils::append_delimiter "${yaml_file}"
}

function yaml-utils::get_component_url() {
	local component=$1
	arg=components.\"${component}\".url
	yaml-utils::get_param components.yaml "${arg}"
}

function yaml-utils::get_component_commit() {
	local component=$1
	arg=components.\"${component}\".commit
	yaml-utils::get_param components.yaml "${arg}"
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
		local kind=$(yaml-utils::get_param "${f}" kind)
		local name=$(yaml-utils::get_param "${f}" metadata.name)
		mv "${f}" "${output_dir}/${kind}_${name}.yaml"

	done
}

function yaml-utils::remove_single_quotes_from_yaml() {
	local yaml_file=$1

	sed -i "s/'//g" ${yaml_file}
}

function yaml-utils::unquote_template_variables() {
	local yaml_file=$1

	# Remove quotes from template variables like {{ .Var }}
	# Template variables must be unquoted to work with Go templates
	sed -i 's/: *"\({{[^}]*}}\)"/: \1/g' "${yaml_file}"
	sed -i "s/: *'\({{[^}]*}}\)'/: \1/g" "${yaml_file}"

	# Remove outer single quotes from double-quoted strings
	# This handles cases like '""' -> "" and '"restricted-v2"' -> "restricted-v2"
	# For key-value pairs
	sed -i "s/: *'\\(\"[^\"]*\"\\)'/: \\1/g" "${yaml_file}"
	# For array items
	sed -i "s/- *'\\(\"[^\"]*\"\\)'/- \\1/g" "${yaml_file}"

	# Remove quotes from template variables in array items
	sed -i 's/- *"\({{[^}]*}}\)"/- \1/g' "${yaml_file}"
	sed -i "s/- *'\({{[^}]*}}\)'/- \1/g" "${yaml_file}"
}
