package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func init() {
	conf := zap.NewDevelopmentConfig()

	// 从环境变量中读取日志级别
	logLevel := os.Getenv("LOG")

	switch logLevel {
	case "DEBUG":
		conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "INFO":
		conf.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "WARN":
		conf.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "ERROR":
		conf.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "FATAL":
		conf.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		conf.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	}

	// 配置日志编码和格式
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	conf.EncoderConfig.MessageKey = "msg"
	// conf.OutputPaths = []string{"stdout"}

	logger, _ := conf.Build()
	Log = logger.Sugar()
}
