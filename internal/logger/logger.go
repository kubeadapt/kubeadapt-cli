package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// debug=false returns a no-op logger; debug=true writes colored console output
// to stderr.
func New(debug bool) (*zap.Logger, error) {
	if !debug {
		return zap.NewNop(), nil
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	return cfg.Build()
}
