#!/usr/bin/env bash

set -xe

find ./k8s/ \
    -name '*.yaml' \
    -exec kubectl apply -f {} \;
