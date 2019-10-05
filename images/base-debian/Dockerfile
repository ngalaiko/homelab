FROM debian:10.1-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        curl \
        ca-certificates

RUN curl -OL "https://github.com/multiarch/qemu-user-static/releases/download/v4.1.0-1/qemu-arm-static.tar.gz" \
      && echo "3ebf9fe9030e8a32e11c4db3eee3385cbd7dac27a3b026152aa7183bfef33746  qemu-arm-static.tar.gz" | sha256sum -c - \
      && tar -xvzf qemu-arm-static.tar.gz \
      && mv qemu-arm-static /usr/bin
