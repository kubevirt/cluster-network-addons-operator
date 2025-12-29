#!/bin/bash -e

expected_types="(major|minor|patch|current_rc|major_rc|minor_rc|patch_rc)"
current_type=$1

bump() {
    local version="$1"
    local bump_type="$2"

    # If version is a pre-release (RC), strip it and decide what to do
    if [[ $version =~ -rc-[0-9]+$ ]]; then
        local base_version=$(echo $version | sed 's/-rc-[0-9]\+$//')

        # Patch bump always removes the RC
        if [[ $bump_type == "patch" ]]; then
            echo $base_version
            return
        fi

        # Minor/Major bumps: strip RC, then apply the bump
        version=$base_version
    fi

    case $bump_type in
        major)
            local major=$(echo $version | sed 's/^\([0-9]\+\)\..*$/\1/')
            major=$((++major))
            echo "${major}.0.0"
            ;;
        minor)
            local major_minor=$(echo $version | sed 's/^\([0-9]\+\)\.\([0-9]\+\)\..*$/\1.\2/')
            local major=$(echo $major_minor | cut -d. -f1)
            local minor=$(echo $major_minor | cut -d. -f2)
            minor=$((++minor))
            echo "${major}.${minor}.0"
            ;;
        patch)
            local major_minor_patch=$(echo $version | sed 's/^\([0-9]\+\)\.\([0-9]\+\)\.\([0-9]\+\)$/\1.\2.\3/')
            local major=$(echo $major_minor_patch | cut -d. -f1)
            local minor=$(echo $major_minor_patch | cut -d. -f2)
            local patch=$(echo $major_minor_patch | cut -d. -f3)
            patch=$((++patch))
            echo "${major}.${minor}.${patch}"
            ;;
    esac
}

bump_major() {
    local version=$(hack/version.sh)
    local new_version=$(bump "$version" "major")
    ./hack/version.sh $new_version
}

bump_minor() {
    local version=$(hack/version.sh)
    local new_version=$(bump "$version" "minor")
    ./hack/version.sh $new_version
}

bump_patch() {
    local version=$(hack/version.sh)
    local new_version=$(bump "$version" "patch")
    ./hack/version.sh $new_version
}

bump_major_rc() {
    local version=$(hack/version.sh)
    local new_version=$(bump "$version" "major")
    ./hack/version.sh "${new_version}-rc-0"
}

bump_minor_rc() {
    local version=$(hack/version.sh)
    local new_version=$(bump "$version" "minor")
    ./hack/version.sh "${new_version}-rc-0"
}

bump_patch_rc() {
    local version=$(hack/version.sh)
    local new_version=$(bump "$version" "patch")
    ./hack/version.sh "${new_version}-rc-0"
}

bump_current_rc() {
    local version=$(hack/version.sh)
    local new_version

    if [[ $version =~ -rc-[0-9]+$ ]]; then
        local rc_number=$(echo $version | sed 's/.*-rc-\([0-9]\+\)$/\1/')
        rc_number=$((++rc_number))
        new_version=$(echo $version | sed "s/-rc-[0-9]\+$/-rc-$rc_number/")
    else
        new_version="${version}-rc-0"
    fi

    ./hack/version.sh $new_version
}

if [[ ! $current_type =~ $expected_types ]]; then
    echo "Usage: $0 $expected_types"
    exit 1
fi

bump_$current_type
hack/version.sh
