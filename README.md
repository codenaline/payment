<p align="center">
  <a href="https://github.com/codenaline/payment">
    <img width="720" alt="Payment for Go" src="https://github.com/user-attachments/assets/5fc8ef9a-307b-4bbb-90fe-e54ba6bb0ee6">
  </a>
</p>

<p align="center">
  <a href="https://github.com/codenaline/payment/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/codenaline/payment/actions/workflows/ci.yml/badge.svg?branch=main"></a>
  <a href="https://codecov.io/gh/codenaline/payment"><img alt="Coverage" src="https://codecov.io/gh/codenaline/payment/graph/badge.svg"></a>
  <a href="https://pkg.go.dev/github.com/codenaline/payment"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/codenaline/payment.svg"></a>
  <a href="https://go.dev/"><img alt="Go 1.24 or newer" src="https://img.shields.io/badge/go-%3E%3D1.24-00ADD8?logo=go&logoColor=white"></a>
  <a href="LICENSE"><img alt="MIT License" src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
</p>

<p align="center">
  English | <a href="README.fa.md">پارسی</a>
</p>

# Payment

`payment` is a provider-neutral payment package for Go. It provides a small common API for purchasing, verifying, and optionally refunding payments while keeping provider configuration and protocol details in dedicated packages.

Each client is bound to one gateway. There is no global registry or mutable default, so applications remain in control of provider selection, persistence, retries, reconciliation, and business rules.

## Requirements

- Go 1.24 or newer

## Installation

```sh
go get github.com/codenaline/payment@latest
```

Import the root package and only the provider packages your application uses:

```go
import (
	"github.com/codenaline/payment"
	"github.com/codenaline/payment/zarinpal"
)
```

## Supported providers

| Provider | Purchase | Verify | Refund | Currencies | Sandbox |
| --- | :---: | :---: | :---: | --- | :---: |
| [ZarinPal](https://www.zarinpal.com/) | ✅  | ✅  | ❌ | IRR | ✅  |
| [NextPay](https://nextpay.org/) | ✅  | ✅  | ✅  | IRR, IRT | ❌ |

Refunding is an optional capability. Calling `Client.Refund` with a gateway that does not implement `payment.Refunder` returns `payment.ErrUnsupported`.

## Quick start

Create a gateway, bind it to a client, and initiate a payment:

```go
package main

import (
	"context"
	"fmt"

	"github.com/codenaline/payment"
	"github.com/codenaline/payment/zarinpal"
)

func main() {
	gateway, err := zarinpal.New(zarinpal.Config{
		MerchantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Sandbox:    true,
	})
	if err != nil {
		panic(err)
	}

	client := payment.NewClient(gateway)
	result, err := client.Purchase(context.Background(), payment.PurchaseRequest{
		OrderID: "order-1234",
		Amount: payment.Money{
			Amount:   100_000,
			Currency: payment.CurrencyIRR,
		},
		CallbackURL: "https://example.com/payments/callback",
		Description: "Order #1234",
	})
	if err != nil {
		panic(err)
	}

	// Store the transaction before redirecting the customer.
	fmt.Println(result.Transaction.ID)
	fmt.Println(result.RedirectURL)
}
```

Persist the transaction ID, order ID, amount, currency, and current status before redirecting the customer to `PurchaseResponse.RedirectURL`.

`Money.Amount` is an integer expressed in the selected currency unit. The package does not convert currencies or units; pass the unit expected by the configured provider.

## Verify a payment

The browser callback is not proof of payment. After the provider redirects the customer, verify the transaction from a trusted server using the stored transaction ID and original amount:

```go
transaction, err := client.Verify(ctx, payment.VerifyRequest{
	TransactionID: storedTransactionID,
	Amount: payment.Money{
		Amount:   100_000,
		Currency: payment.CurrencyIRR,
	},
})
if err != nil {
	return err
}

if transaction.Status == payment.StatusPaid {
	// Persist the paid state before fulfilling the order.
}
```

Make callback processing idempotent. Store the verified state before fulfillment and return the existing successful result when the same paid callback is received again.

ZarinPal response codes `100` (verified) and `101` (already verified) both return a paid transaction without an error.

## Provider configuration

### ZarinPal

```go
gateway, err := zarinpal.New(zarinpal.Config{
	MerchantID: "your-merchant-id",
	Sandbox:    true,       // Optional; false by default.
	HTTPClient: httpClient, // Optional.
})
```

ZarinPal accepts IRR. Purchases require a positive amount, an absolute callback URL, and a non-empty description. If `HTTPClient` is nil, the provider uses an HTTP client with a 30-second timeout.

### NextPay

```go
gateway, err := nextpay.New(nextpay.Config{
	APIKey:     "your-api-key",
	HTTPClient: httpClient, // Optional.
})
```

NextPay accepts IRR and IRT. Purchases require a positive amount, a non-empty `OrderID`, and an absolute callback URL. If `HTTPClient` is nil, the provider uses an HTTP client with a 30-second timeout.

NextPay forwards these optional `PurchaseRequest.Metadata` keys when present: `customer_phone`, `payer_name`, and `allowed_card`.

## Multiple gateways

Create one client for each configured gateway and keep provider selection in the application:

```go
zarinpalGateway, err := zarinpal.New(zarinpal.Config{
	MerchantID: "your-merchant-id",
})
if err != nil {
	return err
}

nextpayGateway, err := nextpay.New(nextpay.Config{
	APIKey: "your-api-key",
})
if err != nil {
	return err
}

zarinpalClient := payment.NewClient(zarinpalGateway)
nextpayClient := payment.NewClient(nextpayGateway)

// Select a client using application-owned business rules.
_ = zarinpalClient
_ = nextpayClient
```

`payment.NewClient` panics when passed a nil gateway. A client cannot switch gateways after construction.

## Refunds

NextPay implements the optional `payment.Refunder` capability. ZarinPal currently does not.

```go
refund, err := client.Refund(ctx, payment.RefundRequest{
	TransactionID: transactionID,
	Amount:        amount,
	Reason:        "customer request",
})
if errors.Is(err, payment.ErrUnsupported) {
	// Use a provider-specific or manual refund process.
}
```

Applications should persist the refund result and reconcile it according to their own policies.

## Error handling

The root package exposes portable sentinel errors. Use `errors.Is` for provider-independent decisions:

```go
switch {
case errors.Is(err, payment.ErrInvalidRequest):
	// Fix the request; retrying it unchanged will not help.
case errors.Is(err, payment.ErrNetwork):
	// Apply the application's retry and reconciliation policy.
case errors.Is(err, payment.ErrDeclined):
	// Ask the customer to use another payment method.
case errors.Is(err, payment.ErrTransactionNotFound):
	// Reconcile the stored transaction details.
case errors.Is(err, payment.ErrCanceled):
	// Record the canceled payment.
case errors.Is(err, payment.ErrProvider):
	// Handle an unclassified provider failure.
}
```

Provider packages expose typed errors for diagnostics. Use `errors.As` only when provider-specific details are needed:

```go
var providerError *zarinpal.Error
if errors.As(err, &providerError) {
	fmt.Printf("ZarinPal %s failed with code %d: %s\n",
		providerError.Operation,
		providerError.Code,
		providerError.Message,
	)
}
```

Avoid showing raw provider errors to customers because they may contain operational details.

## Custom gateways

Implement `payment.Gateway` to add a provider without global registration:

```go
type CustomGateway struct{}

func (*CustomGateway) Purchase(
	ctx context.Context,
	request payment.PurchaseRequest,
) (payment.PurchaseResponse, error) {
	return payment.PurchaseResponse{}, nil
}

func (*CustomGateway) Verify(
	ctx context.Context,
	request payment.VerifyRequest,
) (payment.Transaction, error) {
	return payment.Transaction{}, nil
}

client := payment.NewClient(&CustomGateway{})
```

Implement `payment.Refunder` when the provider supports refunds. Custom providers should wrap the root sentinel errors and expose a provider-specific error type when callers need more detail.

## Application responsibilities

This package intentionally does not manage:

- Gateway selection or routing
- Transaction, order, and refund persistence
- Callback authentication and idempotency
- Retry, timeout, and reconciliation policies
- Logging, metrics, or tracing
- Fulfillment and other business rules

Never commit credentials or include merchant IDs, API keys, customer information, or complete callback data in logs and public bug reports.

## Contributing and support

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request. Use [GitHub Discussions](https://github.com/codenaline/payment/discussions) for usage questions and follow [SECURITY.md](SECURITY.md) to report vulnerabilities privately.

## License

Released under the [MIT License](LICENSE).
