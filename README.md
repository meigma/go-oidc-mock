# go-oidc-mock

`go-oidc-mock` is a mock OIDC/OAuth server for tests. It is intended to run as a
stand-in identity provider so applications can exercise real protocol behavior
instead of mocking their own authentication integration.

The current implementation supports a minimal authorization-code flow. It serves
discovery, JWKS, authorization, token, and userinfo endpoints through the
embedded OIDC provider, auto-approving valid authorization requests for a mounted
profile or a built-in fallback mock user.

## Prerequisites

- Go 1.26.4
- Moon 2.x
- Python 3.14.3 and uv 0.11.0 for the MkDocs documentation project
- Docker for container builds and the local Compose smoke stack

## Quickstart

Run the server with Docker Compose:

```sh
docker compose up --build
```

The included Compose stack mounts `examples/profiles/default.json` into the
container at `/etc/go-oidc-mock/profiles`.

Then smoke the protocol endpoints:

```sh
curl -sS localhost:8080/healthz
curl -sS localhost:8080/readyz
curl -sS localhost:8080/.well-known/openid-configuration
curl -sS localhost:8080/jwks.json
```

Run a public-client PKCE flow:

```sh
VERIFIER=abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~
CHALLENGE="$(printf '%s' "$VERIFIER" | openssl dgst -sha256 -binary | openssl base64 -A | tr '+/' '-_' | tr -d '=')"

curl -iG http://localhost:8080/oauth2/authorize \
  --data-urlencode response_type=code \
  --data-urlencode client_id=go-oidc-mock-public \
  --data-urlencode redirect_uri=http://localhost:3000/callback \
  --data-urlencode scope="openid profile email" \
  --data-urlencode state=smoke \
  --data-urlencode code_challenge="$CHALLENGE" \
  --data-urlencode code_challenge_method=S256

# Copy the code query parameter from the Location header.
CODE=...

curl -sS http://localhost:8080/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d grant_type=authorization_code \
  -d client_id=go-oidc-mock-public \
  -d redirect_uri=http://localhost:3000/callback \
  -d code="$CODE" \
  -d code_verifier="$VERIFIER"
```

Build and run the binary directly:

```sh
moon run root:build
./bin/go-oidc-mock serve
```

The default issuer is `http://localhost:8080`. Set it when the server is reached
through a different host, port, or path:

```sh
GO_OIDC_MOCK_ISSUER_URL=https://auth.test.example ./bin/go-oidc-mock serve
```

## Commands

| Command | Description |
| --- | --- |
| `serve` (default) | Run the mock OIDC/OAuth server. |
| `version` | Print version, commit, and build date. |
| `openapi` | Write the OpenAPI 3.0.3 spec to stdout or a file (`--output/-o`). |

```sh
./bin/go-oidc-mock openapi -o docs/docs/openapi.yaml
./bin/go-oidc-mock version
```

## Configuration

Flags bind to Viper, so every setting is also a `GO_OIDC_MOCK_*` environment
variable. Precedence is flag > env > default.

| Flag | Env var | Default | Description |
| --- | --- | --- | --- |
| `--addr` | `GO_OIDC_MOCK_ADDR` | `:8080` | host:port the API listens on |
| `--metrics-addr` | `GO_OIDC_MOCK_METRICS_ADDR` | `:9090` | dedicated `/metrics` listener; empty serves `/metrics` on `--addr` |
| `--issuer-url` | `GO_OIDC_MOCK_ISSUER_URL` | `http://localhost:8080` | external issuer URL advertised by discovery metadata |
| `--profile-dir` | `GO_OIDC_MOCK_PROFILE_DIR` | `/etc/go-oidc-mock/profiles` | directory containing mounted JSON profile templates |
| `--log-level` | `GO_OIDC_MOCK_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `--log-format` | `GO_OIDC_MOCK_LOG_FORMAT` | `json` | `json` or `text` |
| `--read-timeout` | `GO_OIDC_MOCK_READ_TIMEOUT` | `5s` | reading an entire request |
| `--read-header-timeout` | `GO_OIDC_MOCK_READ_HEADER_TIMEOUT` | `5s` | reading request headers |
| `--write-timeout` | `GO_OIDC_MOCK_WRITE_TIMEOUT` | `10s` | writing the response |
| `--idle-timeout` | `GO_OIDC_MOCK_IDLE_TIMEOUT` | `120s` | idle keep-alive connections |
| `--request-timeout` | `GO_OIDC_MOCK_REQUEST_TIMEOUT` | `15s` | per-request processing |
| `--shutdown-grace` | `GO_OIDC_MOCK_SHUTDOWN_GRACE` | `15s` | graceful shutdown window |
| `--cors-allowed-origins` | `GO_OIDC_MOCK_CORS_ALLOWED_ORIGINS` | _(none)_ | allowed CORS origins; empty disables CORS |
| `--trusted-proxy-header` | `GO_OIDC_MOCK_TRUSTED_PROXY_HEADER` | _(none)_ | proxy header to read the client IP from; empty trusts the TCP peer |
| `--rate-limit-enabled` | `GO_OIDC_MOCK_RATE_LIMIT_ENABLED` | `true` | enable per-client rate limiting |
| `--rate-limit-rps` | `GO_OIDC_MOCK_RATE_LIMIT_RPS` | `10` | sustained per-client request rate |
| `--rate-limit-burst` | `GO_OIDC_MOCK_RATE_LIMIT_BURST` | `20` | per-client burst size |
| `--tracing-enabled` | `GO_OIDC_MOCK_TRACING_ENABLED` | `false` | enable OpenTelemetry tracing; exporter uses standard `OTEL_*` env vars |

`--issuer-url` must be an absolute `http` or `https` URL with no query or
fragment. A trailing slash is normalized away before endpoint URLs are derived.

## Profiles

At startup, the server reads non-recursive `*.json` profile files from
`--profile-dir`. Missing or empty directories are allowed and fall back to the
built-in mock user.

The default profile is the file whose `id` is `default`, or the first profile by
filename when no profile uses that ID. Valid authorization requests are
auto-approved for that profile.

```json
{
  "id": "default",
  "label": "Default user",
  "subject": "user-123",
  "claims": {
    "name": "Default User",
    "email": "user@example.test",
    "email_verified": true
  },
  "custom_claims": {
    "roles": ["tester"]
  }
}
```

## Protocol Surface

Implemented now:

- `GET /.well-known/openid-configuration`
- `GET /jwks.json`
- `GET /oauth2/authorize`
- `POST /oauth2/authorize`
- `GET /oauth2/authorize/*`
- `POST /oauth2/authorize/*`
- `POST /oauth2/token`
- `GET /userinfo`
- `POST /userinfo`

Built-in clients:

- `go-oidc-mock-public`: public authorization-code client with S256 PKCE.
- `go-oidc-mock-confidential`: confidential authorization-code client with
  secret `go-oidc-mock-secret` and `client_secret_basic`.

Both clients allow `http://localhost:3000/callback` and
`http://127.0.0.1:3000/callback`. Valid authorization requests are
auto-approved for the selected startup profile; without mounted profiles, the
fallback subject is `go-oidc-mock-user` with `name`, `email`, and
`email_verified` claims.

Operational endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /metrics` on `--metrics-addr` by default
- `GET /docs`
- `GET /openapi.yaml`
- `GET /openapi.json`

## Development

```sh
moon run root:format
moon run root:lint
moon run root:build
moon run root:test
moon run root:openapi
moon run root:check
moon run docs:build
```

Run `moon run root:openapi` after changing Huma API operations so the committed
`docs/docs/openapi.yaml` stays in sync. Standard OIDC/OAuth protocol endpoints
are described by discovery metadata and are intentionally not OpenAPI
operations.
