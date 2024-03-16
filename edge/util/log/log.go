package logutil

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger = newConsoleLogger()

func newConsoleLogger() *zap.SugaredLogger {
	enc := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:       "_",
		LevelKey:         "_",
		TimeKey:          "_",
		CallerKey:        "_",
		StacktraceKey:    "_",
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.RFC3339TimeEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	})
	core := zapcore.NewCore(enc, zapcore.Lock(os.Stdout), zapcore.DebugLevel)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()
}

func GetLogger() *zap.SugaredLogger {
	return logger
}
