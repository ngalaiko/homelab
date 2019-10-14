FROM alpine:3.9 as downloader

RUN apk add --update --no-cache \
        curl \
    && curl -OL "https://github.com/multiarch/qemu-user-static/releases/download/v4.1.0-1/qemu-arm-static.tar.gz" \
    && echo "3ebf9fe9030e8a32e11c4db3eee3385cbd7dac27a3b026152aa7183bfef33746  qemu-arm-static.tar.gz" | sha256sum -c - \
    && tar -xvzf qemu-arm-static.tar.gz \
    && mv qemu-arm-static /usr/bin


FROM hypriot/rpi-alpine:3.6

RUN apk add --update --no-cache \
    ca-certificates

COPY --from=downloader /usr/bin/qemu-arm-static /usr/bin/qemu-arm-static
