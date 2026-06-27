---
title: go-oidc-mock Docs
slug: /
description: Mock OIDC/OAuth server for tests.
---

# go-oidc-mock

`go-oidc-mock` is a configurable mock OIDC/OAuth server for integration and
end-to-end tests. It gives test environments a protocol-shaped stand-in identity
provider instead of requiring each application to mock its own auth layer.

This implementation is intentionally small: discovery, JWKS, authorization,
token, and userinfo work for a minimal authorization-code flow with built-in
test clients and a mounted or fallback auto-approved user.

## Quick Start

```sh
docker compose up --build

curl -sS localhost:8080/.well-known/openid-configuration
curl -sS localhost:8080/jwks.json
```

The default issuer is `http://localhost:8080`. Override it with
`GO_OIDC_MOCK_ISSUER_URL` or `--issuer-url` when clients reach the mock through a
different URL.

The Compose stack mounts `examples/profiles/default.json` into
`/etc/go-oidc-mock/profiles`. Override the profile directory with
`GO_OIDC_MOCK_PROFILE_DIR` or `--profile-dir`.

## Endpoints

- `GET /.well-known/openid-configuration`
- `GET /jwks.json`
- `GET /oauth2/authorize`
- `POST /oauth2/token`
- `GET /userinfo`

Built-in authorization-code clients:

- `go-oidc-mock-public`: public client with S256 PKCE.
- `go-oidc-mock-confidential`: confidential client with secret
  `go-oidc-mock-secret`.

Both clients allow `http://localhost:3000/callback` and
`http://127.0.0.1:3000/callback`. Successful authorization requests are
auto-approved for the selected startup profile, falling back to subject
`go-oidc-mock-user` when no profiles are mounted.

Operational endpoints include `/healthz`, `/readyz`, `/metrics`, `/docs`, and
`/openapi.yaml`.

## API Reference

The [API Reference](api.md) is generated from the committed OpenAPI
specification. Standard OIDC/OAuth protocol endpoints are described by
discovery metadata, not OpenAPI. A running server also serves interactive docs
at `/docs`.
