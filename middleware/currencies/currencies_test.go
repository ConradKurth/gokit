package currencies_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ConradKurth/gokit/middleware/currencies"
	"github.com/ConradKurth/gokit/middleware/userinfo"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/currency"
)

func Test_ConditionallyInjectWithCurrency(t *testing.T) {
	ctx := context.Background()
	ctx, err := currencies.ConditionallyInjectWithCurrency(ctx, currency.HKD)
	assert.NoError(t, err)

	out, err := currencies.GetCurrency(ctx)
	assert.NoError(t, err)
	assert.Equal(t, currency.HKD, out)
}

func Test_ConditionallyInjectWithCurrency_HasBeenSet(t *testing.T) {
	ctx := context.Background()
	ctx, err := currencies.InjectCurrency(ctx, currency.USD)
	assert.NoError(t, err)

	ctx, err = currencies.ConditionallyInjectWithCurrency(ctx, currency.HKD)
	assert.NoError(t, err)

	out, err := currencies.GetCurrency(ctx)
	assert.NoError(t, err)
	assert.Equal(t, currency.USD, out)
}

func TestCurrenciesMiddleware(t *testing.T) {
	tt := []struct {
		Name        string
		Expected    currency.Unit
		Header      map[string]string
		ErrExpected error
	}{
		{
			Name:     "Default to USD when currency set",
			Expected: currencies.DefaultCurrency,
		},
		{
			Name:     "USD from the currency header",
			Expected: currency.USD,
			Header: map[string]string{
				currencies.CurrencyHTTPHeader: "USD",
			},
		},
		{
			Name:     "HKD from the region header",
			Expected: currency.HKD,
			Header: map[string]string{
				currencies.CurrencyHTTPHeader: "HKD",
			},
		},
		{
			Name:        "Unsupported currency",
			ErrExpected: currencies.ErrInvalidCurrencySet,
			Header: map[string]string{
				currencies.CurrencyHTTPHeader: "unsupported",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(userinfo.NewMiddleware())
			r.Use(currencies.NewMiddleware())

			r.Get("/v1/do", func(w http.ResponseWriter, r *http.Request) {
				l, err := currencies.GetCurrency(r.Context())
				if tc.ErrExpected != nil {
					assert.EqualError(t, err, tc.ErrExpected.Error())
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, tc.Expected, l)
			})

			req, err := http.NewRequest(http.MethodGet, "/v1/do", nil)
			require.NoError(t, err)
			for k, v := range tc.Header {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			expectedCode := http.StatusOK
			if tc.ErrExpected != nil {
				expectedCode = http.StatusBadRequest
			}
			assert.Equal(t, rec.Code, expectedCode)
		})
	}
}
