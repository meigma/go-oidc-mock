# Technical Notes

- Use hexagonal architecture at all times. Keep business logic isolated from CLI, filesystem, network, storage, and other external adapters.
- Prefer functional testing before calling any feature complete. Unit tests are useful, but they do not prove the tool works the way the design intends.
- Take an agile approach to development. Avoid waterfall: underspecify when useful, prototype early, learn from the result, and refine from working behavior.
- `go-oidc-mock` is currently a static OIDC/OAuth protocol shell: discovery and empty JWKS work, while authorize, token, and userinfo intentionally return 501.
- The template todo, Cedar/API-key authz, PostgreSQL, migrations, sqlc, mockery, seed data, and integration-test surfaces were removed in session 001.
