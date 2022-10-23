package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"time"
)

var Logger *zap.SugaredLogger

// InitLogger Log output is saved to the app provided log file.
func InitLogger(development bool, logFilePath string) {
	cfg := zap.NewProductionConfig()
	if logFilePath != "" {
		cfg.OutputPaths = []string{logFilePath}
	}
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = syslogTimeEncoder
	cfg.Development = development
	if !development {
		cfg.EncoderConfig.StacktraceKey = ""
		cfg.EncoderConfig.CallerKey = ""
	}
	zapLogger, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}
	Logger = zapLogger.Sugar()
	defer Logger.Sync()
}

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("Jan  2 15:04:05"))
}
