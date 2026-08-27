# Contributing Guide

Thank you for considering a contribution to `payment`.

## Reporting issues

- Use the bug report form for reproducible defects.
- Use the feature request form for API proposals and provider support.
- Use [GitHub Discussions](https://github.com/codenaline/payment/discussions) for usage questions.
- Report vulnerabilities privately as described in [SECURITY.md](SECURITY.md).

Include the package version or commit, Go version, provider, sandbox or production environment, and a minimal reproduction. Remove merchant IDs, API keys, authorities, customer details, and other sensitive payment data.

## Submitting pull requests

1. Fork the repository and create a focused branch.
2. Make one coherent change with clear commit messages.
3. Add tests for behavioral changes and regressions.
4. Update exported comments, examples, and the changelog when appropriate.
5. Open a pull request and link related issues.

Keep gateway routing, persistence, retries, user preferences, and business rules outside the core package. Provider-specific protocol details and errors belong in the provider package.

## Development requirements

- Go 1.24 or newer.
- No live provider credentials are required; tests must use local HTTP test servers.

Run these checks before submitting:

```sh
gofmt -w *.go nextpay/*.go zarinpal/*.go
go vet ./...
go test ./...
go test -race ./...
```

## Code style

- Follow standard Go conventions and existing package patterns.
- Keep the root package provider-neutral.
- Wrap portable errors so callers can use `errors.Is`.
- Preserve provider details through provider error types and `errors.As`.
- Accept `context.Context` for network operations.
- Reuse configured HTTP clients and always close response bodies.
- Document every exported declaration.
- Avoid global registries and mutable gateway switching.

## Public API changes

Describe compatibility impact and migration steps for exported API changes. New optional operations should use capability interfaces rather than expanding the required `Gateway` interface.
