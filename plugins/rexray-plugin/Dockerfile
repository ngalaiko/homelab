FROM arm32v7/debian:7.11-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        wget 

RUN wget -q https://dl.bintray.com/rexray/rexray/stable/0.11.3/rexray-Linux-ARMv7-0.11.3.tar.gz -O rexray.tar.gz \
        && tar -xvzf rexray.tar.gz -C /usr/bin \
        && rm rexray.tar.gz

RUN mkdir -p /run/docker/plugins /var/lib/libstorage/volumes

ADD config.yaml /etc/rexray/config.yaml

ENTRYPOINT ["rexray", "-c", "/etc/rexray/config.yaml"]
CMD ["--help"]
