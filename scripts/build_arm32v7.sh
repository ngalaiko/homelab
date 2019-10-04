#!/usr/bin/env bash

set -e

echo "Building images..."

# NOTE: ./server is the root
BUILD_IMAGES=(
    "./containers/doh-proxy:ngalayko/doh-proxy"
    "./containers/fathom:ngalayko/fathom"
    "./containers/miniflux:ngalayko/miniflux"
    "./containers/openvpn:ngalayko/openvpn"
    "./containers/pihole:ngalayko/pihole"
    "./containers/postgres:ngalayko/postgres"
    "./containers/remark:ngalayko/remark"
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
