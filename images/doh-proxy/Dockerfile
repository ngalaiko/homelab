FROM docker.pkg.github.com/ngalaiko/server/base-debian:latest

RUN apt-get update \
	&& apt-get -y --no-install-recommends install \
        python3 \
        python3-pip \
        python3-setuptools \
        dnsutils \
	&& pip3 install doh-proxy

COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

ENTRYPOINT [ "./entrypoint.sh" ]
