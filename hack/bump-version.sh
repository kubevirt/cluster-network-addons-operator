#!/bin/bash -e

expected_types="(major|minor|patch)"
current_type=$1

bump() {
    local version="$1"
    local bump_type="$2"

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
}

if [[ ! $current_type =~ $expected_types ]]; then
    echo "Usage: $0 $expected_types"
    exit 1
fi

bump_$current_type
hack/version.sh
