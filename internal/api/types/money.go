package types

import (
	"fmt"
	"strconv"
)

// Money keeps Amount as a decimal string (e.g. "12.4700") because the API
// serializes it that way to avoid float drift across PostgreSQL NUMERIC, JSON
// and float64. Convert only at the rendering boundary via AsFloat.
type Money struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// AsFloat parses Amount. A zero-value Money yields 0 and a non-nil error.
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

func (m Money) IsZero() bool {
	return m.Amount == "" && m.Currency == ""
}

// String renders as "$12.4700", or "-" when zero, or "<currency> <amount>" when
// the amount does not parse.
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
