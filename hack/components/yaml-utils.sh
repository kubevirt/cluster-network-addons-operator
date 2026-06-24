#!/usr/bin/env bash

set -xeo pipefail

source hack/components/docker-utils.sh

export OCI_BIN=${OCI_BIN:-$(docker-utils::determine_cri_bin)}

function __yq() {
  yq "$@"
}

# Convert yq v3 path syntax to v4 syntax
# v3: path.to.array(field=="value").property or path.to.array(field==value).property
# v4: path.to.array[] | select(.field == "value") | .property
function yaml-utils::convert_path_v3_to_v4() {
	local path="$1"

	# Check if path contains v3-style filter syntax like (name=="value")
	# Use string operations instead of regex for the outer check
	if [[ "$path" == *"("*")"* ]]; then
		# Extract the part before the filter, the filter, and the part after
		local before_filter="${path%%(*}"
		local filter_and_after="${path#*(}"
		local filter="${filter_and_after%%)*}"
		local after_filter="${filter_and_after#*)}"

		# Parse the filter: field=="value" (with quotes)
		if [[ "$filter" =~ ([a-zA-Z_][a-zA-Z0-9_-]*)\ *==\ *\"([^\"]+)\" ]]; then
			local field="${BASH_REMATCH[1]}"
			local value="${BASH_REMATCH[2]}"
			# Convert to v4 syntax: before[] | select(.field == "value")
			echo "${before_filter}[] | select(.${field} == \"${value}\")${after_filter}"
		# Parse the filter: field==value (without quotes - bareword)
		elif [[ "$filter" =~ ([a-zA-Z_][a-zA-Z0-9_-]*)\ *==\ *([a-zA-Z0-9_-]+) ]]; then
			local field="${BASH_REMATCH[1]}"
			local value="${BASH_REMATCH[2]}"
			# In v4, barewords need to be quoted as strings
			echo "${before_filter}[] | select(.${field} == \"${value}\")${after_filter}"
		else
			# Can't parse filter, return original
			echo "$path"
		fi
	else
		# No filter, return as-is
		echo "$path"
	fi
}

function yaml-utils::get_param() {
	local yaml_file=$1
	local arg=$2

	# Convert v3 path syntax to v4
	local yq_path="$(yaml-utils::convert_path_v3_to_v4 "${arg}")"

	# Add leading dot if not present (unless it starts with '(' for filter expressions)
	# For pipe expressions, add the dot at the beginning before the first path component
	if [[ "${yq_path}" != .* ]] && [[ "${yq_path}" != \(* ]]; then
		yq_path=".${yq_path}"
	fi
	__yq eval "${yq_path}" "${yaml_file}"
}

function yaml-utils::set_param() {
	local yaml_file=$1
	local path=$2
	local value="$3"

	# Convert v3 path syntax to v4
	local yq_path="$(yaml-utils::convert_path_v3_to_v4 "${path}")"

	# For piped expressions (filters), need special handling for assignment
	# Format: (.path[] | select(.field == "value")).property
	if [[ "${yq_path}" == *\|*select* ]]; then
		# Find everything after the closing paren of select(...)
		# Use string manipulation instead of regex to avoid escaping issues
		local temp="${yq_path#*select\(}"
		local temp2="${temp#*)}"
		local after_select="$temp2"
		# Get everything up to and including select(...)
		local base_path="${yq_path%${after_select}}"

		# Add leading dot to base_path if not present
		if [[ "${base_path}" != .* ]]; then
			base_path=".${base_path}"
		fi

		# Construct the proper yq v4 assignment syntax
		yq_path="(${base_path})${after_select}"
	else
		# Add leading dot if not present (unless it starts with '(' or '.')
		if [[ "${yq_path}" != .* ]] && [[ "${yq_path}" != \(* ]]; then
			yq_path=".${yq_path}"
		fi
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
	elif [[ "$value" == \"*\" ]] && [[ "$value" != *"{{"* ]]; then
		# Quoted string literal without Go templates - preserve double-quote style
		# This handles cases like '"10m"' or '"restricted-v2"' which should preserve quotes
		local unquoted_value="${value:1:-1}"
		export YQ_VALUE="${unquoted_value}"
		__yq eval "${yq_path} = (strenv(YQ_VALUE) | . style=\"double\")" -i "${yaml_file}"
		unset YQ_VALUE
	elif { [[ "$value" == "["* ]] || [[ "$value" == "{"* ]]; } && [[ "$value" != *"{{"* ]]; then
		# Non-empty JSON array or object without Go templates - use from_json to parse it
		export YQ_VALUE="${value}"
		__yq eval "${yq_path} = (strenv(YQ_VALUE) | from_json)" -i "${yaml_file}"
		unset YQ_VALUE
	else
		# Regular value - use strenv
		export YQ_VALUE="${value}"
		# For template variables, use style="single" to ensure consistent quoting
		if [[ "$value" == *"{{"* ]]; then
			__yq eval "${yq_path} = (strenv(YQ_VALUE) | . style=\"single\")" -i "${yaml_file}"
			# yq v4 sometimes preserves existing quote styles even with style="single"
			# Normalize double-quoted template variables to single-quoted for consistency
			# Extract the last component of the path for use in sed pattern
			local key_pattern
			if [[ "${yq_path}" == *"."* ]]; then
				key_pattern="${yq_path##*.}"
			else
				key_pattern="${yq_path#.}"
			fi
			# Normalize double quotes to single quotes for this specific key
			sed -i "s/^\\(  *${key_pattern}: *\\)\"\\({{.*}}\\)\"\$/\\1'\\2'/g" "${yaml_file}"
		else
			__yq eval "${yq_path} = strenv(YQ_VALUE)" -i "${yaml_file}"
		fi
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

	# Convert v3 path syntax to v4
	local yq_path="$(yaml-utils::convert_path_v3_to_v4 "${path}")"

	# For piped expressions (filters), wrap in parentheses for deletion
	if [[ "${yq_path}" == *\|* ]]; then
		yq_path="(${yq_path})"
	fi

	# Add leading dot if not present (unless it starts with '(' for filter expressions)
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

		if [ "$(head -n 1 "${yaml_file}")" != "---" ]; then
			echo -e "---\n$(cat "${yaml_file}")" > "${yaml_file}"
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

	# Unquote template variables like {{ .Var }} that were quoted by yq v4's strenv
	yaml-utils::unquote_template_variables "${yaml_file}"
}

function yaml-utils::unquote_template_variables() {
	local yaml_file=$1

	# Remove quotes from template variables like {{ .Var }}
	# Template variables must be unquoted to work with Go templates
	# Use lazy matching for the content between {{ and }}
	sed -i 's/: *"\({{.*}}\)"/: \1/g' "${yaml_file}"
	sed -i "s/: *'\({{.*}}\)'/: \1/g" "${yaml_file}"

	# Remove outer single quotes from double-quoted strings
	# This handles cases like '""' -> "" and '"restricted-v2"' -> "restricted-v2"
	# For key-value pairs
	sed -i "s/: *'\\(\"[^\"]*\"\\)'/: \\1/g" "${yaml_file}"
	# For array items
	sed -i "s/- *'\\(\"[^\"]*\"\\)'/- \\1/g" "${yaml_file}"

	# Remove quotes from template variables in array items
	sed -i 's/- *"\({{.*}}\)"/- \1/g' "${yaml_file}"
	sed -i "s/- *'\({{.*}}\)'/- \1/g" "${yaml_file}"
}
