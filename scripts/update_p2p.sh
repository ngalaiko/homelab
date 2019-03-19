#!/usr/bin/env sh

set -e

echo "Deploying..."

HOST=galaiko.rocks docker stack deploy \
    --resolve-image never \
    -c ./compose/arm32/docker-compose.p2p.yaml.skip \
    server

echo "Done."
