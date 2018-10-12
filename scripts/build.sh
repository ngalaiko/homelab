#!/usr/bin/env sh

set -e

BUILD_IMAGES=(
    "../blog:ngalaiko/blog"
    "../remark:ngalaiko/remark"
    "../autoheal:ngalaiko/autoheal"
    "../vpn:ngalaiko/vpn"
)

docker login -u "${DOCKER_HUB_LOGIN}" -p "${DOCKER_HUB_PASSWORD}"

for build_image in "${BUILD_IMAGES[@]}"; do
    image="${build_image#*:}"
    build="${build_image%:*}" 

    docker build "${build}" -t "${image}"
    docker push  "${image}"
done

