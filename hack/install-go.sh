#!/bin/bash -xe

destination=$1

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
go_mod_version=$("$script_dir/go-version.sh")
version="go${go_mod_version}"

arch=$(uname -m | sed 's/x86_64/amd64/')
tarball=$version.linux-$arch.tar.gz
url=https://dl.google.com/go/

mkdir -p $destination
curl -L $url/$tarball -o $destination/$tarball
tar -xf $destination/$tarball -C $destination
