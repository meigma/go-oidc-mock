# Full OIDC Mock Service Implementation Plan

## Status

- Status: lightweight plan for future implementation sessions
- Last updated: 2026-06-26
- Companion design: [DESIGN.md](DESIGN.md)

This plan is intentionally phase-based rather than task-complete architecture.
Each phase should prove a useful slice of behavior, then let the next session
adjust based on what was learned.

## Phase 1: Protocol Library Spike

Prove that `github.com/luikyv/go-oidc` can be embedded in the existing service
without fighting the current router, lifecycle, or endpoint paths.

Success criteria:

- The service starts with `go-oidc` wired in behind the existing HTTP stack.
- Discovery and JWKS are served on the current public paths.
- JWKS returns at least one signing key.
- Health, readiness, metrics, logging, and fallbacks still behave as before.
- A clear fallback decision is recorded if endpoint mounting or provider setup
  is not clean enough.

## Phase 2: Minimal Authorization Code Flow

Add the smallest static-client flow that lets a test application complete an
authorization-code exchange.

Success criteria:

- At least one static client can be configured at startup.
- `/oauth2/authorize` can issue an authorization code.
- `/oauth2/token` can exchange that code for tokens.
- PKCE S256 works for public clients.
- Client-secret authentication works for confidential clients.
- Existing endpoint paths remain stable in discovery metadata and behavior.

## Phase 3: Mounted Profile Templates

Load reusable user profiles from Docker Compose-friendly JSON files mounted into
a well-known directory.

Success criteria:

- The service reads one profile per JSON file at startup.
- The default profile directory is `/etc/go-oidc-mock/profiles`.
- `--profile-dir` and `GO_OIDC_MOCK_PROFILE_DIR` can override the directory.
- Startup validation catches malformed profile files with useful errors.
- The first schema stays small and can still evolve after real usage.

## Phase 4: Combined Authorization Page

Replace placeholder authorization behavior with a tester-facing page that
combines consent, profile selection, and just-in-time user editing.

Success criteria:

- The page shows enough client and scope context for a tester to approve or
  deny the request.
- A mounted profile can prefill the editable user fields.
- Testers can edit the subject, common standard claims, and custom claims JSON.
- Approval and denial both return protocol-correct responses.
- The page remains a mock testing tool, not a general identity-management UI.

## Phase 5: Grant-Local User Snapshot

Persist the approved user and claims on the authorization grant so issued tokens
and userinfo responses remain stable for the life of that grant.

Success criteria:

- Approval records the effective subject, standard claims, custom claims,
  selected profile ID, and approved scopes.
- ID token claims are populated from the snapshot.
- `/userinfo` responses are populated from the same snapshot.
- Custom claims cannot override protocol-owned reserved claims.
- The implementation avoids custom managers unless the spike shows they are
  necessary.

## Phase 6: Refresh Token Behavior

Verify that refresh-token flows continue to represent the originally approved
user snapshot rather than current profile file contents or later edits.

Success criteria:

- Refresh-token grants are enabled for configured clients.
- Refreshed tokens preserve the original subject and claims.
- Changing a profile file after startup does not affect an existing grant.
- Any library limitations around refresh state are documented before expanding
  the design.

## Phase 7: Compose And Operator Polish

Make the proven flow easy to run from Docker Compose without adding admin APIs
or persistent infrastructure.

Success criteria:

- Compose examples mount profile files and set required runtime env vars.
- README instructions describe the minimal working flow.
- Token lifetime and issuer settings are configurable through flags/env vars.
- Startup failures for profile, client, issuer, or signing-key problems are
  clear enough to fix without reading source code.
- No OpenAPI or admin endpoint surface is introduced for v1.

## Phase 8: Consolidation Checkpoint

Stop after the working vertical slice and update the design from implementation
evidence before adding broader features.

Success criteria:

- The implementation can complete an end-to-end OIDC authorization-code flow
  using a mounted profile and edited JIT claims.
- The design document reflects any library or routing lessons learned.
- Deferred items remain deferred unless the working slice proves they are
  necessary.
- The next implementation backlog is based on observed gaps, not speculative
  completeness.
