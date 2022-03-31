#!/usr/bin/env bash

set -ex

git_base_tag=$1
branch_name=$2

validate_input() {
	if [ -z "${git_base_tag}" ] || [ -z "${branch_name}" ]; then
		echo "input params missing: git_base_tag=${git_base_tag}, branch_name=${branch_name}. exiting"
		exit 1
	fi
	if ! git describe --tags ${git_base_tag}; then
		echo "git_base_tag=${git_base_tag} does not exist. existing"
    exit 1
  fi
}

update_gitaction_workflows() {
	echo 'Add release to gitActions workflows'
	source hack/components/yaml-utils.sh
	workflows_path=".github/workflows"

	yaml-utils::set_param ${workflows_path}/component-bumper.yaml 'jobs.bump.strategy.matrix.branch[+]' ${branch_name}

	for f in ${workflows_path}/{prepare-version,test-release-notes}.yaml; do
		yaml-utils::set_param ${f} 'on.workflow_dispatch.inputs.baseBranch.options[+]' ${branch_name}
	done
}

validate_input
update_gitaction_workflows

