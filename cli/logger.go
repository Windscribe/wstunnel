package cli

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// Logger is never nil, so the proxy can log before or without InitLogger.
var Logger = zap.NewNop().Sugar()

// InitLogger initializes the logger. It never exits the process: the proxy runs inside the host app or its VPN
// extension, so a log file that can't be opened (for example on iOS before the first unlock) falls back to stdout.
func InitLogger(development bool, logFilePath string) {
	Logger = newLogger(development, logFilePath)
	if logFilePath != "" {
		Logger.Info("Logging to stdout and file: ", zap.String("file", logFilePath))
	} else {
		Logger.Info("Logging to stdout")
	}
}

// newLogger builds the proxy's logger. It always returns one: stdout only when the file can't be opened, and a
// no-op logger when even that fails.
func newLogger(development bool, logFilePath string) *zap.SugaredLogger {
	zapLogger, err := buildLogger(development, logFilePath)
	if err == nil {
		return zapLogger.With(zap.String("mod", "wstunnel")).Sugar()
	}
	fileErr := err
	if logFilePath != "" {
		if zapLogger, err = buildLogger(development, ""); err == nil {
			logger := zapLogger.With(zap.String("mod", "wstunnel")).Sugar()
			logger.Errorf("Can't log to %s, logging to stdout only: %s", logFilePath, fileErr)
			return logger
		}
	}
	return zap.NewNop().Sugar()
}

func buildLogger(development bool, logFilePath string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	outputPaths := []string{"stdout"}
	if logFilePath != "" {
		outputPaths = append(outputPaths, logFilePath)
	}
	cfg.OutputPaths = outputPaths

	cfg.Encoding = "json"
	cfg.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
	cfg.EncoderConfig.EncodeLevel = levelEncoder
	cfg.EncoderConfig.EncodeTime = syslogTimeEncoder
	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.CallerKey = ""
	cfg.EncoderConfig.NameKey = "mod"
	cfg.EncoderConfig.TimeKey = "tm" // Important: Set the TimeKey

	if !development {
		cfg.EncoderConfig.StacktraceKey = ""
	}

	return cfg.Build(zap.AddCallerSkip(1))
}

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func levelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(level.String())
}
