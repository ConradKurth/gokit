package converter

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"golang.org/x/text/currency"
)

// MockCurrencyConverter implements a mock interface of the currency converter for testing.
type MockCurrencyConverter struct {
	mock.Mock
}

func (m *MockCurrencyConverter) LatestExchangeRate(ctx context.Context, from, to currency.Unit) (decimal.Decimal, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockCurrencyConverter) LatestExchangeRates(ctx context.Context, from currency.Unit) (map[currency.Unit]decimal.Decimal, error) {
	args := m.Called(ctx, from)
	return args.Get(0).(map[currency.Unit]decimal.Decimal), args.Error(1)
}

func (m *MockCurrencyConverter) GetHistoricalRate(ctx context.Context, from, to currency.Unit, date time.Time) (decimal.Decimal, error) {
	args := m.Called(ctx, from, to, date)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}
