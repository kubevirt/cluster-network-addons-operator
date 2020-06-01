#!/bin/bash -xe
tools_file=$1
tools=$(grep "_" $tools_file |  sed 's/.*_ *"//' | sed 's/"//g')
for tool in $tools; do
    $GO install ./vendor/$tool
done
