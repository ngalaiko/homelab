#!/usr/bin/env sh

set -e

echo "Deploying..."

files=""
for file in $(ls ./compose/amd64/*.yaml); do
    files="${files} -f ${file} "
done

docker-compose ${files} up --build -d --remove-orphans

echo "Done."
