#!/usr/bin/env bash

set -e

echo "Building images..."

# NOTE: ./server is the root
BUILD_IMAGES=(
    "./containers/proxy:ngalayko/proxy"
    "./containers/blog:ngalayko/blog"
    "./containers/analytics:ngalayko/matomo"
    "./containers/nginx:ngalayko/nginx"
    "./containers/dns:ngalayko/pihole"
    "./containers/mysql:ngalayko/mysql"
)

docker login -u "${DOCKER_HUB_LOGIN}" -p "${DOCKER_HUB_PASSWORD}"

for build_image in "${BUILD_IMAGES[@]}"; do
    build="${build_image%:*}" 
    image="${build_image#*:}"

    echo "Building ${image}..."

    docker build -f "${build}/Dockerfile.arm32v7" "${build}" -t "${image}:arm32v7"
    docker push "${image}:arm32v7"
done

echo "Done"
