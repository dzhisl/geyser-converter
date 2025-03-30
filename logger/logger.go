package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() {
	var cfg zap.Config

	cfg = zap.NewDevelopmentConfig()
	cfg.EncoderConfig.StacktraceKey = ""
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Build the logger and handle any errors
	var err error
	logger, err := cfg.Build()
	if err != nil {
		zap.L().Fatal("Error building logger", zap.Error(err))
	}

	// Set the global logger to the newly created instance
	zap.ReplaceGlobals(logger)
}
