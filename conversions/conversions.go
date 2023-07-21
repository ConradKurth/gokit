package conversions

import (
	"context"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/shopspring/decimal"
	"golang.org/x/text/currency"
)

// CurrencyConverter converts one currency to another
type CurrencyConverter interface {
	LatestExchangeRates(ctx context.Context, from currency.Unit) (map[currency.Unit]decimal.Decimal, error)
	LatestExchangeRate(ctx context.Context, from, to currency.Unit) (decimal.Decimal, error)
	GetHistoricalRate(ctx context.Context, from, to currency.Unit, date time.Time) (decimal.Decimal, error)
}

// AmountInAndAdd will convert the from currency and then add to passed in to currency
func AmountInAndAdd(ctx context.Context, from, addTo *money.Money, code currency.Unit, converter CurrencyConverter) (*money.Money, error) {
	out, err := amountInHelper(from, code, func() (decimal.Decimal, error) {
		return converter.LatestExchangeRate(ctx, currency.MustParseISO(from.Currency().Code), code)
	})
	if err != nil {
		return nil, err
	}
	return addTo.Add(out)
}

// AmountIn coverts the Money amount to a new currency based on the ISO 4217 Currency Code.
func AmountIn(ctx context.Context, from *money.Money, code currency.Unit, converter CurrencyConverter) (*money.Money, error) {
	return amountInHelper(from, code, func() (decimal.Decimal, error) {
		return converter.LatestExchangeRate(ctx, currency.MustParseISO(from.Currency().Code), code)
	})
}

// AmountInHistoricalAndAdd will convert the from currency historically and then add to passed in to currency
func AmountInHistoricalAndAdd(ctx context.Context, from, addTo *money.Money, code currency.Unit, date time.Time, converter CurrencyConverter) (*money.Money, error) {
	out, err := amountInHelper(from, code, func() (decimal.Decimal, error) {
		return converter.GetHistoricalRate(ctx, currency.MustParseISO(from.Currency().Code), code, date)
	})
	if err != nil {
		return nil, err
	}
	return addTo.Add(out)
}

// AmountInHistorical coverts the Money amount to a new currency based on the ISO 4217 Currency Code and the historical date.
func AmountInHistorical(ctx context.Context, from *money.Money, code currency.Unit, date time.Time, converter CurrencyConverter) (*money.Money, error) {
	return amountInHelper(from, code, func() (decimal.Decimal, error) {
		return converter.GetHistoricalRate(ctx, currency.MustParseISO(from.Currency().Code), code, date)
	})
}

// AmountInHistoricalWithRates is a helper function to convert the from currency to something with specific rates instead of hitting an api
func AmountInHistoricalWithRates(ctx context.Context, from *money.Money, code currency.Unit, date time.Time, converter CurrencyConverter, rates map[currency.Unit]decimal.Decimal) (*money.Money, error) {
	if v, ok := rates[code]; ok {
		return amountInHelper(from, code, func() (decimal.Decimal, error) {
			return v, nil
		})
	}
	return AmountInHistorical(ctx, from, code, date, converter)
}

func amountInHelper(from *money.Money, code currency.Unit, f func() (decimal.Decimal, error)) (*money.Money, error) {
	if from.Currency().Code == code.String() {
		return from, nil
	}

	amt := decimal.NewFromInt(from.Amount())
	exRate, err := f()
	if err != nil {
		return nil, err
	}
	newAmt := amt.Mul(exRate).Round(2)
	return money.New(newAmt.IntPart(), code.String()), nil
}

// NewFromFloat coverts a float64 / currency code combo to the money struct
func NewFromFloat(amount float64, code currency.Unit) *money.Money {
	a := roundDecimalToInt(decimal.NewFromFloat(amount))
	return money.New(a.IntPart(), code.String())
}

// roundDecimalToInt takes in a decimal rounds it to the seconds place and multiples it by 100 to make it a whole int
func roundDecimalToInt(d decimal.Decimal) decimal.Decimal {
	return d.Round(2).Mul(decimal.NewFromInt(100))
}

func NewFromDecimal(d decimal.Decimal, code currency.Unit) *money.Money {
	d = roundDecimalToInt(d)
	return money.New(d.IntPart(), code.String())
}
