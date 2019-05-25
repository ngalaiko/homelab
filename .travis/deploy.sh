#!/bin/bash

set -e

eval "$(ssh-agent -s)" # Start ssh-agent cache
chmod 600 .travis/deploy_rsa # Allow read access to the private key
ssh-add .travis/deploy_rsa # Add the private key to SSH

# move ENV_* variables to the remote server
# NOTE: they are moved without ENV_ prefix
VARS="echo env variables updated"
for v in $(env); do
    if [[ "${v}" =~ ^ENV_* ]]; then
        VARS="${VARS} && export ${v/ENV_/}"
    fi
done

ssh -i .travis/deploy_rsa $USER@$IP <<EOF
    ${VARS}
    cd ${DEPLOY_DIR}

    echo ${VARS}

    git pull --force

    ./scripts/deploy_arm32v7.sh
EOF
