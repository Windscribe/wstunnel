package cli

import (
	"fmt"
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
		cfg.OutputPaths = []string{logFilePath, "stdout"}
	}
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
	cfg.EncoderConfig.EncodeLevel = levelEncoder
	cfg.EncoderConfig.EncodeTime = syslogTimeEncoder
	cfg.EncoderConfig.CallerKey = ""
	cfg.Development = development
	if !development {
		cfg.EncoderConfig.StacktraceKey = ""
	}
	zapLogger, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}
	Logger = zapLogger.Sugar()
	defer Logger.Sync()
}

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func levelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]-", level))
}
