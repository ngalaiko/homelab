#!/usr/bin/env sh

set -e

echo "Deploying..."

ARCH="$1"
if [ -z "${ARCH}" ]; then
    echo "specify archetecture: arm32 or amd64"
    exit 0
fi

files=""
for file in $(ls ./compose/"${ARCH}"/*.yaml); do
    files="${files} -f ${file} "
done

docker-compose ${files} up --build -d --remove-orphans

echo "Done."
