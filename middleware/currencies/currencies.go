package currencies

import (
	"context"
	"errors"
	"net/http"

	"github.com/ConradKurth/gokit/logger"
	"github.com/ConradKurth/gokit/responses"
	"golang.org/x/text/currency"
)

type contextKey string

func (c contextKey) String() string {
	return "currency-key-" + string(c)
}

const (
	locationKey        = contextKey("")
	CurrencyHTTPHeader = "X-Currency"
)

var (
	// ErrNoCurrencySet is returned when the context contains no currency.
	ErrNoCurrencySet = errors.New("no currency is set")
	// ErrInvalidCurrencySet is returned if an unsupported currency is set in the context.
	ErrInvalidCurrencySet = errors.New("invalid currency is set")

	DefaultCurrency = currency.USD
	validCurrencies = map[currency.Unit]struct{}{
		currency.USD: {},
		currency.GBP: {},
		currency.EUR: {},
		currency.CNY: {},
		currency.CAD: {},
		currency.HKD: {},
	}
)

// NewMiddleware will create a new currency middleware to inject our location into it.
func NewMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get(CurrencyHTTPHeader)
			if header == "" {
				header = DefaultCurrency.String()
			}

			cur, err := currency.ParseISO(header)
			if err != nil {
				logger.GetLogger(r.Context()).ErrorCtx(r.Context(),
					"unable to parse currency",
					logger.ErrField(err),
					logger.Any("currency", header))
				responses.Empty(w, http.StatusBadRequest)
				return
			}

			ctx, err := AddToCtx(r.Context(), cur)
			if err != nil {
				logger.GetLogger(r.Context()).ErrorCtx(r.Context(),
					"unable to add currency",
					logger.ErrField(err),
					logger.Any("currency", header))
				responses.Empty(w, http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetCurrency will get the users local currency.
func GetCurrency(ctx context.Context) (currency.Unit, error) {
	v, ok := ctx.Value(locationKey).(currency.Unit)
	if !ok {
		return DefaultCurrency, ErrNoCurrencySet
	}
	return v, nil
}

// AddToCtx will add a currency to a context.
func AddToCtx(ctx context.Context, cur currency.Unit) (context.Context, error) {
	if !isValidCurrency(cur) {
		return ctx, ErrInvalidCurrencySet
	}
	return context.WithValue(ctx, locationKey, cur), nil
}

// ConditionallyInjectWithCurrency will only add the currency if it is not been set.
// Useful for situation where methods can be called from the API as well as lambdas.
func ConditionallyInjectWithCurrency(ctx context.Context, cur currency.Unit) (context.Context, error) {
	_, ok := ctx.Value(locationKey).(currency.Unit)
	if ok {
		return ctx, nil
	}
	return AddToCtx(ctx, cur)
}

// InjectCurrency will inject the location into a context based on the currency.
func InjectCurrency(ctx context.Context, cur currency.Unit) (context.Context, error) {
	return AddToCtx(ctx, cur)
}

// MustInjectCurrency will inject the location into a context based on the currency
// and panic if there is an error. Not great to use except for tests.
func MustInjectCurrency(ctx context.Context, cur currency.Unit) context.Context {
	ctx, err := AddToCtx(ctx, cur)
	if err != nil {
		panic(err)
	}
	return ctx
}

func isValidCurrency(c currency.Unit) bool {
	_, ok := validCurrencies[c]
	return ok
}
