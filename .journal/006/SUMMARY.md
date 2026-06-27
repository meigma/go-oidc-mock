---
id: 006
title: Combined Authorization Page
date: 2026-06-26
status: complete
repos_touched: [go-oidc-mock]
related_sessions: [002, 005]
---

## Goal
Continue implementation from the mounted-profile baseline and land phase 4 of the original plan: a combined authorization page for profile selection, just-in-time claim editing, approval, and denial.

## Outcome
The goal was met. PR #12 was reviewed, squash-merged into `master` as `48b2bf3`, local `master` was fast-forwarded, and the phase 4 implementation worktree plus feature branch were removed.

## Key Decisions
- Use `go-oidc` `StatusPending` auth sessions -> this keeps `/oauth2/authorize` provider-owned and resumes through the existing `/oauth2/authorize/{sessionID}` callback route.
- Keep the first UI server-rendered with inline CSS -> enough for tester workflows without adding a frontend build pipeline or new public API surface.
- Preserve phase 3 profile defaults -> mounted `default` still wins, otherwise the first mounted profile wins, while the built-in mock user remains available as a fallback option.
- Snapshot approved edits through the existing grant store keys -> ID token and `/userinfo` callbacks continue to read from one grant-local source of truth.

## Changes
- `internal/oidc/authorization_page.go` - added the interactive authorization policy, HTML page, approve/deny handling, edited profile validation, and grant snapshot wiring.
- `internal/oidc/service.go` - replaced unconditional auto-approval with the new authorization-page policy.
- `internal/app/flow_test.go` - updated functional OIDC flows to drive page rendering and callback posts, with coverage for mounted profiles, edited claims, denial redirects, validation errors, and first-mounted-profile defaults.

## Open Threads
- Refresh-token behavior remains deferred; a later phase still needs to verify refreshed tokens preserve the original approved grant snapshot.
- Dynamic clients, admin APIs, persistent signing keys, profile save-back, and hot reload remain intentionally out of scope.
- Session 002's draft design should eventually be refreshed from the implemented phase 2 through phase 4 evidence.

## References
- PR #12: https://github.com/meigma/go-oidc-mock/pull/12
- Merge commit: `48b2bf3ac5f5796bd145e315e436e1e8cdbf6138`
- Prior implementation baseline: `.journal/005/SUMMARY.md`
- Original implementation direction: `.journal/002/IMPLEMENTATION_PLAN.md`

## Lessons
- In a Worktrunk checkout, `gh pr merge --delete-branch` can complete the GitHub merge but fail local cleanup when it tries to switch to a default branch that is already checked out elsewhere. Verify the PR state before retrying, then fast-forward the main checkout and clean up with `wt remove`.
