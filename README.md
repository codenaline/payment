# Payment

`payment` is a provider-neutral payment package for Go. It defines a small set
of payment contracts and keeps provider-specific configuration and protocol
details in separate packages.

The package does not choose gateways, route payments, or manage user
preferences. An application can configure multiple gateways at the same time
by creating one immutable `payment.Client` for each gateway.

## Requirements

- Go 1.24 or newer

## Installation

```sh
go get github.com/codenaline/payment
```

Provider packages are imported separately. Zarinpal is currently included:

```go
import "github.com/codenaline/payment/zarinpal"
```

## Quick start

Create the provider first so configuration errors can be handled normally:

```go
package main

import (
	"context"
	"fmt"

	"github.com/codenaline/payment"
	"github.com/codenaline/payment/zarinpal"
)

func main() {
	gateway, err := zarinpal.New("your-merchant-id")
	if err != nil {
		panic(err)
	}

	client := payment.NewClient(gateway)
	result, err := client.Purchase(context.Background(), payment.PurchaseRequest{
		Amount: payment.Money{
			Amount:   100_000,
			Currency: "IRR",
		},
		CallbackURL: "https://example.com/payments/callback",
		Description: "Order #1234",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(result.RedirectURL)
}
```

`Money.Amount` is expressed in the currency's smallest unit. For example,
USD 10.50 is represented as `Money{Amount: 1050, Currency: "USD"}`.

## Verifying a payment

Persist the transaction ID returned by `Purchase`. After the provider redirects
the payer to the callback URL, use that ID and the original amount to verify the
payment:

```go
transaction, err := client.Verify(ctx, payment.VerifyRequest{
	TransactionID: transactionID,
	Amount: payment.Money{
		Amount:   100_000,
		Currency: "IRR",
	},
})
if err != nil {
	return err
}

if transaction.Status == payment.StatusPaid {
	// Fulfill the order.
}
```

Verification should happen on a trusted server. Do not treat the browser
redirect alone as proof of payment.

## Multiple gateways

Create an independent client for every configured gateway. The application
owns gateway selection and routing:

```go
zarinpalGateway, err := zarinpal.New("zarinpal-merchant-id")
if err != nil {
	return err
}

zarinpalClient := payment.NewClient(zarinpalGateway)
nextpayClient := payment.NewClient(nextpayGateway)

// Choose the appropriate client using application-specific business rules.
_ = zarinpalClient
_ = nextpayClient
```

A client never changes its gateway after construction. This makes multiple
clients safe to use independently and keeps routing rules outside the package.

## Optional capabilities

Every gateway supports purchasing and verification. Refunds and transaction
lookup are optional capabilities:

```go
refund, err := client.Refund(ctx, payment.RefundRequest{
	TransactionID: transactionID,
	Amount:        amount,
	Reason:        "customer request",
})
if errors.Is(err, payment.ErrUnsupported) {
	// The configured gateway does not support refunds.
}
```

`FindTransaction` follows the same convention and returns
`payment.ErrUnsupported` when the gateway does not implement
`payment.TransactionFinder`.

## Error handling

The root package exposes portable sentinel errors. Use `errors.Is` for logic
that should work with every provider:

```go
switch {
case errors.Is(err, payment.ErrInvalidRequest):
	// Correct the request.
case errors.Is(err, payment.ErrNetwork):
	// Retry according to the application's policy.
case errors.Is(err, payment.ErrDeclined):
	// Ask the customer to use another payment method.
case errors.Is(err, payment.ErrTransactionNotFound):
	// Reconcile the stored transaction information.
}
```

Providers expose their own error types for provider-specific diagnostics. Use
`errors.As` only when the application needs those details:

```go
var providerError *zarinpal.Error
if errors.As(err, &providerError) {
	fmt.Printf("Zarinpal operation %s failed with code %d: %s\n",
		providerError.Operation,
		providerError.Code,
		providerError.Message,
	)
}
```

## Zarinpal options

Use the sandbox during development:

```go
gateway, err := zarinpal.New(
	"your-merchant-id",
	zarinpal.WithSandbox(),
)
```

An application can also supply its own `http.Client`:

```go
gateway, err := zarinpal.New(
	"your-merchant-id",
	zarinpal.WithHTTPClient(httpClient),
)
```

The default client has a 30-second timeout.

## Custom gateways

Applications and third-party packages can provide their own gateways without
registering them globally. Implement `payment.Gateway` and pass the value to
`payment.NewClient`:

```go
type CustomGateway struct{}

func (*CustomGateway) Purchase(
	ctx context.Context,
	request payment.PurchaseRequest,
) (payment.PurchaseResponse, error) {
	// Call the provider.
	return payment.PurchaseResponse{}, nil
}

func (*CustomGateway) Verify(
	ctx context.Context,
	request payment.VerifyRequest,
) (payment.Transaction, error) {
	// Call the provider.
	return payment.Transaction{}, nil
}

client := payment.NewClient(&CustomGateway{})
```

Implement `payment.Refunder` or `payment.TransactionFinder` when the provider
supports those optional operations.

## Design principles

- Provider-neutral core contracts
- Immutable, single-gateway clients
- Application-owned routing and business rules
- Explicit provider configuration
- Standard `errors.Is` and `errors.As` integration
- No global registry or mutable gateway selection

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
