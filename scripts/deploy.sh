#!/usr/bin/env sh

set -e

echo "Deploying..."

files=""
for file in $(ls *.yaml); do
    files="${files} -f ${file} "
done

COMPOSE_HTTP_TIMEOUT=120 docker-compose ${files} up --build -d --remove-orphans

echo "Done."
