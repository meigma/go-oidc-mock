---
id: 002
title: Temporary Design Document
date: 2026-06-26
status: complete
repos_touched: [go-oidc-mock]
related_sessions: [001]
---

## Goal
Create temporary local-session design artifacts that future sessions can use to
implement the full `go-oidc-mock` service without over-specifying the design
before a working spike exists.

## Outcome
The goal was met. The session produced a structured design document and a
lightweight phased implementation plan in the personal journal. No
implementation PR was opened or merged because this session only changed journal
artifacts.

## Key Decisions
- Use `github.com/luikyv/go-oidc` as the first protocol spike dependency -> it
  is an embeddable Go OpenID Provider library with provider handlers, static
  clients, in-memory managers, endpoint overrides, and callback points for
  mock-specific behavior.
- Keep `github.com/zitadel/oidc/v3` as the fallback -> it is a capable OP/RP
  implementation, but its storage/client/login surface looks broader than
  needed for this mock.
- Make the authorization page the v1 user surface -> testers can select a
  profile and edit the effective user/claims just in time, avoiding service
  restarts for identity variations.
- Treat profiles as templates, not live grant state -> approval snapshots the
  effective user and claims onto the grant so ID token, userinfo, and refresh
  behavior stay stable.
- Load profiles as mounted JSON files -> Docker Compose users can mount one
  profile per file into a well-known directory while normal service settings
  remain flags/env vars.
- Drop OpenAPI as a v1 goal -> the service is not expected to expose admin
  endpoints, and the protocol endpoints are already described by OIDC/OAuth
  discovery and standards.

## Changes
- `.journal/002/DESIGN.md` - captures the full-service design direction,
  including goals, non-goals, architecture shape, HTTP surface, protocol library
  choice, JIT user/profile model, validation, runtime configuration, risks, and
  deferred work.
- `.journal/002/IMPLEMENTATION_PLAN.md` - captures the lightweight phased plan:
  protocol library spike, minimal authorization-code flow, mounted profiles,
  combined authorization page, grant-local user snapshots, refresh behavior,
  Compose/operator polish, and consolidation checkpoint.
- `.journal/002/NOTES.md` - records the running session decisions and the final
  closeout handoff.
- `.journal/TECH_NOTES.md` - now points future agents at the design and phased
  implementation plan.
- `.journal/INDEX.md` - marks session 002 complete.

## Open Threads
- No service implementation has been started from these artifacts yet.
- The first implementation session should prove `go-oidc` endpoint mounting,
  JWKS/signing, static clients, and grant-local JIT snapshot behavior before
  expanding the design.
- The profile JSON schema is intentionally illustrative and should be refined
  from the spike.
- Hot reload, save-as-profile, admin/control APIs, dynamic clients, additional
  grant types, and persistent storage remain deferred.

## References
- Design artifact: `.journal/002/DESIGN.md`
- Phased plan artifact: `.journal/002/IMPLEMENTATION_PLAN.md`
- Prior session: `.journal/001/SUMMARY.md`
- Key journal commits:
  - `9d94e9f docs(journal): record protocol library proposal`
  - `bc3a03c docs(journal): capture jit user model`
  - `ffe7e08 docs(journal): capture compose profile loading`
  - `4bf91cb docs(journal): resolve service design questions`
  - `9c497ba docs(journal): formalize service design draft`
  - `eeae4ea docs(journal): add phased implementation plan`
