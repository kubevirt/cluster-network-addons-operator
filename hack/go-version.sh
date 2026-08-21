#!/bin/bash -e
grep ^toolchain go.mod | awk '{print $2}' | sed 's/^go//'
