#!/usr/bin/env bash

set -e

echo "Building images..."

# NOTE: ./server is the root
BUILD_IMAGES=(
    "./containers/dyn-dns:ngalayko/dyn-dns"
    "./containers/blog:ngalayko/blog"
    "./containers/dns-wrapper:ngalayko/dnscrypt"
    "./containers/goaccess:ngalayko/goaccess"
    "./containers/dns:ngalayko/pihole"
    "./containers/traefik:ngalayko/proxy"
    "./containers/remark:ngalayko/remark"
    "./containers/node-exporter:ngalayko/node-exporter"
    "./containers/grafana:ngalayko/grafana"
    "./containers/prometheus:ngalayko/prometheus"
    "./containers/nginx:ngalayko/nginx"
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
