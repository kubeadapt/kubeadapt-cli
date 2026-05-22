package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoney_AsFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		money   Money
		want    float64
		wantErr bool
	}{
		{"typical USD amount", Money{Amount: "12.4700", Currency: "USD"}, 12.47, false},
		{"zero amount with currency", Money{Amount: "0.0000", Currency: "USD"}, 0.0, false},
		{"larger amount", Money{Amount: "1234.5678", Currency: "USD"}, 1234.5678, false},
		{"empty amount", Money{Amount: "", Currency: "USD"}, 0, true},
		{"not a number", Money{Amount: "not-a-number", Currency: "USD"}, 0, true},
		{"completely empty", Money{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.money.AsFloat()
			require.Equal(t, tt.wantErr, err != nil, "AsFloat() err = %v, wantErr = %v", err, tt.wantErr)
			assert.InDelta(t, tt.want, got, 1e-9)
		})
	}
}

func TestMoney_IsZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		money Money
		want  bool
	}{
		{"zero value", Money{}, true},
		{"zero amount but currency set", Money{Amount: "0.0000", Currency: "USD"}, false},
		{"amount only", Money{Amount: "12.4700"}, false},
		{"currency only", Money{Currency: "USD"}, false},
		{"both populated", Money{Amount: "12.4700", Currency: "USD"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.money.IsZero())
		})
	}
}

func TestMoney_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		money Money
		want  string
	}{
		{"USD formatted with dollar sign", Money{Amount: "12.4700", Currency: "USD"}, "$12.4700"},
		{"EUR formatted with currency code", Money{Amount: "12.4700", Currency: "EUR"}, "EUR 12.4700"},
		{"zero value renders as dash", Money{}, "-"},
		{"unparseable amount falls back verbatim", Money{Amount: "abc", Currency: "USD"}, "abc USD"},
		{"USD zero amount keeps four decimals", Money{Amount: "0.0000", Currency: "USD"}, "$0.0000"},
		{"GBP formatting", Money{Amount: "1234.5678", Currency: "GBP"}, "GBP 1234.5678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.money.String())
		})
	}
}

func TestMoney_JSONUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  Money
	}{
		{
			name:  "standard USD payload",
			input: `{"amount":"12.4700","currency":"USD"}`,
			want:  Money{Amount: "12.4700", Currency: "USD"},
		},
		{
			name:  "EUR payload",
			input: `{"amount":"99.9999","currency":"EUR"}`,
			want:  Money{Amount: "99.9999", Currency: "EUR"},
		},
		{
			name:  "empty object",
			input: `{}`,
			want:  Money{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Money
			require.NoError(t, json.Unmarshal([]byte(tt.input), &got))
			assert.Equal(t, tt.want, got)
		})
	}
}
