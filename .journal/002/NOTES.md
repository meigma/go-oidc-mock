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
