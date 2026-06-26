# Contributing

Thank you for your interest in contributing to `go-oidc-mock`.
For private vulnerability reporting, use [SECURITY.md](SECURITY.md) instead of
public channels.

## Reporting Bugs

Report non-security bugs through GitHub issues. Include the following details
when possible:

- version, commit, or environment details
- steps to reproduce
- expected behavior
- actual behavior
- logs, screenshots, or a minimal reproduction

## Pull Requests

Contributors should:

1. Keep changes focused and scoped to a single problem.
2. Add or update tests when behavior changes.
3. Update documentation when user-facing behavior changes.
4. Use Conventional Commit subjects, such as `feat: add token endpoint`.
5. Make sure `moon run root:check` passes before requesting review.

## Local Setup

```sh
moon run root:check
```

Useful project commands:

```sh
moon run root:format
moon run root:lint
moon run root:build
moon run root:test
moon run root:openapi
moon run docs:build

docker compose up --build
curl -sS localhost:8080/.well-known/openid-configuration
```

## Release Changes

Release Please reads Conventional Commit subjects to build changelogs and
release PRs. Keep release-impacting commits clear; routine docs, CI, and
maintenance commits should use the appropriate non-release type.
