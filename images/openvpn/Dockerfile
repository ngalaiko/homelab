FROM docker.pkg.github.com/ngalaiko/server/base-alpine

RUN apk add --update \
        openssl \
        easy-rsa \
        openvpn \
        iptables \
        bash

# Configure tun
RUN mkdir -p /dev/net && \
     mknod /dev/net/tun c 10 200 
