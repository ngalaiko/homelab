#!/usr/bin/env bash

set -e

for v in $(env); do
    if [[ "${v}" =~ ^ENV_* ]]; then
        name="${v%=*}"
        value="${v#*=}"

        echo "${value}" > "./secrets/${name/ENV_/}"
    fi
done

files=""
for file in $(find ./k8s -name '*.yaml'); do
    files="${files} -f ${file} "
done

kubectl apply --prune --all ${files}
