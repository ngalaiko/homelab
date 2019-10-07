FROM alpine as downloader
ENV REMARK_VERSION=1.3.2
ENV REMARK_SHA256SUM=e95b0dc7a576aed28852a67637b760b518bbda96fe56a227aec3f1e266f91ba7
WORKDIR /
RUN apk add --no-cache --update \
    curl
RUN curl -OL "https://github.com/umputun/remark/archive/v${REMARK_VERSION}.tar.gz" \
    && sha256sum v${REMARK_VERSION}.tar.gz \
    && echo "${REMARK_SHA256SUM}  v${REMARK_VERSION}.tar.gz" | sha256sum -c - \
    && tar -xvzf v${REMARK_VERSION}.tar.gz \
    && mv remark-${REMARK_VERSION} remark \
    && chmod +x /remark/entrypoint.sh

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
ARG NODE_ENV=production
COPY --from=build-frontend-deps /srv/frontend/node_modules /srv/frontend/node_modules
COPY --from=downloader /remark/frontend /srv/frontend
RUN cd /srv/frontend && \
    npm run build && \
    rm -rf ./node_modules


FROM docker.pkg.github.com/ngalaiko/server/base-alpine

WORKDIR /srv

COPY entrypoint.sh /entrypoint.sh
COPY --from=downloader /remark/backend/scripts/backup.sh /usr/local/bin/backup
COPY --from=downloader /remark/backend/scripts/restore.sh /usr/local/bin/restore
COPY --from=downloader /remark/backend/scripts/import.sh /usr/local/bin/import
RUN chmod +x /entrypoint.sh /usr/local/bin/backup /usr/local/bin/restore /usr/local/bin/import

COPY --from=build-backend /build/backend/remark42 /srv/remark42
COPY --from=build-frontend /srv/frontend/public/ /srv/web

RUN ln -s /srv/remark42 /usr/bin/remark42

HEALTHCHECK --interval=30s --timeout=3s CMD curl --fail http://localhost:8080/ping || exit 1

ENTRYPOINT ["/entrypoint.sh"]
