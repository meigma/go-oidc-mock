# Full OIDC Mock Service Design

## Status

- Status: draft for implementation spike
- Last updated: 2026-06-26
- Location: local session journal
- Scope: future implementation sessions for `go-oidc-mock`

This document captures the current implementation direction. It is intentionally
not a full architecture manual. The next useful step is a small working spike
that proves the protocol library, routing, mounted profile files, and JIT user
snapshot behavior.

## Summary

`go-oidc-mock` should become a Docker Compose-friendly mock OIDC/OAuth provider
for test environments. The service should use an existing OpenID Provider
library for protocol mechanics, load profile templates from mounted JSON files,
and let testers edit the effective user and claims just-in-time on the
authorization page.

The preferred protocol dependency is `github.com/luikyv/go-oidc`. The first
implementation should preserve the current public endpoint paths, drop OpenAPI
as a project goal, and keep ordinary service configuration on flags/env vars.

## Current State

Session 001 turned the repository from a generic API template into an OIDC mock
service shell. Discovery and an empty JWKS endpoint exist. Authorization, token,
and userinfo endpoints currently return explicit `501 Not Implemented`
responses.

The current implementation is useful as an HTTP service scaffold, but it does
not yet execute an OAuth/OIDC flow.

## Goals

- Implement a usable authorization code flow for test clients.
- Support refresh tokens, ID tokens, access tokens, userinfo, JWKS, and provider
  discovery.
- Reduce custom protocol code by delegating OAuth/OIDC mechanics to a proven Go
  OpenID Provider library.
- Preserve the public endpoint paths already advertised by the shell.
- Let testers create or adjust users at authorization time without restarting
  the service.
- Support reusable user templates through mounted JSON files.
- Keep configuration simple for Docker Compose users.

## Non-Goals

- Do not build a production identity provider.
- Do not add persistent database storage in v1.
- Do not add admin APIs in v1.
- Do not preserve or expand OpenAPI generation for protocol endpoints.
- Do not build profile save/edit management in v1.
- Do not implement every OAuth/OIDC grant type in the first pass.
- Do not implement hot reload for mounted profile templates until real use shows
  it is worth the complexity.

## Design Principles

- Prefer a working protocol spike over speculative architecture.
- Let the protocol library own standards-heavy behavior.
- Keep mock-specific behavior small, explicit, and local.
- Treat profile files as test data, not service configuration.
- Snapshot user/claim data onto the grant so issued tokens remain stable.
- Keep startup configuration in flags/env vars.

## Proposed Architecture

The service should keep the existing CLI, config, chi router, observability, and
Docker packaging. The OIDC implementation should move behind a protocol adapter
that constructs and mounts a `go-oidc` provider.

Primary pieces:

- CLI/config: reads issuer URL, bind addresses, token settings, and profile
  directory.
- Profile loader: reads one JSON profile per file from a configured directory.
- Protocol adapter: builds a `go-oidc` provider from issuer, clients, signing
  keys, profile templates, and auth policy.
- Combined authorization page: shows requested client/scope context, lets the
  tester select a profile, edit effective claims, and approve or deny.
- Grant snapshot: stores the approved user and claims for later token/userinfo
  callbacks.
- Existing router: owns health, readiness, metrics, logging, recovery, and exact
  protocol path mounting.

## HTTP Surface

Preserve these public paths:

| Path | Purpose |
| --- | --- |
| `/.well-known/openid-configuration` | OIDC discovery |
| `/jwks.json` | JSON Web Key Set |
| `/oauth2/authorize` | authorization endpoint and JIT user page |
| `/oauth2/token` | token endpoint |
| `/userinfo` | userinfo endpoint |

The protocol handler should be mounted on exact paths through the existing chi
router. It should not take over `/`, `/healthz`, `/readyz`, `/metrics`, or router
fallback behavior.

OpenAPI should be dropped as a goal for this service. Standard OIDC/OAuth
endpoints are already described by the protocol and discovery metadata, and v1
does not need admin endpoints.

## Protocol Library

Use `github.com/luikyv/go-oidc` as the first spike dependency.

Reasons:

- It is a Go OpenID Provider implementation, not only an OAuth2 toolkit or OIDC
  client verifier.
- It exposes `provider.New(...)` and `Provider.Handler()` for embedding in an
  existing HTTP service.
- It supports static clients.
- It has default in-memory managers for grants, auth sessions, refresh tokens,
  and related state.
- It exposes endpoint override options.
- It lets mock behavior live in callbacks such as auth policy, ID token claims,
  userinfo claims, token options, and grant handling.
- Its dependency set and integration surface look smaller than
  `github.com/zitadel/oidc/v3` for this use case.

Expected first configuration:

- Authorization code flow.
- Refresh token grant.
- PKCE S256.
- Static clients.
- Public-client support plus `client_secret` authentication.
- RS256 signing.
- Endpoint overrides for current public paths.
- Grant-local JIT user snapshot read by ID token and userinfo callbacks.

If the spike shows `go-oidc` cannot support the required path or grant-snapshot
behavior cleanly, fall back to `github.com/zitadel/oidc/v3`.

## User And Profile Model

The authorization page is the primary user creation surface for v1. Profiles are
templates, not the source of truth for issued grants.

First-pass behavior:

1. A tester starts the normal authorization flow from their application.
2. The authorization page shows the client, requested scopes, and editable user
   fields.
3. The tester selects a profile to prefill fields.
4. The tester edits standard fields and custom claims as needed.
5. Approval writes the effective user snapshot onto the authorization grant.
6. ID token and userinfo callbacks read claims from the grant snapshot.
7. Refresh token flows continue to use the original grant snapshot.

This avoids requiring a service restart for every user variation while keeping
already-issued grants stable.

## Profile Templates

Profiles should be loaded from JSON files mounted into the container. Use one
profile per file.

Default runtime path:

```text
/etc/go-oidc-mock/profiles/*.json
```

Configuration override:

```text
--profile-dir
GO_OIDC_MOCK_PROFILE_DIR
```

The exact schema should be proven in the spike. A reasonable initial shape is:

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

The service should startup-load templates in v1. Hot reload can be deferred.
Saving edited users back to profiles is also deferred.

## JIT User Snapshot

On approval, the effective user should be copied into grant-local data. The
snapshot should include:

- selected profile ID, if any
- subject
- standard claims
- custom claims
- approved scopes

Illustrative shape:

```json
{
  "profile_id": "default",
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

The exact storage key and shape inside `go-oidc` grant data should be decided by
the spike.

## Validation

Keep validation focused on test ergonomics and protocol safety.

Apply these rules in v1:

- `sub` must be present and non-empty after profile selection and user edits.
- Standard editable fields should have obvious type checks, such as
  `email_verified` being boolean.
- Custom claims must be a JSON object.
- Custom claims must not override protocol-owned or reserved claims.

Reserved claims should include at least:

```text
iss aud exp iat nbf jti nonce azp at_hash c_hash
```

Avoid building a broad schema language in v1.

## Runtime Configuration

Expected Docker Compose model:

- The service image contains the binary and default startup behavior.
- Compose mounts profile JSON files into the profile directory.
- Flags/env vars configure service behavior.
- Profile templates are test data, not env-var payloads.

Continue to use flags/env vars for:

- issuer URL
- HTTP bind address
- metrics bind address
- profile directory
- logging settings
- token lifetimes
- rate limits and timeouts

Do not add admin APIs for v1.

## Implementation Slices

1. Add `go-oidc` dependency and build the smallest provider that serves
   discovery and JWKS on the existing paths.
2. Mount library-owned protocol paths through the existing chi router without
   breaking health, readiness, metrics, or fallbacks.
3. Add static client configuration sufficient for an end-to-end authorization
   code flow.
4. Add profile directory loading for one JSON profile file.
5. Add the combined authorization page with profile selection and editable user
   claims.
6. Store the approved user snapshot in the grant path.
7. Populate ID token and userinfo claims from the grant snapshot.
8. Verify refresh token behavior keeps using the original snapshot.
9. Update README and Compose examples once the spike proves the shape.

## Spike Acceptance Criteria

The next implementation spike is successful if:

- `/.well-known/openid-configuration` advertises the preserved endpoint paths.
- `/jwks.json` returns a non-empty key set.
- A client can complete authorization code flow against `/oauth2/authorize` and
  `/oauth2/token`.
- The authorization page can load at least one mounted profile JSON file.
- Edited JIT claims appear in the ID token and userinfo response.
- Refresh token flow continues to use the original approved user snapshot.
- Existing health, readiness, metrics, logging, and router fallbacks still work.

## Alternatives Considered

- `github.com/zitadel/oidc/v3@v3.47.5`: solid OP/RP implementation with a
  working example server and in-memory storage, but the storage/client/login
  interface is broader than needed for a simple mock. Keep as the fallback.
- `github.com/ory/fosite@v0.49.0`: battle-tested OAuth2/OIDC framework, but too
  low-level for the current goal. It leaves more protocol assembly in this
  service.
- `github.com/oauth2-proxy/mockoidc@v0.0.0-20240214162133-caebfff84d25`: close
  to the desired behavior and useful as a reference, but it is an untagged test
  server package with fixed `/oidc/...` defaults and a narrow public model.
- `github.com/go-oauth2/oauth2/v4@v4.5.4`: easy OAuth2 server package, but not
  enough OIDC provider surface.

## Risks

- `go-oidc` endpoint overrides or handler mounting may not fit the existing
  router cleanly. Mitigation: prove exact-path mounting first; fall back to
  ZITADEL if needed.
- Grant-local snapshots may not flow naturally into ID token, userinfo, and
  refresh behavior. Mitigation: try `grant.Store` and callbacks first; introduce
  custom managers only if necessary.
- The authorization page can grow into a full identity-management UI.
  Mitigation: keep v1 to profile selection, direct field edits, custom claims
  JSON, approve, and deny.
- Profile template schema could get over-specified before use. Mitigation:
  start with one illustrative JSON shape and refine from a working spike.

## Deferred Work

- Save edited JIT users as profiles.
- Hot reload mounted profiles.
- Admin or control APIs.
- Dynamic client registration.
- Additional grant types beyond v1 needs.
- Persistent storage.
