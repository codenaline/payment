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

# Payment

`payment` is a small, provider-neutral payment package for Go. It defines common contracts for creating, verifying, and optionally refunding payments while keeping provider configuration, protocols, and errors in provider packages.

The package deliberately does not manage gateway selection, routing, persistence, retries, reconciliation, or business rules. Applications can configure several gateways by creating one immutable client for each gateway.

## Requirements

- Go 1.24 or newer

## Installation

Install the module and import only the providers your application uses:

```sh
go get github.com/codenaline/payment@latest
```

```go
import (
	"github.com/codenaline/payment"
	"github.com/codenaline/payment/zarinpal"
)
```

## Supported drivers

| Provider | Purchase | Verify | Refund | Currencies | Sandbox |
| --- | :---: | :---: | :---: | --- | :---: |
| [ZarinPal](zarinpal) | Yes | Yes | No | IRR | Yes |
| [NextPay](nextpay) | Yes | Yes | Yes | IRR, IRT | No |

Refund support is exposed as an optional capability. A custom gateway can implement the core `payment.Gateway` interface and, when applicable, `payment.Refunder`.

## Upcoming drivers

The following drivers are candidates for future releases:

- Zibal
- IDPay
- Pay.ir
- Stripe

This list is not a delivery commitment and has no fixed order. If you need one
of these drivers—or another provider—you are welcome to implement it and
[create a pull request](https://github.com/codenaline/payment/compare). Read the
[contribution guide](CONTRIBUTING.md) first, and open a
[provider integration proposal](https://github.com/codenaline/payment/issues/new?template=feature_request.yml)
when the API or expected behavior needs discussion.

## Quick start

Create a gateway, wrap it in a client, and initiate a payment:

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

	// Persist result.Transaction.ID with the order before redirecting the payer.
	fmt.Println(result.RedirectURL)
}
```

Redirect the payer to `PurchaseResponse.RedirectURL`. Persist the transaction ID, order ID, amount, currency, and status in your own database.

`Money.Amount` is an integer in the selected currency unit. The package does not convert between currencies or units; use the unit required by the configured provider.

## Verify a payment

After the provider redirects the payer to your callback endpoint, verify the transaction on a trusted server using the stored transaction ID and original amount:

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

A browser redirect alone is not proof of payment. Callback handlers should be idempotent: persist the verified state and return the existing successful result when a paid callback is repeated.

ZarinPal verification codes `100` (verified now) and `101` (already verified) both produce a paid transaction without an error.

## Multiple gateways

Create an independent client for each configured gateway. The application decides which client to use:

```go
zarinpalGateway, err := zarinpal.New(zarinpal.Config{
	MerchantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
})
if err != nil {
	return err
}

nextpayGateway, err := nextpay.New(nextpay.Config{
	APIKey: "nextpay-api-key",
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

A client does not switch or mutate its gateway after construction. There is no global gateway registry.

## Provider configuration

### ZarinPal

```go
gateway, err := zarinpal.New(zarinpal.Config{
	MerchantID: "your-merchant-id",
	Sandbox:    true,       // Optional; defaults to false.
	HTTPClient: httpClient, // Optional.
})
```

ZarinPal accepts IRR. When `HTTPClient` is nil, the driver uses its default HTTP client with a 30-second timeout.

### NextPay

```go
gateway, err := nextpay.New(nextpay.Config{
	APIKey:     "your-api-key",
	HTTPClient: httpClient, // Optional.
})
```

NextPay accepts IRR and IRT and requires `PurchaseRequest.OrderID` when creating a payment.

## Refunds

Refund is an optional gateway capability. `Client.Refund` returns `payment.ErrUnsupported` when the configured gateway does not implement it:

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

## Error handling

The root package exposes portable sentinel errors. Use `errors.Is` for provider-independent decisions:

```go
switch {
case errors.Is(err, payment.ErrInvalidRequest):
	// Correct the request; retrying it unchanged will not help.
case errors.Is(err, payment.ErrNetwork):
	// Apply the application's retry and reconciliation policy.
case errors.Is(err, payment.ErrDeclined):
	// Ask the customer to use another payment method.
case errors.Is(err, payment.ErrTransactionNotFound):
	// Reconcile the stored transaction information.
case errors.Is(err, payment.ErrCanceled):
	// Record that the payment was canceled.
case errors.Is(err, payment.ErrProvider):
	// Handle an unclassified provider failure.
}
```

Provider packages expose their own error types. Use `errors.As` only when provider-specific diagnostics are needed:

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

Do not expose provider errors directly to customers when they may contain operational details.

## Custom gateways

Implement `payment.Gateway` to integrate another provider without registering it globally:

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

Implement `payment.Refunder` if the provider supports refunds. Custom gateways should wrap the portable sentinel errors and expose a provider-specific error type when callers need additional details.

## Project scope

The package provides:

- Common payment types and gateway interfaces
- Immutable clients bound to one gateway
- Bundled ZarinPal and NextPay drivers
- Portable and provider-specific error handling
- Optional gateway capabilities such as refunds

Applications remain responsible for:

- Gateway selection and routing
- Transaction and order persistence
- Callback validation and idempotency
- Retry, timeout, and reconciliation policies
- Logging, metrics, and tracing
- Fulfillment and all other business rules

## Contributing

Pull requests are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Credits

- [Mahdi Rezaei](https://github.com/mahdirezaei-dev)

## License

The MIT License. See [LICENSE](LICENSE) for details.