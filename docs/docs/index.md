---
title: go-oidc-mock Docs
slug: /
description: Mock OIDC/OAuth server for tests.
---

# go-oidc-mock

`go-oidc-mock` is a configurable mock OIDC/OAuth server for integration and
end-to-end tests. It gives test environments a protocol-shaped stand-in identity
provider instead of requiring each application to mock its own auth layer.

This first pass is intentionally small: discovery and JWKS work, and the OAuth
flow endpoints are present but return `501 Not Implemented`.

## Quick Start

```sh
docker compose up --build

curl -sS localhost:8080/.well-known/openid-configuration
curl -sS localhost:8080/jwks.json
curl -sS -o /dev/null -w '%{http_code}\n' localhost:8080/oauth2/authorize
```

The default issuer is `http://localhost:8080`. Override it with
`GO_OIDC_MOCK_ISSUER_URL` or `--issuer-url` when clients reach the mock through a
different URL.

## Endpoints

- `GET /.well-known/openid-configuration`
- `GET /jwks.json`
- `GET /oauth2/authorize` returns `501`
- `POST /oauth2/token` returns `501`
- `GET /userinfo` returns `501`

Operational endpoints include `/healthz`, `/readyz`, `/metrics`, `/docs`, and
`/openapi.yaml`.

## API Reference

The [API Reference](api.md) is generated from the committed OpenAPI
specification. A running server also serves interactive docs at `/docs`.
