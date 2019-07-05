#!/usr/bin/env bash

set -e

echo "Building images..."

# NOTE: ./server is the root
BUILD_IMAGES=(
    "./containers/openvpn:ngalayko/openvpn"
    "./containers/pihole:ngalayko/pihole"
    "./containers/remark:ngalayko/remark"
    "./containers/home-assistant:ngalayko/home-assistant"
    "./containers/dyn-dns:ngalayko/dyn-dns"
    "./containers/grafana:ngalayko/grafana"
    "./containers/prometheus:ngalayko/prometheus"
    "./containers/docker-ap:ngalayko/docker-ap"
    "./containers/miniflux:ngalayko/miniflux"
    "./containers/traefik:ngalayko/proxy"
    "./containers/fathom:ngalayko/fathom"
    "./containers/postgres:ngalayko/postgres"
    "./containers/review-slots:ngalayko/review-slots"
    "./containers/doh-proxy:ngalayko/doh-proxy"
    "./containers/node-exporter:ngalayko/node-exporter"
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
