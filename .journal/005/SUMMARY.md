---
id: 005
title: Continue Implementation
date: 2026-06-26
status: complete
repos_touched: [go-oidc-mock]
related_sessions: [002, 004]
---

## Goal
Continue implementation from the merged minimal authorization-code flow and land
the next small working slice of the mock OIDC service.

## Outcome
The goal was met. PR #11 was squash-merged into `master` as `873b672`, local
`master` was fast-forwarded to the merge commit, and the implementation
worktree plus feature branches were removed.

## Key Decisions
- Use mounted profiles immediately in auto-approval -> this made profile files
  observable through ID token and `/userinfo` behavior without waiting for an
  authorization page.
- Keep profile domain validation in `internal/oidc` and filesystem loading in
  `internal/adapter/profilefile` -> this preserves the repo's hexagonal
  boundary between business rules and external adapters.
- Treat missing or empty profile directories as non-fatal -> existing users can
  keep the built-in fallback mock user without mounting files.
- Prefer `id=default` and otherwise first filename order -> this gives Compose
  users a deterministic default profile before profile-selection UI exists.

## Changes
- `internal/oidc/profile.go`, `internal/oidc/service.go` - added the profile
  model, validation, default selection, fallback profile, and profile-backed
  auto-approval grant snapshots.
- `internal/adapter/profilefile/loader.go` - added non-recursive sorted JSON
  profile loading from a configured directory with strict JSON decoding.
- `internal/config/config.go`, `internal/app/app.go` - added
  `--profile-dir` / `GO_OIDC_MOCK_PROFILE_DIR` and wired loaded profiles into
  provider construction.
- `internal/adapter/profilefile/*_test.go`, `internal/oidc/*_test.go`,
  `internal/app/flow_test.go`, `internal/config/config_test.go` - added loader,
  profile, config, and functional OIDC flow coverage for mounted claims in ID
  tokens and `/userinfo`.
- `examples/profiles/default.json`, `compose.yaml`, `README.md`,
  `docs/docs/index.md` - added sample mounted profile data and documented the
  profile directory behavior.

## Open Threads
- The authorization page, explicit profile selection UI, JIT claim editing,
  refresh-token behavior, dynamic clients, admin APIs, and persistent signing
  keys remain deferred.
- Session 002's temporary design artifacts should still be refreshed from the
  implemented phase 2 and phase 3 evidence before treating them as reference
  architecture.

## References
- PR #11: https://github.com/meigma/go-oidc-mock/pull/11
- Merge commit: `873b6722f9161b07205a35a32ca9b71799fab798`
- Feature commit before squash: `4d8f175bf2372d3c016745cb0caa862f162daf4b`
- Prior session: `.journal/004/SUMMARY.md`
