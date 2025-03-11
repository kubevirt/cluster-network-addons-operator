#!/bin/bash

# the script print package usage of all the repos/brances/go.mods
# it allows to check which ones should be updated if any
#
# usage ./checkmod.sh <package>
# for example ./checkmod.sh google.golang.org/grpc
# a default package is included for easy sanity check

MOD=${1:-"golang.org/x/net"}

REPO=()
RELEASES=()
GOMOD=()

REPO+=("https://raw.githubusercontent.com/kubevirt/cluster-network-addons-operator")
RELEASES+=("main release-0.97 release-0.95 release-0.93 release-0.91 release-0.89 release-0.79")
GOMOD+=("go.mod")

REPO+=("https://raw.githubusercontent.com/containernetworking/plugins")
RELEASES+=("main")
GOMOD+=("go.mod")

REPO+=("https://raw.githubusercontent.com/kubevirt/bridge-marker")
RELEASES+=("main")
GOMOD+=("go.mod")

REPO+=("https://raw.githubusercontent.com/k8snetworkplumbingwg/ovs-cni")
RELEASES+=("main release-0.36 release-0.34 release-0.32 release-0.31 release-0.29")
GOMOD+=("go.mod")

REPO+=("https://raw.githubusercontent.com/k8snetworkplumbingwg/kubemacpool")
RELEASES+=("main release-0.45 release-0.44 release-0.43 release-0.42 release-0.41 release-0.39")
GOMOD+=("go.mod")

REPO+=("https://raw.githubusercontent.com/kubevirt/kubesecondarydns")
RELEASES+=("main")
GOMOD+=("go.mod")

REPO+=("https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-dynamic-networks-controller")
RELEASES+=("main")
GOMOD+=("go.mod")

REPO+=("https://raw.githubusercontent.com/kubevirt/macvtap-cni")
RELEASES+=("main")
GOMOD+=("go.mod")

for ((i=0; i<${#REPO[@]}; i++)); do
    repo_url="${REPO[i]}"
    releases="${RELEASES[i]}"
    gomods="${GOMOD[i]}"

    for release in $releases; do
        for gomod in $gomods; do
            go_mod_url="$repo_url/$release/$gomod"
            if curl -s -o /dev/null -I -w "%{http_code}" "$go_mod_url" | grep -q 200; then
                result=$(curl -s "$go_mod_url" | grep "$MOD")
                if [ $? -eq 0 ]; then
                    echo $go_mod_url $result
                fi
            else
                echo "ERROR: go.mod file not found in $go_mod_url"
                exit 1
            fi
        done
    done
done

