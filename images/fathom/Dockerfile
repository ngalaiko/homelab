FROM alpine as downloader
ENV FATHOM_VERSION=1.2.1
ENV FATHOM_SHA256SUM=d63bf114928e0bed8b3ec08393f6a8c7c7ae62220e8c5140f7cfae1908f52668
WORKDIR /
RUN apk add --no-cache --update \
    curl
RUN curl -OL "https://github.com/usefathom/fathom/archive/v${FATHOM_VERSION}.tar.gz" \
    && sha256sum v${FATHOM_VERSION}.tar.gz \
    && echo "${FATHOM_SHA256SUM}  v${FATHOM_VERSION}.tar.gz" | sha256sum -c - \
    && tar -xvzf v${FATHOM_VERSION}.tar.gz \
    && mv fathom-${FATHOM_VERSION} fathom


FROM node:alpine AS assetbuilder
COPY --from=downloader /fathom /app
WORKDIR /app
RUN npm install && NODE_ENV=production ./node_modules/gulp/bin/gulp.js


FROM golang:latest AS binarybuilder
RUN go get -u github.com/gobuffalo/packr/packr
COPY --from=downloader /fathom /go/src/github.com/usefathom/fathom
WORKDIR /go/src/github.com/usefathom/fathom
COPY --from=assetbuilder /app/assets/build ./assets/build
RUN go get -u github.com/gobuffalo/packr/... \
    && GOOS=linux GOARCH=arm GOARM=7 ${GOPATH}/bin/packr build -v -ldflags '-w' -o fathom ./main.go


FROM docker.pkg.github.com/ngalaiko/server/base-alpine
COPY --from=binarybuilder /go/src/github.com/usefathom/fathom/fathom .
CMD ["./fathom", "server"]
