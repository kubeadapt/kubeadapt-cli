package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a zap.Logger for CLI use.
// debug=true  → human-readable colored console output to stderr
// debug=false → no-op logger (zero allocation overhead)
func New(debug bool) (*zap.Logger, error) {
	if !debug {
		return zap.NewNop(), nil
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	return cfg.Build()
}
