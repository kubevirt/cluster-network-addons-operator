#!/bin/bash -xe

destination=$1

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
go_mod_path="$script_dir/../go.mod"

go_mod_version=$(grep '^toolchain' "$go_mod_path" 2>/dev/null | awk '{print $2}' | sed 's/^go//' || true)
if [ -z "$go_mod_version" ]; then
    go_mod_version=$(grep '^go ' "$go_mod_path" | awk '{print $2}')
fi

if [[ "$go_mod_version" =~ ^[0-9]+\.[0-9]+$ ]]; then
    go_mod_version="${go_mod_version}.0"
fi

version="go${go_mod_version}"

tarball=$version.linux-amd64.tar.gz
url=https://dl.google.com/go/

mkdir -p $destination
curl -L $url/$tarball -o $destination/$tarball
tar -xf $destination/$tarball -C $destination
