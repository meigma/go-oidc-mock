---
id: 002
title: Temporary Design Document
started: 2026-06-26
---

## 2026-06-26 13:08 — Kickoff
Goal for the session: begin work on a temporary design document in the local session journal folder that will let future sessions implement the full `go-oidc-mock` service.
Current state of the world: session 001 completed the rebrand from `template-go-api` and left a static OIDC/OAuth protocol shell with discovery and empty JWKS working, while authorize, token, and userinfo remain explicit 501 placeholders.
Plan: prime this session first, then continue with an intentionally agile design document that captures enough direction for future implementation without over-specifying the service upfront.

## 2026-06-26 13:18 — Protocol library research
Researched Go OIDC/OAuth provider candidates against the current service shape. Proposed `github.com/luikyv/go-oidc` as the first spike target because it is a configurable OP library with `Provider.Handler()`, default in-memory managers, static clients, endpoint overrides, and callbacks for mock-specific claims and auth behavior.
Captured the comparison in `.journal/002/DESIGN.md`: ZITADEL `oidc` is the fallback, Fosite is too low-level for the current goal, `mockoidc` is useful as a behavior reference but not a clean service dependency, and `go-oauth2/oauth2` is not enough OIDC provider surface.

## 2026-06-26 13:37 — JIT user model
Captured the user's preferred user model: the OAuth authorization/consent page should let testers select a profile and edit the effective user/claims just-in-time, rather than requiring a mock restart or fixture refresh for each identity edge case.
Updated `.journal/002/DESIGN.md` to treat profiles as templates and to snapshot the approved user/claim data onto the grant so ID token, userinfo, and refresh-token behavior remain stable for already-issued authorizations.

## 2026-06-26 13:42 — Docker Compose profile templates
Captured the runtime assumption that `go-oidc-mock` will primarily run under Docker Compose. Profile templates should be JSON data files mounted into a well-known container directory, while ordinary service settings stay on flags/env vars.
Updated `.journal/002/DESIGN.md` to keep template loading simple: startup-load mounted `*.json` profile files first, prove the path and schema with a spike, and defer hot reload or save-as-profile behavior until the JIT flow works.

## 2026-06-26 13:47 — Resolved open questions
Resolved the open design questions with the user's agreement: combine login and consent; use light JSON plus reserved-claim validation; defer save-as-profile; use one profile per JSON file; mount protocol handlers on exact chi paths; drop OpenAPI; preserve current OIDC endpoint paths; try grant-local JIT snapshots before custom managers; and default profile loading to a Compose-mounted directory with a flag/env override.
Updated `.journal/002/DESIGN.md` to replace the open-question lists with resolved decisions and a short set of remaining spike checks.

## 2026-06-26 13:49 — Formal design draft
Transformed `.journal/002/DESIGN.md` from rough notes into a structured engineering design document. The new shape covers status, summary, current state, goals, non-goals, design principles, proposed architecture, HTTP surface, protocol library choice, JIT user/profile model, runtime configuration, implementation slices, acceptance criteria, alternatives, risks, and deferred work.
Kept the profile JSON schema illustrative rather than final so the next spike can prove the shape without locking in unnecessary detail.

## 2026-06-26 13:56 — Phased implementation plan
Added `.journal/002/IMPLEMENTATION_PLAN.md` as a lightweight companion to the design document. The plan sequences future work through protocol-library spike, minimal auth-code flow, mounted profiles, combined authorization page, grant-local snapshots, refresh behavior, Compose/operator polish, and a consolidation checkpoint.
Kept each phase focused on a brief outcome and success criteria so implementation details can be decided from working code rather than over-specified upfront.

## 2026-06-26 14:23 — Close
Closed the session as a journal-only design session. No implementation PRs were opened or merged; the main `master` worktree was clean, aligned with `origin/master`, and had no tracked `.journal` files.
Handoff artifacts are `.journal/002/DESIGN.md`, `.journal/002/IMPLEMENTATION_PLAN.md`, and `.journal/002/SUMMARY.md`. Future implementation work should start with the protocol-library spike and revise the design from working evidence.
