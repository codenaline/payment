<p align="center">
  <a href="https://github.com/codenaline/payment">
    <img width="720" alt="Payment برای Go" src="https://github.com/user-attachments/assets/5fc8ef9a-307b-4bbb-90fe-e54ba6bb0ee6">
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
  <a href="README.md">English</a> | پارسی 
</p>

# Payment

پکیج `payment` یک رابط مستقل از درگاه برای پرداخت در Go فراهم می‌کند. این پکیج یک API کوچک و مشترک برای ایجاد، تأیید و در صورت پشتیبانی بازپرداخت تراکنش‌ها دارد و تنظیمات و جزئیات پروتکل هر سرویس‌دهنده را در پکیج اختصاصی آن نگه می‌دارد.

هر کلاینت به یک درگاه متصل است. رجیستری سراسری یا درگاه پیش‌فرض قابل‌تغییری وجود ندارد؛ بنابراین انتخاب درگاه، ذخیره‌سازی، تلاش مجدد، تطبیق تراکنش‌ها و قوانین کسب‌وکار در اختیار اپلیکیشن باقی می‌ماند.

## پیش‌نیاز

- Go نسخه 1.24 یا جدیدتر

## نصب

```sh
go get github.com/codenaline/payment@latest
```

پکیج اصلی و فقط درگاه‌هایی را import کنید که اپلیکیشن به آن‌ها نیاز دارد:

```go
import (
	"github.com/codenaline/payment"
	"github.com/codenaline/payment/zarinpal"
)
```

## درگاه‌های پشتیبانی‌شده

| Provider | Purchase | Verify | Refund | Currencies | Sandbox |
| --- | :---: | :---: | :---: | --- | :---: |
| [زرین پال](https://www.zarinpal.com/) | ✅  | ✅  | ❌ | IRR | ✅  |
| [نکست پی](https://nextpay.org/) | ✅  | ✅  | ✅  | IRR, IRT | ❌ |
| [سامان کیش](https://www.sep.ir/) | ✅ | ✅ | ❌ | IRR | ❌ |

بازپرداخت یک قابلیت اختیاری است. فراخوانی `Client.Refund` برای درگاهی که `payment.Refunder` را پیاده‌سازی نکرده باشد، خطای `payment.ErrUnsupported` را برمی‌گرداند.

## شروع سریع

یک درگاه بسازید، آن را به کلاینت متصل کنید و پرداخت را ایجاد کنید:

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

قبل از هدایت مشتری به `PurchaseResponse.RedirectURL`، شناسه تراکنش، شناسه سفارش، مبلغ، ارز و وضعیت فعلی را ذخیره کنید.

فیلد `Money.Amount` یک عدد صحیح بر اساس واحد ارز انتخاب‌شده است. این پکیج ارز یا واحد پول را تبدیل نمی‌کند؛ مبلغ را با واحد مورد انتظار درگاه تنظیم‌شده ارسال کنید.

## تأیید پرداخت

بازگشت مرورگر به‌تنهایی اثبات پرداخت نیست. پس از بازگشت مشتری از درگاه، تراکنش را در یک سرور مورد اعتماد و با استفاده از شناسه تراکنش ذخیره‌شده و مبلغ اولیه تأیید کنید:

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

پردازش callback باید idempotent باشد. وضعیت تأییدشده را پیش از تحویل سفارش ذخیره کنید و در صورت دریافت دوباره callback یک پرداخت موفق، همان نتیجه قبلی را برگردانید.

کدهای `100` (تأییدشده) و `101` (قبلاً تأییدشده) زرین‌پال هر دو یک تراکنش پرداخت‌شده و بدون خطا برمی‌گردانند.

سامان کیش پس از پرداخت یک `RefNum` جدید به callback ارسال می‌کند. برای سامان کیش ابتدا فیلدهای callback را اعتبارسنجی کنید و سپس `RefNum` را به‌عنوان `VerifyRequest.TransactionID` بفرستید؛ توکن ایجاد پرداخت را برای verify ارسال نکنید. همچنین از استفاده یک `RefNum` برای بیش از یک سفارش جلوگیری کنید.

## تنظیم درگاه‌ها

### زرین‌پال

```go
gateway, err := zarinpal.New(zarinpal.Config{
	MerchantID: "your-merchant-id",
	Sandbox:    true,       // Optional; false by default.
	HTTPClient: httpClient, // Optional.
})
```

زرین‌پال ارز IRR را می‌پذیرد. ایجاد پرداخت به مبلغ مثبت، آدرس callback مطلق و توضیح غیرخالی نیاز دارد. اگر `HTTPClient` برابر nil باشد، درگاه از یک کلاینت HTTP با timeout سی‌ثانیه‌ای استفاده می‌کند.

### نکست پی

```go
gateway, err := nextpay.New(nextpay.Config{
	APIKey:     "your-api-key",
	HTTPClient: httpClient, // Optional.
})
```

نکست پی ارزهای IRR و IRT را می‌پذیرد. ایجاد پرداخت به مبلغ مثبت، `OrderID` غیرخالی و آدرس callback مطلق نیاز دارد. اگر `HTTPClient` برابر nil باشد، درگاه از یک کلاینت HTTP با timeout سی‌ثانیه‌ای استفاده می‌کند.

نکست پی در صورت وجود، این کلیدهای اختیاری `PurchaseRequest.Metadata` را ارسال می‌کند: `customer_phone`، `payer_name` و `allowed_card`.

### سامان کیش

```go
gateway, err := sep.New(sep.Config{
	TerminalID: 12345678,
	HTTPClient: httpClient, // Optional.
})
```

سامان کیش ارز IRR را می‌پذیرد. ایجاد پرداخت به مبلغ مثبت، `OrderID` غیرخالی و آدرس callback مطلق نیاز دارد. اگر `HTTPClient` برابر nil باشد، درگاه از یک کلاینت HTTP با timeout سی‌ثانیه‌ای استفاده می‌کند.

IP عمومی سرور پذیرنده باید در سامان کیش ثبت شده باشد. پس از callback موفق، `RefNum` را حداکثر ظرف ۳۰ دقیقه verify کرده و آن را با سفارش و مبلغ ذخیره‌شده تطبیق دهید. سامان کیش محیط sandbox عمومی ارائه نمی‌کند.

## استفاده از چند درگاه

برای هر درگاه تنظیم‌شده یک کلاینت بسازید و انتخاب درگاه را در اپلیکیشن انجام دهید:

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

ارسال درگاه nil به `payment.NewClient` باعث panic می‌شود. درگاه یک کلاینت پس از ساخته‌شدن قابل‌تغییر نیست.

## بازپرداخت

نکست پی قابلیت اختیاری `payment.Refunder` را پیاده‌سازی می‌کند. زرین‌پال و سامان کیش در حال حاضر از آن پشتیبانی نمی‌کنند.

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

اپلیکیشن باید نتیجه بازپرداخت را ذخیره و مطابق سیاست‌های خود تطبیق دهد.

## مدیریت خطا

پکیج اصلی خطاهای sentinel مستقل از درگاه را ارائه می‌دهد. برای تصمیم‌گیری مستقل از سرویس‌دهنده از `errors.Is` استفاده کنید:

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

پکیج هر درگاه برای عیب‌یابی خطای type‌شده ارائه می‌دهد. فقط زمانی که جزئیات اختصاصی درگاه لازم است از `errors.As` استفاده کنید:

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

خطای خام درگاه را به مشتری نمایش ندهید؛ این خطا ممکن است شامل جزئیات عملیاتی باشد.

## افزودن درگاه اختصاصی

برای افزودن یک سرویس‌دهنده بدون ثبت سراسری، `payment.Gateway` را پیاده‌سازی کنید:

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

اگر درگاه از بازپرداخت پشتیبانی می‌کند، `payment.Refunder` را نیز پیاده‌سازی کنید. درگاه اختصاصی باید خطاهای sentinel پکیج اصلی را wrap کند و در صورت نیاز به جزئیات بیشتر، نوع خطای مختص سرویس‌دهنده را ارائه دهد.

## مسئولیت‌های اپلیکیشن

این پکیج عمداً موارد زیر را مدیریت نمی‌کند:

- انتخاب یا مسیریابی درگاه
- ذخیره تراکنش، سفارش و بازپرداخت
- اعتبارسنجی callback و idempotency
- سیاست‌های تلاش مجدد، timeout و تطبیق
- ثبت log، metric یا trace
- تحویل سفارش و سایر قوانین کسب‌وکار

اطلاعات ورود را commit نکنید و شناسه پذیرنده، API key، اطلاعات مشتری یا داده کامل callback را در log و گزارش عمومی خطا قرار ندهید.

## مشارکت و پشتیبانی

از مشارکت شما استقبال می‌شود. پیش از ایجاد pull request، [CONTRIBUTING.md](CONTRIBUTING.md) را مطالعه کنید. پرسش‌های مربوط به استفاده را در [GitHub Discussions](https://github.com/codenaline/payment/discussions) مطرح کنید و آسیب‌پذیری‌ها را مطابق [SECURITY.md](SECURITY.md) به‌صورت خصوصی گزارش دهید.

## مجوز

این پروژه تحت [مجوز MIT](LICENSE) منتشر شده است.
