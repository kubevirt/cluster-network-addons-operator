#!/bin/bash -e
toolchain=$(grep '^toolchain' go.mod | awk '{print $2}' | sed 's/^go//')
if [ -n "$toolchain" ]; then
    echo "$toolchain"
else
    grep '^go' go.mod | awk '{print $2}'
fi
