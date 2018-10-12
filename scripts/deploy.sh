#!/usr/bin/env sh

set -e

files=""
for file in $(ls *.yaml); do
    files="${files} -f ${file} "
done

docker-compose ${files} up --build -d --remove-orphans
