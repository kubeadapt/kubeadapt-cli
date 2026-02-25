package output

import "fmt"

// FormatCost formats a cost value as a dollar amount.
func FormatCost(v float64) string {
	return fmt.Sprintf("$%.2f", v)
}

// FormatCostPtr formats an optional cost value.
func FormatCostPtr(v *float64) string {
	if v == nil {
		return "-"
	}
	return FormatCost(*v)
}

// FormatPercent formats a percentage value.
func FormatPercent(v float64) string {
	return fmt.Sprintf("%.1f%%", v)
}

// FormatPercentPtr formats an optional percentage value.
func FormatPercentPtr(v *float64) string {
	if v == nil {
		return "-"
	}
	return FormatPercent(*v)
}

// FormatMemoryGB formats a memory value in GB.
func FormatMemoryGB(v float64) string {
	if v < 1 {
		return fmt.Sprintf("%.0f MB", v*1024)
	}
	return fmt.Sprintf("%.1f GB", v)
}

// FormatBool formats a boolean value as Yes/No.
func FormatBool(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

// FormatOptionalString formats a pointer-to-string value, or returns "-" if nil.
func FormatOptionalString(v *string) string {
	if v == nil {
		return "-"
	}
	return *v
}

// FormatInt formats an integer value.
func FormatInt(v int) string {
	return fmt.Sprintf("%d", v)
}

// FormatFloat formats a float value.
func FormatFloat(v float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, v)
}

// FormatFloatPtr formats an optional float value.
func FormatFloatPtr(v *float64, decimals int) string {
	if v == nil {
		return "-"
	}
	return FormatFloat(*v, decimals)
}

// FormatMemoryGBPtr formats an optional memory value in GB.
func FormatMemoryGBPtr(v *float64) string {
	if v == nil {
		return "-"
	}
	return FormatMemoryGB(*v)
}

// FormatIntPtr formats an optional integer value.
func FormatIntPtr(v *int) string {
	if v == nil {
		return "-"
	}
	return FormatInt(*v)
}

// ShortID truncates a UUID or long ID to the first 8 characters.
func ShortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
