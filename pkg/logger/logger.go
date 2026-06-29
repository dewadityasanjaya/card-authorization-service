package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(env string) {
	var err error

	if env == "production" {
		// Production: JSON format, only Info level and above
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		Log, err = cfg.Build()
	} else {
		// Development: human-readable, colorized, Debug level
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		Log, err = cfg.Build()
	}

	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

// Convenience wrappers so you don't import zap everywhere
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries — call this on app shutdown
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
