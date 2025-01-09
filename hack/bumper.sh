#!/bin/bash -e

# package bump helper
# make sure you already have a local repo with remote upstream, and origin branches
# no untracked files allowed on folder
# it does reset --hard so beware, also pushes to save time

# ./hack/bumper.sh CVE-2021-38561 golang.org/x/text@v0.3.7 release-0.89

# to skip semver check (for example in case package is dropped from go.mod after updating)
# export SKIP=true
# optional: to update go lang level use for example
# export DESIRED_VERSION=1.18

SKIP=${SKIP:-false}

if [ $# -ne 3 ]; then
    echo "Syntax: $0 <CVE> <TARGET_PACK> <BR>"
    exit 1
fi

CVE=$1
TARGET_PACK=$2
BR=$3

PACK="${TARGET_PACK/@*}"
[ "$(git rev-parse --abbrev-ref HEAD)" != "$BR" ] && git checkout "$BR"
git reset --hard upstream/"$(git symbolic-ref --short HEAD)" # || true # try removing the suffix, if it doesnt work try git fetch upstream first
git pull upstream "$(git symbolic-ref --short HEAD)"
git push -u origin
go mod edit -dropreplace="${PACK}"
go mod edit -require="${TARGET_PACK}"

current_version=$(awk '$1 == "go" {print $2}' go.mod)
desired_version=${DESIRED_VERSION:-$current_version}
if [[ "$(printf "%s\n" "$current_version" "$desired_version" | sort -V | head -n1)" != "$desired_version" ]]; then
  echo "Updating Go version in go.mod to $desired_version"
  go mod edit -go=$desired_version
else
  echo "Go version in go.mod is already $desired_version or later. No update needed."
fi

if grep -q '^vendor:' Makefile; then
    rm -rf build/_output/
    make vendor
else
    go mod tidy
    go mod vendor
fi

if grep -q '^vet:' Makefile; then
    make vet
fi

if [ $SKIP != true ]; then
	required_version="${TARGET_PACK/*@/}"
	actual_version=$(go list -m -json $PACK | jq -r .Version)
	if [ "$(printf '%s\n' "$required_version" "$actual_version" | sort -V | head -n1)" != "$required_version" ]; then
	    echo "Actual version $actual_version is less than $required_version"
	    exit 1
	fi
fi

git checkout -b "${BR}_${CVE}_$(openssl rand -hex 4)"
git add . --all
git commit -s -m "$( [ "$BR" == "main" ] && echo "" || echo "[$BR] " )$CVE: Bump $PACK"

git push --set-upstream origin "$(git rev-parse --abbrev-ref HEAD)"

