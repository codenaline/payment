# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.0] - 2026-08-28

### Added

- Provider-neutral gateway, client, transaction, money, and status contracts.
- Portable sentinel errors compatible with `errors.Is`.
- Provider-specific errors compatible with `errors.As`.
- Zarinpal purchase and verification support, including sandbox mode.
- NextPay purchase, verification, and refund support.
- Immutable clients for applications using multiple gateways concurrently.
- Open-ended `Currency` type with `CurrencyIRR` and `CurrencyIRT` constants.
- Custom HTTP client configuration for bundled providers.

### Security

- Bounded response-body handling and strict HTTP response validation.

[Unreleased]: https://github.com/codenaline/payment/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/codenaline/payment/releases/tag/v0.1.0
