package conversions_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ConradKurth/gokit/conversions"
	"github.com/ConradKurth/gokit/converter"
	"github.com/Rhymond/go-money"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/currency"
)

func Test_AmountInAndAdd(t *testing.T) {
	tt := []struct {
		Name         string
		MockExchange func() *converter.MockCurrencyConverter
		From         *money.Money
		AddTo        *money.Money
		Code         currency.Unit
		Expected     *money.Money
		ErrExpected  bool
	}{
		{
			Name: "Exchange from USD to GBP",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("LatestExchangeRate", context.Background(), currency.USD, currency.GBP).Once().Return(decimal.NewFromFloat(0.5), nil)
				return m
			},
			From:     money.New(100, "USD"),
			AddTo:    money.New(200, "GBP"),
			Code:     currency.GBP,
			Expected: money.New(250, "GBP"),
		},
		{
			Name: "Error in the exchange",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("LatestExchangeRate", context.Background(), currency.USD, currency.GBP).Once().Return(decimal.NewFromFloat(0.5), errors.New("bad error"))
				return m
			},
			From:        money.New(100, "USD"),
			AddTo:       money.New(200, "GBP"),
			Code:        currency.GBP,
			ErrExpected: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := conversions.AmountInAndAdd(context.Background(), tc.From, tc.AddTo, tc.Code, tc.MockExchange())
			if tc.ErrExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, actual, tc.Expected)
		})
	}
}

func Test_AmountInHistoricalAndAdd(t *testing.T) {
	d := time.Now().UTC()
	tt := []struct {
		Name         string
		MockExchange func() *converter.MockCurrencyConverter
		From         *money.Money
		AddTo        *money.Money
		Code         currency.Unit
		Expected     *money.Money
		ErrExpected  bool
	}{
		{
			Name: "Exchange from USD to GBP",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("GetHistoricalRate", context.Background(), currency.USD, currency.GBP, d).Once().Return(decimal.NewFromFloat(0.5), nil)
				return m
			},
			From:     money.New(100, "USD"),
			AddTo:    money.New(200, "GBP"),
			Code:     currency.GBP,
			Expected: money.New(250, "GBP"),
		},
		{
			Name: "Error in the exchange",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("GetHistoricalRate", context.Background(), currency.USD, currency.GBP, d).Once().Return(decimal.NewFromFloat(0.5), errors.New("bad error"))
				return m
			},
			From:        money.New(100, "USD"),
			AddTo:       money.New(200, "GBP"),
			Code:        currency.GBP,
			ErrExpected: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := conversions.AmountInHistoricalAndAdd(context.Background(), tc.From, tc.AddTo, tc.Code, d, tc.MockExchange())
			if tc.ErrExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, actual, tc.Expected)
		})
	}
}

func Test_AmountInHistorical(t *testing.T) {
	d := time.Now().UTC()
	tt := []struct {
		Name         string
		MockExchange func() *converter.MockCurrencyConverter
		From         *money.Money
		Code         currency.Unit
		Expected     *money.Money
		ErrExpected  bool
	}{
		{
			Name: "Exchange from USD to GBP",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("GetHistoricalRate", context.Background(), currency.USD, currency.GBP, d).Once().Return(decimal.NewFromFloat(0.5), nil)
				return m
			},
			From:     money.New(100, "USD"),
			Code:     currency.GBP,
			Expected: money.New(50, "GBP"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := conversions.AmountInHistorical(context.Background(), tc.From, tc.Code, d, tc.MockExchange())
			if tc.ErrExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, actual, tc.Expected)
		})
	}
}

func Test_AmountInHistoricalDefault(t *testing.T) {
	d := time.Now().UTC()
	tt := []struct {
		Name         string
		Rates        map[currency.Unit]decimal.Decimal
		MockExchange func() *converter.MockCurrencyConverter
		From         *money.Money
		Code         currency.Unit
		Expected     *money.Money
		ErrExpected  bool
	}{
		{
			Name: "Exchange from USD to GBP",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("GetHistoricalRate", context.Background(), currency.USD, currency.GBP, d).Once().Return(decimal.NewFromFloat(0.5), nil)
				return m
			},
			From:     money.New(100, "USD"),
			Code:     currency.GBP,
			Expected: money.New(50, "GBP"),
		},
		{
			Name:         "Exchange from USD to GBP with default rate",
			Rates:        map[currency.Unit]decimal.Decimal{currency.GBP: decimal.NewFromFloat(3)},
			MockExchange: func() *converter.MockCurrencyConverter { return &converter.MockCurrencyConverter{} },
			From:         money.New(100, "USD"),
			Code:         currency.GBP,
			Expected:     money.New(300, "GBP"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := conversions.AmountInHistoricalWithRates(context.Background(), tc.From, tc.Code, d, tc.MockExchange(), tc.Rates)
			if tc.ErrExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, actual, tc.Expected)
		})
	}
}

func Test_AmountIn(t *testing.T) {
	tt := []struct {
		Name         string
		MockExchange func() *converter.MockCurrencyConverter
		From         *money.Money
		Code         currency.Unit
		Expected     *money.Money
		ErrExpected  bool
	}{
		{
			Name: "Exchange from USD to GBP",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("LatestExchangeRate", context.Background(), currency.USD, currency.GBP).Once().Return(decimal.NewFromFloat(0.5), nil)
				return m
			},
			From:     money.New(100, "USD"),
			Code:     currency.GBP,
			Expected: money.New(50, "GBP"),
		},
		{
			Name: "Exchange from 0 USD to GBP",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("LatestExchangeRate", context.Background(), currency.USD, currency.GBP).Once().Return(decimal.NewFromFloat(0.5), nil)
				return m
			},
			From:     money.New(0, "USD"),
			Code:     currency.GBP,
			Expected: money.New(0, "GBP"),
		},
		{
			Name:         "Exchange from 0 USD to USD",
			MockExchange: func() *converter.MockCurrencyConverter { return &converter.MockCurrencyConverter{} },
			From:         money.New(0, "USD"),
			Code:         currency.USD,
			Expected:     money.New(0, "USD"),
		},
		{
			Name: "Exchange from cents USD to GBP",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("LatestExchangeRate", context.Background(), currency.USD, currency.GBP).Once().Return(decimal.NewFromFloat(0.5), nil)
				return m
			},
			From:     money.New(10542, "USD"),
			Code:     currency.GBP,
			Expected: money.New(5271, "GBP"),
		},
		{
			Name: "Exchange from negative USD to GBP",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("LatestExchangeRate", context.Background(), currency.USD, currency.GBP).Once().Return(decimal.NewFromFloat(0.5), nil)
				return m
			},
			From:     money.New(-100, "USD"),
			Code:     currency.GBP,
			Expected: money.New(-50, "GBP"),
		},
		{
			Name: "Error in exchange",
			MockExchange: func() *converter.MockCurrencyConverter {
				m := &converter.MockCurrencyConverter{}
				m.On("LatestExchangeRate", context.Background(), currency.USD, currency.GBP).Once().Return(decimal.NewFromFloat(0.5), errors.New("Bad error"))
				return m
			},
			ErrExpected: true,
			From:        money.New(-100, "USD"),
			Code:        currency.GBP,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := conversions.AmountIn(context.Background(), tc.From, tc.Code, tc.MockExchange())
			if tc.ErrExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, actual, tc.Expected)
		})
	}
}

func TestNewFromFloat(t *testing.T) {
	tt := []struct {
		Name     string
		Amount   float64
		Code     currency.Unit
		Expected *money.Money
	}{
		{
			Name:     "From positive",
			Amount:   100.101,
			Code:     currency.GBP,
			Expected: money.New(10010, "GBP"),
		},
		{
			Name:     "From negative",
			Amount:   -100.101,
			Code:     currency.GBP,
			Expected: money.New(-10010, "GBP"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			actual := conversions.NewFromFloat(tc.Amount, tc.Code)
			assert.Equal(t, actual, tc.Expected)
		})
	}
}

func TestNewFromDecimal(t *testing.T) {
	tt := []struct {
		Name     string
		Amount   decimal.Decimal
		Code     currency.Unit
		Expected *money.Money
	}{
		{
			Name:     "From positive",
			Amount:   decimal.NewFromFloat(100.101),
			Code:     currency.GBP,
			Expected: money.New(10010, "GBP"),
		},
		{
			Name:     "From negative",
			Amount:   decimal.NewFromFloat(-100.101),
			Code:     currency.GBP,
			Expected: money.New(-10010, "GBP"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			actual := conversions.NewFromDecimal(tc.Amount, tc.Code)
			assert.Equal(t, actual, tc.Expected)
		})
	}
}
