#!/usr/bin/env bash

set -e

for v in $(env); do
    if [[ "${v}" =~ ^ENV_* ]]; then
        VARS="${VARS} && export ${v}"
        echo $v
    fi
done

files=""
for file in $(find ./k8s -name '*.yaml'); do
    files="${files} -f ${file} "
done

kubectl apply --prune --all ${files}
