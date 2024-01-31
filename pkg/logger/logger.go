package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func init() {
	conf := zap.NewDevelopmentConfig()
	conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	conf.EncoderConfig.MessageKey = "msg"
	logger, _ := conf.Build()
	Log = logger.Sugar()

}
