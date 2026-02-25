package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// JSON writes the value as indented JSON to stdout.
func JSON(v interface{}) error {
	return JSONTo(os.Stdout, v)
}

// JSONTo writes the value as indented JSON to the given writer.
func JSONTo(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}
