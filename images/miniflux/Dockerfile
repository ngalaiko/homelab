FROM docker.pkg.github.com/ngalaiko/server/base-alpine

ENV MINIFLUX_VERSION=2.0.18 \
	MINIFLUX_SHA256SUM=3807a19c32f15b9b438647adcf8dae4f9980c18846413960d59d76319ac00071

RUN apk add --no-cache --update \
    	curl \
	&& curl -OL "https://github.com/miniflux/miniflux/releases/download/${MINIFLUX_VERSION}/miniflux-linux-armv7" \
    && sha256sum miniflux-linux-armv7 \
    && echo "${MINIFLUX_SHA256SUM}  miniflux-linux-armv7" | sha256sum -c - \
    && chmod +x miniflux-linux-armv7

CMD ["./miniflux-linux-armv7"]
