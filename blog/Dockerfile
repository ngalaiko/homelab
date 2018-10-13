FROM golang:1.10-alpine as builder

RUN apk add --update --no-cache \
    git

RUN go get -v -u  github.com/magefile/mage && \
    go get -d -v -u github.com/gohugoio/hugo && \
    cd ${GOPATH}/src/github.com/gohugoio/hugo && \
    mage -v vendor && \
    mage -v install

WORKDIR /app
ADD ./ .
RUN hugo -t temple

FROM ngalayko/nginx

COPY --from=builder /app/public/ /usr/share/nginx/html
