package region

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ConradKurth/gokit/logger"
	"github.com/ConradKurth/gokit/responses"
	"golang.org/x/text/currency"
)

// Region are different regions we support accepting currency in
type Region string

const (
	regionHTTPHeader = "X-Region"

	// USA is the name of the region we use for USD based payments
	USA Region = "usa"
	// HongKong is name of the region we use for HKD based payments
	HongKong Region = "hongkong"
)

var (
	defaultCurrency  = currency.USD
	DefaultRegion    = USA
	regionCurrencies = map[Region]currency.Unit{
		USA:      currency.USD,
		HongKong: currency.HKD,
	}
)

// String will turn this into a string
func (r Region) String() string {
	return string(r)
}

// GetRegionCurrency will get the currency associated with the region
func (r Region) GetRegionCurrency() (currency.Unit, error) {
	cur, ok := regionCurrencies[r]
	if ok {
		return cur, nil
	}
	return defaultCurrency, fmt.Errorf("invalid region: %s", r)
}

func (r Region) validate() error {
	_, ok := regionCurrencies[r]
	if ok {
		return nil
	}
	return fmt.Errorf("unsupported local: %s", r)
}

type contextKey string

func (c contextKey) String() string {
	return "region-key-" + string(c)
}

const regionKey = contextKey("")

// NewMiddleware will create a new region middleware to inject our region into it.
func NewMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := loadRegionFromHeaders(r.Context(), r.Header)
			if err != nil {
				logger.GetLogger(r.Context()).ErrorCtx(r.Context(),
					"unable to add region from http",
					logger.ErrField(err))
				responses.Empty(w, http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRegion will get the users local
func GetRegion(ctx context.Context) (Region, error) {
	v, ok := ctx.Value(regionKey).(Region)
	if !ok {
		return "", errors.New("no region is set")
	}
	return v, nil
}

// AddToCtx will add a location to a context
func AddToCtx(ctx context.Context, r Region) (context.Context, error) {
	if err := r.validate(); err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, regionKey, r), nil
}

func loadRegionFromHeaders(ctx context.Context, headers http.Header) (context.Context, error) {
	region := Region(strings.ToLower(headers.Get(regionHTTPHeader)))
	_, ok := regionCurrencies[region]
	if ok {
		return AddToCtx(ctx, region)
	}
	return AddToCtx(ctx, DefaultRegion)
}
