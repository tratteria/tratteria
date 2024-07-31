package logging

import (
    "os"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

func InitLogger() error {
    logLevel := os.Getenv("LOG_LEVEL")
    if logLevel == "" {
        logLevel = "info"
    }

    config := zap.Config{
        Encoding:         "json",
        Level:            zap.NewAtomicLevelAt(getLogLevel(logLevel)),
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
        EncoderConfig:    zapcore.EncoderConfig{
            TimeKey:        "timestamp",
            LevelKey:       "level",
            NameKey:        "logger",
            CallerKey:      "caller",
            FunctionKey:    zapcore.OmitKey,
            MessageKey:     "message",
            StacktraceKey:  "stacktrace",
            LineEnding:     zapcore.DefaultLineEnding,
            EncodeLevel:    zapcore.LowercaseLevelEncoder,
            EncodeTime:     zapcore.ISO8601TimeEncoder,
            EncodeDuration: zapcore.SecondsDurationEncoder,
            EncodeCaller:   zapcore.ShortCallerEncoder,
        },
    }

    logger, err := config.Build()
    if err != nil {
        return err
    }

    globalLogger = logger

	return nil
}

func GetLogger(component string) *zap.Logger {
    if globalLogger == nil {
        InitLogger()
    }
    return globalLogger.Named(component)
}

func getLogLevel(level string) zapcore.Level {
    switch level {
    case "debug":
        return zapcore.DebugLevel
    case "info":
        return zapcore.InfoLevel
    case "warn":
        return zapcore.WarnLevel
    case "error":
        return zapcore.ErrorLevel
    default:
        return zapcore.InfoLevel
    }
}

func Sync() {
    if globalLogger != nil {
        globalLogger.Sync()
    }
}