#!/usr/bin/env bash

set -e

files=""
for file in $(find ./k8s -name '*.yaml'); do
    files="${files} -f ${file} "
done

kubectl apply --prune --all ${files}
