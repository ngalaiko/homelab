#!/usr/bin/env sh

git pull

files=""
for file in $(ls *.yml); do
    files="${files} -f ${file} "
done

docker-compose ${files} up --build -d --remove-orphans
