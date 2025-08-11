package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func init() {
	conf := zap.NewDevelopmentConfig()

	// Retrieve log level from environment variable
	logLevel := os.Getenv("LOG")

	level := zap.WarnLevel // default level
	switch logLevel {
	case "DEBUG":
		level = zap.DebugLevel
	case "INFO":
		level = zap.InfoLevel
	case "WARN":
		level = zap.WarnLevel
	case "ERROR":
		level = zap.ErrorLevel
	}
	conf.Level = zap.NewAtomicLevelAt(level)

	// Configure the encoding and format of the logs
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	conf.EncoderConfig.MessageKey = "msg"
	// conf.OutputPaths = []string{"stdout"}

	logger, _ := conf.Build(zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	// logger, _ := conf.Build()
	Log = logger.Sugar()
}

func Debugw(msg string, keysAndValues ...interface{}) {
	Log.Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	Log.Infow(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	Log.Warnw(msg, keysAndValues...)
}

func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

func Warn(args ...interface{}) {
	Log.Warn(args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	Log.Errorw(msg, keysAndValues...)
}

func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	Log.Fatalw(msg, keysAndValues...)
}

func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

// Fatal constructs a message with the provided arguments and calls os.Exit.
// Spaces are added between arguments when neither is a string.
func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

func CheckAndLogError(msg string, err error) {
	if err != nil {
		Log.Errorw(msg, "error", err)
	}
}

// func Info(args ...interface{}) {
//	Log.Info(args...)
// }
