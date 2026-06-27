# Technical Notes

- Use hexagonal architecture at all times. Keep business logic isolated from CLI, filesystem, network, storage, and other external adapters.
- Prefer functional testing before calling any feature complete. Unit tests are useful, but they do not prove the tool works the way the design intends.
- Take an agile approach to development. Avoid waterfall: underspecify when useful, prototype early, learn from the result, and refine from working behavior.
- `go-oidc-mock` serves discovery, JWKS, authorize, token, and userinfo through `github.com/luikyv/go-oidc@v0.23.0` with an ephemeral in-memory RSA signing key.
- Phase 2's built-in clients are `go-oidc-mock-public` and `go-oidc-mock-confidential` with secret `go-oidc-mock-secret`; the built-in fallback user is `go-oidc-mock-user` with `name=Mock User`, `email=user@example.test`, and `email_verified=true`.
- Phase 3's profile loader reads non-recursive `*.json` files from `--profile-dir` / `GO_OIDC_MOCK_PROFILE_DIR` (default `/etc/go-oidc-mock/profiles`), prefers profile ID `default`, otherwise uses first filename order, and falls back to the built-in mock user when the directory is missing or empty.
- The template todo, Cedar/API-key authz, PostgreSQL, migrations, sqlc, mockery, seed data, and integration-test surfaces were removed in session 001.
- Session 002 produced `.journal/002/DESIGN.md` and `.journal/002/IMPLEMENTATION_PLAN.md` for the full-service implementation direction. Start with the protocol-library spike and update the design from working evidence.
- Session 003 landed the phase 1 provider spike in PR #9. Discovery and JWKS are provider-owned exact raw routes, not Huma/OpenAPI-owned endpoints.
- Session 004 landed the minimal authorization-code flow in PR #10. Protocol endpoints are provider-owned raw routes, not OpenAPI operations.
- Session 005 landed mounted profile templates in PR #11. Auto-approval snapshots the selected startup profile into grant-local ID token and userinfo claims; no authorization page or profile-selection UI exists yet.
- Session 006 landed the combined authorization page in PR #12. `/oauth2/authorize` now renders a server-side page, approval/denial resumes through `/oauth2/authorize/{sessionID}`, edited profile claims are snapshotted into the existing grant store for ID token and userinfo callbacks, and refresh-token snapshot behavior is still deferred.
