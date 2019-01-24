# Dyn DNS [![Go Report Card](https://goreportcard.com/badge/github.com/ngalayko/dyn-dns)](https://goreportcard.com/report/github.com/ngalayko/dyn-dns)

## Description

A simple implementation for dynamic dns reconfiguration.

Small, no dependencies, easy to extend.

## How to install

```sh
go install github.com/ngalayko/dyn-dns/cmd/dyn-dns
```

## Quickstart

```sh
make build && ./bin/dyn-dns \
    --dnsProvider=cloudflare \
    --apiToken=<api_token> \
    --email=<user_email> \
    --zoneIdentifier=<zone_identifier>
    --domain=example.com \
    --record=@
```

## Supports

1. Fetchers:
    * [ipify](https://www.ipify.org/)
1. DNS providers:
    * [DigitalOcean](https://www.digitalocean.com/)
    * [Cloudflare](https://www.cloudflare.com/)
