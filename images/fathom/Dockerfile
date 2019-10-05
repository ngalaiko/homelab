FROM docker.pkg.github.com/ngalaiko/server/base-alpine

ENV FATHOM_VERSION=1.2.1 \
	FATHOM_SHA256SUM=d09a81f8bd7c844a4214a500a3de8c20b3fbff292b5a0fa648e9efdc05ecd769

RUN apk add --no-cache --update \
    	curl \
	&& curl -OL "https://github.com/usefathom/fathom/releases/download/v${FATHOM_VERSION}/fathom_${FATHOM_VERSION}_linux_amd64.tar.gz" \
    && sha256sum fathom_${FATHOM_VERSION}_linux_amd64.tar.gz \
    && echo "${FATHOM_SHA256SUM}  fathom_${FATHOM_VERSION}_linux_amd64.tar.gz" | sha256sum -c - \
    && tar -xvzf fathom_${FATHOM_VERSION}_linux_amd64.tar.gz

CMD ["./fathom", "server"]
