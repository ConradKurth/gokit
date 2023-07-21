package converter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ConradKurth/gokit/config"
	"github.com/ConradKurth/gokit/databases/memcached"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/contrib/instrumentation/github.com/bradfitz/gomemcache/memcache/otelmemcache"
	"golang.org/x/text/currency"
)

func latestKey(from currency.Unit) string {
	return fmt.Sprintf("%s-latest", from.String())
}

// CurrencyConverter represents the struct to do currency converting
type CurrencyConverter struct {
	cache  *otelmemcache.Client
	config *config.Config
}

// New initializes the CurrencyConverter struct
func New(c *config.Config, m *otelmemcache.Client) *CurrencyConverter {

	e := &CurrencyConverter{
		cache:  m,
		config: c,
	}

	return e
}

func toDecimalMap(rates map[string]float64) map[currency.Unit]decimal.Decimal {
	out := map[currency.Unit]decimal.Decimal{}
	for k, v := range rates {
		c, err := currency.ParseISO(k)
		// so we do this because not ALL of the currencies returned by the endpoint are supported by the package. soo that is why
		if err != nil {
			continue
		}
		out[c] = decimal.NewFromFloat(v)
	}
	return out
}

func (e *CurrencyConverter) getLatestRate(ctx context.Context, from currency.Unit) error {

	rate, err := e.fetchCurrencyExchange(from)
	if err != nil {
		return err
	}
	rates := toDecimalMap(rate.ConversionRates)
	memcached.SetItem(ctx, e.cache, latestKey(from), rates, memcached.WithExpiration(memcached.OneMinute*5))
	return nil
}

// LatestExchangeRate gets the latest rate from and to a currency based on the ISO 4217 Currency Code.
func (e *CurrencyConverter) LatestExchangeRate(ctx context.Context, from, to currency.Unit) (decimal.Decimal, error) {
	rates, err := e.LatestExchangeRates(ctx, from)
	if err != nil {
		return decimal.Zero, err
	}
	return rates[to], nil
}

// LatestExchangeRates get all the rates from one currency to another
func (e *CurrencyConverter) LatestExchangeRates(ctx context.Context, from currency.Unit) (map[currency.Unit]decimal.Decimal, error) {
	var rates map[currency.Unit]decimal.Decimal
	if !memcached.GetItem(ctx, e.cache, latestKey(from), &rates, true) {
		if err := e.getLatestRate(ctx, from); err != nil {
			return nil, err
		}
		return e.LatestExchangeRates(ctx, from)
	}
	return rates, nil
}

// GetHistoricalRate gets the rate from and to a currency based on the ISO 4217 Currency Code and date.
func (e *CurrencyConverter) GetHistoricalRate(ctx context.Context, from, to currency.Unit, date time.Time) (decimal.Decimal, error) {
	var rates map[currency.Unit]decimal.Decimal
	queryDate := date.Format("2006-01-02")
	currentTime := time.Now().UTC().Format("2006-01-02")
	if queryDate == currentTime { // current date is just latest rate
		return e.LatestExchangeRate(ctx, from, to)
	}

	cacheKey := queryDate + from.String()

	if memcached.GetItem(ctx, e.cache, cacheKey, &rates, false) {
		return rates[to], nil
	}

	url := fmt.Sprintf(`https://v6.exchangerate-api.com/v6/%v/history/%v/%v/%v/%v`, e.config.GetString("exchangerate.key"), from, date.Year(), int(date.Month()), date.Day())
	rate, err := requestHelper(url)
	if err != nil {
		return decimal.Zero, err
	}

	rates = toDecimalMap(rate.ConversionRates)
	memcached.SetItem(ctx, e.cache, cacheKey, rates)

	return rates[to], nil
}

type exchangeResponse struct {
	Result             string             `json:"result"`
	Documentation      string             `json:"documentation"`
	TermsOfUse         string             `json:"terms_of_use"`
	TimeLastUpdateUnix int                `json:"time_last_update_unix"`
	TimeLastUpdateUtc  string             `json:"time_last_update_utc"`
	TimeNextUpdateUnix int                `json:"time_next_update_unix"`
	TimeNextUpdateUtc  string             `json:"time_next_update_utc"`
	BaseCode           string             `json:"base_code"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
}

func requestHelper(url string) (exchangeResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return exchangeResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return exchangeResponse{}, errors.New("Unable to read body")
		}
		return exchangeResponse{}, fmt.Errorf("%v", string(b))
	}

	var exResp exchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&exResp); err != nil {
		return exchangeResponse{}, err
	}
	if exResp.Result != "success" {
		return exchangeResponse{}, errors.New("currency error")
	}

	return exResp, nil
}

func (e *CurrencyConverter) fetchCurrencyExchange(from currency.Unit) (exchangeResponse, error) {
	return requestHelper(fmt.Sprintf("https://v6.exchangerate-api.com/v6/%v/latest/%v", e.config.GetString("exchangerate.key"), from.String()))
}
