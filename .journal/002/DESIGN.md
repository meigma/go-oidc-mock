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
4. Start with static clients, mounted JSON user profiles/templates, generated or
   configured RSA signing key, authorization code flow, refresh tokens, PKCE
   S256, and `client_secret` auth plus public-client support.
5. Implement a mock auth policy whose authorization/consent page lets testers
   select a profile and edit the effective user/claims just-in-time. Keep the
   first version small: prove the JIT user snapshot can flow into ID tokens and
   userinfo before adding richer UI.

## JIT Users And Profiles

Preferred user model: treat the OAuth authorization/consent page as the place
where developers and testers create the effective user for the current login.

This should reduce test setup friction. Testers should not need to restart the
mock server or edit fixture files for every identity edge case. Profiles remain
useful, but they are templates for fast selection rather than the only way to
define users.

First-pass behavior:

1. The consent page shows a profile selector plus editable user fields.
2. Selecting a profile populates defaults such as `sub`, `name`, `email`,
   `email_verified`, groups/roles, and custom claims.
3. The tester can edit the selected values before approving the authorization
   request.
4. Approval writes a snapshot of the effective user and claim data onto the
   authorization grant.
5. ID token and userinfo callbacks read from that grant snapshot, not from the
   mutable profile definition.

Why snapshot at grant time:

- Refresh token flows should continue to return claims for the user that was
  approved originally.
- Userinfo should match the access token's authorization context.
- Editing profiles later should not retroactively mutate already-issued grants.
- The mock stays flexible without making global fixture state surprising.

Likely data shape:

- Profiles: named JSON templates loaded from a mounted runtime directory, later
  editable through a control API if that becomes useful.
- JIT user snapshot: grant-local subject plus standard claims and arbitrary
  custom claims.
- Claim editor: initially a few structured fields plus a raw JSON object for
  custom claims. Avoid over-designing a full identity-management UI.

Runtime loading model:

- Expect the service to be run mostly through Docker Compose.
- Treat user profiles/templates as data files, not env-var payloads. Compose can
  mount JSON files into a well-known directory inside the container.
- Keep ordinary operational settings on flags/env vars: issuer URL, bind
  address, metrics address, logging, timeouts, token lifetimes, and the profile
  template directory if the default needs to move.
- Do not require a service rebuild for profile/template changes. The first pass
  can load templates at startup; hot reload can come later if real use shows it
  matters.
- Use a conservative default such as one directory containing `*.json` profile
  files. Exact path and schema should be proven in the spike rather than
  specified in detail now.

## Resolved Design Decisions

- Combine login and consent into one authorization page for v1. The page should
  show requested client/scope context, profile selection, editable effective
  user fields, custom claims JSON, and approve/deny controls.
- Validate custom claims lightly: require syntactically valid JSON objects and
  protect protocol-owned/reserved claims from being overridden by user-provided
  custom claim data.
- Defer saving edited JIT users back as named profiles. It is useful future
  functionality, but not needed for v1.
- Use one profile per JSON file in the mounted profile directory.
- Mount the `go-oidc` protocol handler through the existing chi router on exact
  protocol paths so health, readiness, metrics, and router fallbacks remain
  owned by the existing service infrastructure.
- Drop OpenAPI as a project goal for the protocol service. The library-owned
  standard OIDC/OAuth endpoints plus discovery metadata are enough; ordinary
  startup configuration should happen through flags/env vars, not admin APIs.
- Preserve the existing public endpoint paths:
  `/.well-known/openid-configuration`, `/jwks.json`, `/oauth2/authorize`,
  `/oauth2/token`, and `/userinfo`.
- Use `go-oidc`'s normal grant flow first: store the JIT user snapshot in the
  grant path and read it from ID token/userinfo callbacks. Avoid custom managers
  unless the spike proves they are necessary.
- Default profile loading to a Docker Compose-friendly mounted directory, likely
  `/etc/go-oidc-mock/profiles/*.json`, with a flag/env var override such as
  `--profile-dir` / `GO_OIDC_MOCK_PROFILE_DIR`.

Remaining spike checks:

- Prove endpoint override and exact-path mounting with a tiny running prototype.
- Prove a mounted one-profile JSON file can populate the combined authorization
  page.
- Prove the approved JIT user snapshot survives into ID token, userinfo, and
  refresh-token behavior without custom managers.

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
