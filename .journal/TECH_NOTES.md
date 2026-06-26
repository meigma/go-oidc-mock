# Technical Notes

- Use hexagonal architecture at all times. Keep business logic isolated from CLI, filesystem, network, storage, and other external adapters.
- Prefer functional testing before calling any feature complete. Unit tests are useful, but they do not prove the tool works the way the design intends.
- Take an agile approach to development. Avoid waterfall: underspecify when useful, prototype early, learn from the result, and refine from working behavior.
- `go-oidc-mock` serves discovery and JWKS through `github.com/luikyv/go-oidc@v0.23.0` with an ephemeral in-memory RSA signing key; authorize, token, and userinfo intentionally remain Huma 501 placeholders until phase 2.
- The template todo, Cedar/API-key authz, PostgreSQL, migrations, sqlc, mockery, seed data, and integration-test surfaces were removed in session 001.
- Session 002 produced `.journal/002/DESIGN.md` and `.journal/002/IMPLEMENTATION_PLAN.md` for the full-service implementation direction. Start with the protocol-library spike and update the design from working evidence.
- Session 003 landed the phase 1 provider spike in PR #9. Discovery and JWKS are provider-owned exact raw routes, not Huma/OpenAPI-owned endpoints.
