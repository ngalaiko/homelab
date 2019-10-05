FROM alpine as downloader
ENV REMARK_VERSION=1.4.0
ENV REMARK_SHA256SUM=7d2b405a44ef877195a88da143e652e0a02fc7d48cf343e19c55a687d0237880
WORKDIR /
RUN apk add --no-cache --update \
    curl
RUN curl -OL "https://github.com/umputun/remark/archive/v${REMARK_VERSION}.tar.gz" \
    && sha256sum v${REMARK_VERSION}.tar.gz \
    && echo "${REMARK_SHA256SUM}  v${REMARK_VERSION}.tar.gz" | sha256sum -c - \
    && tar -xvzf v${REMARK_VERSION}.tar.gz \
    && mv remark-${REMARK_VERSION} remark

FROM umputun/baseimage:buildgo-latest as build-backend
COPY --from=downloader /remark/backend /build/backend
WORKDIR /build/backend
ENV GOFLAGS="-mod=vendor"
RUN GOOS=linux GOARCH=arm GOARM=7 go build -o remark42 -ldflags "-X main.revision=${version} -s -w" ./app


FROM node:alpine as build-frontend-deps
COPY --from=downloader /remark/frontend/package.json /srv/frontend/package.json
COPY --from=downloader /remark/frontend/package-lock.json  /srv/frontend/package-lock.json
RUN cd /srv/frontend && CI=true npm ci


FROM node:10.11-alpine as build-frontend
COPY --from=build-frontend-deps /srv/frontend/node_modules /srv/frontend/node_modules
COPY --from=downloader /remark/frontend /srv/frontend
RUN cd /srv/frontend && \
    if [ -z "$SKIP_FRONTEND_TEST" ] ; then npx run-p lint test build ; \
    else echo "skip frontend tests and lint" ; npm run build ; fi && \
    rm -rf ./node_modules


FROM hypriot/rpi-alpine:3.6
WORKDIR /srv
COPY --from=downloader /remark/backend/scripts/backup.sh /usr/local/bin/backup
COPY --from=downloader /remark/backend/scripts/restore.sh /usr/local/bin/restore
COPY --from=downloader /remark/backend/scripts/import.sh /usr/local/bin/import
COPY --from=build-backend /build/backend/remark42 /srv/remark42
COPY --from=build-frontend /srv/frontend/public/ /srv/web
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/ping || exit 1
CMD ["/srv/remark42", "server"]
