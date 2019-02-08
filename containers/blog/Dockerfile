FROM alpine:3.9

ENV HUGO_VERSION 0.44
ENV HUGO_BINARY hugo_extended_${HUGO_VERSION}_Linux-64bit.tar.gz

ENV GLIBC_VERSION 2.23-r3

RUN set -x \
  && apk add --no-cache --update \
	wget \
	ca-certificates \
	libstdc++ \
	imagemagick

# Install glibc.
RUN set -x \
	&& wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub \
	&& wget -q "https://github.com/sgerrand/alpine-pkg-glibc/releases/download/$GLIBC_VERSION/glibc-$GLIBC_VERSION.apk" \
	&& wget -q "https://github.com/sgerrand/alpine-pkg-glibc/releases/download/$GLIBC_VERSION/glibc-bin-$GLIBC_VERSION.apk" \
	&& wget -q "https://github.com/sgerrand/alpine-pkg-glibc/releases/download/$GLIBC_VERSION/glibc-i18n-$GLIBC_VERSION.apk" \
	&& apk --no-cache add \
		"glibc-$GLIBC_VERSION.apk" \
		"glibc-bin-$GLIBC_VERSION.apk" \
		"glibc-i18n-$GLIBC_VERSION.apk" 

# Install hugo.
RUN set -x \
	&& wget https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/${HUGO_BINARY} \
	&& tar xzf ${HUGO_BINARY} \
	&& rm -r ${HUGO_BINARY} \
	&& mv hugo /usr/bin

WORKDIR /app

COPY ./ .

RUN set -x \
   && mogrify -resize "800>x" -quality 82 -format jpg ./content/media/* \
   && hugo -v -t hermit
