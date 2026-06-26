---
id: 003
title: Phase 1 Provider Spike
date: 2026-06-26
status: complete
repos_touched: [go-oidc-mock]
related_sessions: [002]
---

## Goal
Review the prior session output, turn the phase 1 plan into a small implementation spike, and land it after review.

## Outcome
The goal was met. PR #9 was squash-merged after review, local `master` was fast-forwarded to the merge commit, and the feature worktree and remote branch were cleaned up.

## Key Decisions
- Embedded `github.com/luikyv/go-oidc@v0.23.0` only for discovery and JWKS -> this proves the provider can sit behind the existing HTTP stack without taking on the full authorization flow yet.
- Kept `/oauth2/authorize`, `/oauth2/token`, and `/userinfo` as Huma `501` endpoints -> phase 1 stays a protocol-library fit check, with auth-code behavior deferred to phase 2.
- Generated one ephemeral in-memory RSA signing key at startup -> enough to prove a non-empty JWKS without committing to persistent key management.
- Mounted provider-owned discovery and JWKS as exact raw routes -> current paths and middleware behavior stay intact while removing those documents from the Huma/OpenAPI surface.

## Changes
- `go.mod`, `go.sum` - added `github.com/luikyv/go-oidc@v0.23.0`.
- `internal/oidc/service.go`, `internal/oidc/types.go` - built the provider adapter, issuer normalization, ephemeral RSA key generation, and provider handler exposure.
- `internal/adapter/http/router.go`, `internal/adapter/http/api.go`, `internal/adapter/http/ratelimit.go` - added exact raw route support and kept protocol routes inside middleware and rate-limit handling.
- `internal/app/app.go`, `internal/app/app_test.go` - mounted provider-backed discovery and JWKS while preserving explicit 501 placeholders for authorize, token, and userinfo.
- `internal/oidc/httpapi/handler.go`, `internal/oidc/httpapi/api_test.go`, `internal/oidc/service_test.go` - adjusted tests around provider-owned discovery/JWKS and non-empty signing keys.
- `docs/docs/openapi.yaml`, `internal/cli/openapi_test.go` - removed discovery/JWKS from the generated OpenAPI contract.

## Open Threads
- Phase 2 still needs the real authorization-code flow, static clients, PKCE behavior, token issuance, refresh-token decisions, userinfo claims, and grant-local JIT user snapshots.
- Signing keys are still ephemeral spike keys; persistent or configured key material remains deferred.
- Keep revising session 002's design from working evidence instead of treating it as fixed architecture.

## References
- PR #9: https://github.com/meigma/go-oidc-mock/pull/9
- Merge commit: `164cfb7ff31cddc8d24cfe6b2a62dd30cf4ddbbf`
- Prior implementation direction: `.journal/002/DESIGN.md`, `.journal/002/IMPLEMENTATION_PLAN.md`
