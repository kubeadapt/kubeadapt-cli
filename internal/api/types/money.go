package types

import (
	"fmt"
	"strconv"
)

// Money represents a monetary amount as returned by the Kubeadapt API.
//
// The API always serializes monetary amounts as decimal STRINGS with four
// decimal places (for example "12.4700") to eliminate float drift across
// PostgreSQL NUMERIC, JSON, and Go's float64. We preserve that contract here:
// Amount is kept as a string and converted to float64 only at the rendering
// boundary via AsFloat.
type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// AsFloat parses the Amount field as a float64. It returns a non-nil error
// when the amount is empty or not a valid decimal. Callers should treat a
// zero-value Money (Amount == "" && Currency == "") as "no value" — AsFloat
// returns 0 and a non-nil error in that case.
func (m Money) AsFloat() (float64, error) {
	if m.Amount == "" {
		return 0, fmt.Errorf("money: empty amount")
	}
	f, err := strconv.ParseFloat(m.Amount, 64)
	if err != nil {
		return 0, fmt.Errorf("money: parse amount %q: %w", m.Amount, err)
	}
	return f, nil
}

// IsZero reports whether the Money is the zero value (no amount AND no currency).
func (m Money) IsZero() bool {
	return m.Amount == "" && m.Currency == ""
}

// String renders Money for human-friendly output. Example: `$12.4700 USD`.
// Falls back to "-" if the Money is the zero value, or "<currency> <amount>"
// if the currency code is unknown. Always prints exactly four decimal places
// when the amount parses as a number.
func (m Money) String() string {
	if m.IsZero() {
		return "-"
	}
	f, err := strconv.ParseFloat(m.Amount, 64)
	if err != nil {
		return m.Amount + " " + m.Currency
	}
	formatted := fmt.Sprintf("%.4f", f)
	if m.Currency == "USD" {
		return "$" + formatted
	}
	return m.Currency + " " + formatted
}
