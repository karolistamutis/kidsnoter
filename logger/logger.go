package logger

import (
	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

// init initializes the logger with default production settings
func init() {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel) // Set default level
	zapLogger, err := cfg.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	Log = zapLogger.Sugar()
}

// Configure sets up the logger based on the verbose flag
func Configure(verbose int) {
	var cfg zap.Config
	switch verbose {
	case 0:
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case 1:
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case 2:
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	default:
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	zapLogger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	Log = zapLogger.Sugar()
}

// Sync flushes any buffered log entries
func Sync() {
	_ = Log.Sync()
}
