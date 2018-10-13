#!/usr/bin/env sh

set -e

echo "Deploying..."

files=""
for file in $(ls *.yaml); do
    files="${files} -f ${file} "
done

docker-compose --verbose ${files} up --build -d --remove-orphans

echo "Done."
