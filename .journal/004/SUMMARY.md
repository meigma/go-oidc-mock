---
id: 004
title: Continue Implementation
date: 2026-06-26
status: complete
repos_touched: [go-oidc-mock]
related_sessions: [003]
---

## Goal
Continue implementation from the phase 1 provider spike and land the next small working slice of the mock OIDC service.

## Outcome
The goal was met. PR #10 was squash-merged into `master` as `cf8c5a5`, local `master` was fast-forwarded to the merge commit, and the implementation worktree plus feature branch were removed.

## Key Decisions
- Move `/oauth2/authorize`, `/oauth2/token`, and `/userinfo` to provider-owned raw routes -> standard OIDC/OAuth protocol endpoints should be exercised through `go-oidc`, not Huma placeholder operations.
- Keep phase 2 auto-approved and static -> built-in clients and one fixed mock user prove the authorization-code flow without adding profile files, editable UI, or runtime client registries yet.
- Keep `oidc.NewService(issuerURL)` as the default constructor with internal options -> production startup stays simple while tests can inject custom static clients.
- Use functional app-level tests as the main proof -> the valuable behavior is end-to-end redirect, token, code reuse, and userinfo behavior through the real router and provider handlers.

## Changes
- `internal/oidc/service.go` - configured `go-oidc` for authorization-code flow, S256 PKCE, static built-in clients, client auth methods, and auto-approval for `go-oidc-mock-user`.
- `internal/app/app.go` and `internal/adapter/http/*` - mounted provider-owned raw protocol routes for authorize, token, userinfo, discovery, JWKS, and authorize callback wildcard paths.
- `internal/app/flow_test.go` and `internal/app/app_test.go` - added functional coverage for public PKCE, missing PKCE failure, confidential Basic auth, code reuse rejection, userinfo, callback wildcard routing, and rate-limit behavior.
- `internal/oidc/httpapi/*` - removed the Huma `501` protocol placeholder package and tests.
- `internal/cli/openapi_test.go` and `docs/docs/openapi.yaml` - updated OpenAPI expectations so protocol endpoints are not documented as Huma operations.
- `README.md` and `docs/docs/*` - documented the built-in clients, mock user defaults, provider-owned protocol endpoints, and PKCE smoke flow.

## Open Threads
- Profile directories, editable authorization pages, dynamic or configurable clients, refresh-token behavior, admin APIs, and persistent signing key configuration remain deferred.
- Session 002's temporary design artifacts should be revised from this phase 2 evidence before later phases treat them as reference material.
- The release dry-run/container CI jobs were skipped for this feature branch; the non-skipped hosted checks passed.

## References
- PR #10: https://github.com/meigma/go-oidc-mock/pull/10
- Merge commit: `cf8c5a575cdcf12bed751a2078d2ab18ae0d398b`
- Prior session: `.journal/003/SUMMARY.md`
