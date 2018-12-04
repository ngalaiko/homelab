FROM golang:1.10-alpine as builder
COPY . /go/src/github.com/vieux/docker-volume-sshfs
WORKDIR /go/src/github.com/vieux/docker-volume-sshfs
RUN set -ex \
    && apk add --no-cache --virtual .build-deps \
    gcc libc-dev \
    && GOOS=linux GOARCH=arm GOARM=7 go build -o /go/bin/docker-volume-sshfs --ldflags '-extldflags "-static"' \
    && apk del .build-deps
CMD ["/go/bin/docker-volume-sshfs"]

FROM apcheamitru/arm32v7-alpine:3.7

RUN apk update && apk add sshfs
RUN mkdir -p /run/docker/plugins /mnt/state /mnt/volumes
COPY --from=builder /go/bin/docker-volume-sshfs .
CMD ["docker-volume-sshfs"]
