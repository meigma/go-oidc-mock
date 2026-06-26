# Temporary Design Notes

This document is intentionally lightweight. It captures decisions and open
questions for future implementation sessions without trying to fully specify the
service before a working prototype exists.

## Protocol Library Proposal

Propose using `github.com/luikyv/go-oidc` as the first library spike.

Why this candidate:

- It is a configurable OpenID Connect Provider library, not only an OAuth2
  toolkit or client-side verifier.
- Latest checked module version on 2026-06-26: `v0.23.0`.
- It exposes `provider.New(...)` plus `Provider.Handler()`, so the protocol
  implementation can be mounted inside the existing HTTP service.
- It has default in-memory managers for grants, auth sessions, refresh tokens,
  and related state when managers are passed as `nil`.
- It supports static clients through `provider.WithStaticClients(...)`.
- It supports the flows this project needs first: authorization code, refresh
  token, discovery, JWKS, token issuance, userinfo, PKCE, and standard client
  authentication configuration.
- It lets the mock-specific behavior stay local through small callbacks:
  auth policy, ID token claims, userinfo claims, token options, and static
  client/user fixture mapping.
- Its dependency set is smaller than the other full OP candidate inspected
  (`github.com/zitadel/oidc/v3`) and it avoids the broad storage interface that
  ZITADEL requires.

Likely first integration shape:

1. Keep `internal/oidc` as the domain-facing package for fixture clients, users,
   signing keys, and issuer configuration.
2. Add a protocol adapter that builds a `go-oidc` provider from those fixtures.
3. Configure endpoints to preserve the current public shape:
   `/.well-known/openid-configuration`, `/jwks.json`, `/oauth2/authorize`,
   `/oauth2/token`, and `/userinfo`.
4. Start with static clients, static users, generated or configured RSA signing
   key, authorization code flow, refresh tokens, PKCE S256, and `client_secret`
   auth plus public-client support.
5. Implement a mock auth policy that auto-selects a configured default user for
   the first prototype. Add UI or test-control endpoints only after the protocol
   flow works end to end.

Open questions for the spike:

- Verify the `go-oidc` handler can be mounted cleanly under the existing chi
  router without interfering with health, readiness, metrics, or RFC 9457
  fallbacks.
- Decide whether protocol OpenAPI remains Huma-generated, becomes static docs,
  or is omitted for the library-owned protocol endpoints.
- Confirm endpoint override behavior for `/oauth2/authorize`, `/oauth2/token`,
  and `/jwks.json` against a running prototype.
- Confirm whether `go-oidc` can issue exactly the claim shapes we want from a
  compact fixture model without custom managers.

## Other Candidates Checked

- `github.com/zitadel/oidc/v3@v3.47.5`: solid OP/RP implementation with working
  example server and in-memory storage, but the storage/client/login interface is
  broader than needed for a simple mock. Keep as the fallback if `go-oidc`
  proves too hard to shape.
- `github.com/ory/fosite@v0.49.0`: battle-tested OAuth2/OIDC framework, but it
  is lower level. We would still own more protocol assembly than desired.
- `github.com/oauth2-proxy/mockoidc@v0.0.0-20240214162133-caebfff84d25`: close
  to the desired behavior and useful as a reference, but it is an untagged
  test-server package with fixed `/oidc/...` defaults and a narrow public model.
- `github.com/go-oauth2/oauth2/v4@v4.5.4`: easy OAuth2 server package, but not
  enough OIDC provider surface for this project.
