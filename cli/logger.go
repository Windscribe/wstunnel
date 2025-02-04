package cli

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"time"
)

var Logger *zap.SugaredLogger

// InitLogger initializes the logger.
func InitLogger(development bool, logFilePath string) {
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

	zapLogger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		log.Fatal(err)
	}

	Logger = zapLogger.With(zap.String("mod", "wstunnel")).Sugar()

	if logFilePath != "" {
		Logger.Info("Logging to stdout and file: ", zap.String("file", logFilePath))
	} else {
		Logger.Info("Logging to stdout")
	}
}

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func levelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(level.String())
}
