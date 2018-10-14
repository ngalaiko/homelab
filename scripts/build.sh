#!/usr/bin/env bash

set -e

echo "Building images..."

# NOTE: ./server is the root
BUILD_IMAGES=(
    "./containers/proxy:ngalayko/proxy"
    "./containers/dns:ngalayko/pihole"
    "./containers/mysql:ngalayko/mysql"
    "./containers/nginx:ngalayko/nginx"
    "./containers/analytics:ngalayko/matomo"
    "./containers/remark:ngalayko/remark"
    "./containers/autoheal:ngalayko/autoheal"
    "./containers/vpn:ngalayko/vpn"
    "./containers/blog:ngalayko/blog"
)

docker login -u "${DOCKER_HUB_LOGIN}" -p "${DOCKER_HUB_PASSWORD}"

for build_image in "${BUILD_IMAGES[@]}"; do
    build="${build_image%:*}" 
    image="${build_image#*:}"

    echo "Building ${image}..."

    docker build "${build}" -t "${image}"
    docker push  "${image}"
done

echo "Done"
