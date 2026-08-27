# Security Policy

## Supported versions

Until the first stable release, security fixes are applied to the latest `v0.x` release and the `main` branch. After `v1.0.0`, only the latest release line is guaranteed to receive security updates.

Use a supported Go toolchain and the latest patch release of this package.

## Reporting a vulnerability

Do not open a public issue for a suspected vulnerability.

Use [GitHub private vulnerability reporting](https://github.com/codenaline/payment/security/advisories/new) or email `mahdirezaei.dev@gmail.com`.

Include:

- A clear description of the issue and its impact.
- The affected `payment` version or commit.
- The affected provider and operation.
- The Go version and operating system.
- Minimal reproduction steps or a proof of concept.
- Any suggested mitigation.

Do not include live merchant IDs, API keys, authorities, customer information, or other payment data unless the maintainers explicitly request a secure transfer. Please avoid public disclosure until a fix or mitigation is available.
