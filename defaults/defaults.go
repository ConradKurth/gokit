package defaults

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/text/currency"
)

type Fields struct {
	Created time.Time `db:"created_at"`
	Update  time.Time `db:"updated_at"`
}

// RatesInfo is storing our rates in the database
type RatesInfo map[currency.Unit]decimal.Decimal

// Value returns the value of the rates
func (i RatesInfo) Value() (driver.Value, error) {
	return json.Marshal(i)
}

var emptyJSON = []byte("{}")

// Scan implements the interface to scan the value into the db
func (i *RatesInfo) Scan(src interface{}) error {
	var source []byte
	switch t := src.(type) {
	case nil:
		return nil
	case string:
		source = []byte(t)
	case []byte:
		if len(t) == 0 {
			source = emptyJSON
		} else {
			source = t
		}
	default:
		return errors.New("Incompatible type for JSONText")
	}

	if err := json.Unmarshal(source, i); err != nil {
		return err
	}

	return nil
}
