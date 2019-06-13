#!/bin/bash -e

version=${OLM_VERSION:-0.10.0}

echo 'Install OLM on cluster'
./cluster/kubectl.sh apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/${version}/crds.yaml
./cluster/kubectl.sh apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/${version}/olm.yaml

echo 'Wait for OLM to become ready'
./cluster/kubectl.sh rollout status -w deployment/olm-operator -n olm
./cluster/kubectl.sh rollout status -w deployment/catalog-operator -n olm

retries=50
until [[ $retries == 0 || $new_csv_phase == 'Succeeded' ]]; do
    new_csv_phase=$(./cluster/kubectl.sh get csv -n olm packageserver.v${version} -o jsonpath='{.status.phase}' 2>/dev/null || echo 'Waiting for CSV to appear')
    if [[ $new_csv_phase != "$csv_phase" ]]; then
        csv_phase=$new_csv_phase
        echo "Package server phase: $csv_phase"
    fi
    sleep 1
    retries=$((retries - 1))
done

if [ $retries == 0 ]; then
    echo 'CSV "packageserver" failed to reach phase succeeded'
    exit 1
fi

./cluster/kubectl.sh rollout status -w deployment/packageserver -n olm
