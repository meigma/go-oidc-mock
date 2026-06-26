---
id: 001
title: Rebrand From Template
date: 2026-06-26
status: complete
repos_touched: [go-oidc-mock]
related_sessions: []
---

## Goal
Rebrand the repository from `template-go-api` into `go-oidc-mock`, a mock OIDC/OAuth server intended for test environments. Keep the first pass agile: establish the project identity and protocol-shaped shell, while removing template features that do not serve that purpose.

## Outcome
The goal was met. PR #8 was squash-merged into `master` as `519d073`, local `master` was fast-forwarded, and the implementation worktree and branch were removed.

## Key Decisions
- Keep HTTP service infrastructure and release automation -> these are useful for a mock server even before full OIDC flows exist.
- Remove todo, authz, API-key, PostgreSQL, sqlc, goose, mockery, and integration-test surfaces -> they were template-specific and would add noise to the first OIDC pass.
- Add OIDC discovery, JWKS, and explicit 501 flow endpoints -> this gives the project the right protocol shape without pretending login/token/userinfo are implemented.
- Keep state static and in memory for this pass -> fixtures, users, sessions, token signing, and persistence can be designed from working protocol needs later.

## Changes
- `go.mod`, `cmd/go-oidc-mock`, and local imports - renamed the module, command, and binary surface to `go-oidc-mock`.
- `internal/config` and `internal/cli` - simplified runtime config to service/HTTP concerns plus `--issuer-url` / `GO_OIDC_MOCK_ISSUER_URL`.
- `internal/oidc` and `internal/oidc/httpapi` - added provider metadata, empty JWKS, and 501 placeholders for authorize, token, and userinfo routes.
- `internal/app` and `internal/adapter/http` - wired the OIDC API into the existing Huma/router/health/metrics stack.
- `internal/todo`, `internal/authz`, `internal/adapter/postgres`, `internal/integration`, `hack/sql`, `sqlc.yaml`, and related Moon/Proto files - removed template todo, authz, database, migration, sqlc, mockery, and integration-test surfaces.
- `README.md`, `docs/docs`, `docs/mkdocs.yml`, `CHANGELOG.md`, `CONTRIBUTING.md`, and `SECURITY.md` - rewrote documentation around the OIDC mock purpose and first-pass limitations.
- `.github`, `.goreleaser.yaml`, `ghd.toml`, `Dockerfile`, `compose.yaml`, `moon.yml`, and release metadata - updated names, package/image labels, env prefixes, smoke commands, and repository template settings.

## Open Threads
- Successful login, authorization-code, token, refresh, revocation, introspection, and userinfo flows are intentionally not implemented yet.
- No clients, users, sessions, fixtures, persistence, or token signing keys exist yet.
- JWKS is intentionally empty until signing support is introduced.
- License selection remains an owner decision before publishing.

## References
- PR: https://github.com/meigma/go-oidc-mock/pull/8
- Merge commit: `519d073 feat: rebrand as OIDC mock server`
