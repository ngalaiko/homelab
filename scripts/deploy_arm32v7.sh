#!/usr/bin/env sh

set -e

echo "Deploying..."

files=""
for file in $(ls ./compose/arm32/*.yaml); do
    files="${files} -c ${file} "
done

docker stack deploy --prune --resolve-image never ${files} server

echo "Done."
