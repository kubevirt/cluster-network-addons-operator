#!/bin/bash -xe

destination=$1

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
go_mod_version=$("$script_dir/go-version.sh")
version="go${go_mod_version}"

tarball=$version.linux-amd64.tar.gz
url=https://dl.google.com/go/

mkdir -p $destination
curl -L $url/$tarball -o $destination/$tarball
tar -xf $destination/$tarball -C $destination
