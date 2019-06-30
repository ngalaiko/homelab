#!/usr/bin/env bash

set -e

SECRETS_DIR=./secrets

mkdir -p "${SECRETS_DIR}"

for v in $(env); do
    if [[ "${v}" =~ ^ENV_* ]]; then
        name="${v%=*}"
        value="${v#*=}"

        echo "${value}" > "${SECRETS_DIR}/${name/ENV_/}"
    fi
done

for secret_file in $(find "${SECRETS_DIR}" -type f); do
    file_name="$(basename ${secret_file})"
    secret_name="${file_name,,}" # lowercase
    secret_name="${secret_name//_/-}" # replace '_' with '-'

    # if secret doesn't exist
    if [ -z "$( kubectl get secrets | awk '{print $1}' | grep "^${secret_name}$")" ]; then
        # create secret
        kubectl create secret generic \
            "${secret_name}" \
            --from-file="${secret_file}"
    else
        # update secret
        kubectl get secret "${secret_name}" -o json \
            | jq \
            --arg value "$(echo -n $(cat "${secret_file}") | base64)" \
            ".data[\"${file_name}\"]=\$value" \
            | kubectl apply -f -
    fi
done

files=""
for file in $(find ./k8s -name '*.yaml'); do
    files="${files} -f ${file} "
done

kubectl apply --prune --all ${files}
