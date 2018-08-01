#!/bin/bash

set -e

eval "$(ssh-agent -s)" # Start ssh-agent cache
chmod 600 .travis/id_rsa # Allow read access to the private key
ssh-add .travis/id_rsa # Add the private key to SSH

VARS="echo env variables updated"
for v in $(env); do
    if [[ "${v}" =~ ^ENV_* ]]; then
        VARS="${VARS} && export ${v/ENV_/}"
    fi
done

git config --global push.default matching
git remote add deploy ssh://git@$IP$DEPLOY_DIR
git push deploy master

ssh root@$IP <<EOF
    ${VARS}
    cd ${DEPLOY_DIR}
    git submodule update --init --recursive
    ./scripts/update.sh
    ./scripts/set_dns_password.sh ${ENV_DNS_PASSWORD}
EOF
