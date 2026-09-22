#!/usr/bin/env bash
set -e

PROJECT_ROOT="$(readlink -e $(dirname "$BASH_SOURCE[0]")/../)"
DEPLOY_DIR="${DEPLOY_DIR:-${PROJECT_ROOT}/manifests}"
CONTAINER_PREFIX="${CONTAINER_PREFIX:-quay.io/kubevirt}"
CONTAINER_TAG="${CONTAINER_TAG:-latest}"
IMAGE_PULL_POLICY="${IMAGE_PULL_POLICY:-Always}"

templates=$(cd ${PROJECT_ROOT}/templates && find . -type f -name "*.yaml.in")
for template in $templates; do
	infile="${PROJECT_ROOT}/templates/${template}"

	dir="$(dirname ${DEPLOY_DIR}/${template})"
	dir=${dir/VERSION/$VERSION}
	mkdir -p ${dir}

	file="${dir}/$(basename -s .in $template)"
	file=${file/VERSION/$VERSION}
	rendered=$( \
		${PROJECT_ROOT}/build/_output/bin/manifest-templator \
		--version=${VERSION} \
		--version-replaces=${VERSION_REPLACES} \
		--operator-version=${VERSION} \
		--container-prefix=${CONTAINER_PREFIX} \
		--container-tag=${CONTAINER_TAG} \
		--image-pull-policy=${IMAGE_PULL_POLICY} \
		--input-file=${infile} \
	)
	if [[ ! -z "$rendered" ]]; then
		echo -e "$rendered" > $file
	fi
done
