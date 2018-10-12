#!/usr/bin/env bash

set -e

echo "Building images..."

BUILD_IMAGES=(
    "../blog:ngalayko/blog"
    "../remark:ngalayko/remark"
    "../autoheal:ngalayko/autoheal"
    "../vpn:ngalayko/vpn"
)

docker login -u "${ENV_DOCKER_HUB_LOGIN}" -p "${ENV_DOCKER_HUB_PASSWORD}"

for build_image in "${BUILD_IMAGES[@]}"; do
    build="${build_image%:*}" 
    image="${build_image#*:}"

    echo "Building ${image}..."


    docker build "${build}" -t "${image}"
    docker push  "${image}"
done

echo "Removing unsused images..."

docker rm -v $(docker ps --filter status=exited -q 2>/dev/null)
docker rmi $(docker images --filter dangling=true -q 2>/dev/null)

echo "Done"
