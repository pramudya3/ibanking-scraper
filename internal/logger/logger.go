package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InfoWithField(msg string, fields ...zapcore.Field) {
	log := appLogger.logger.WithOptions(zap.AddCallerSkip(1))
	log.Info(msg, fields...)
}

func WarnWithField(msg string, fields ...zapcore.Field) {
	log := appLogger.logger.WithOptions(zap.AddCallerSkip(1))
	log.Warn(msg, fields...)
}

func ErrorWithField(msg string, fields ...zapcore.Field) {
	appLogger.logger.Error(msg, fields...)
}

func Debug(i ...interface{}) {
	appLogger.sugar.Debug(i...)
}
func Debugf(format string, args ...interface{}) {
	appLogger.sugar.Debugf(format, args...)
}

func Info(i ...interface{}) {
	appLogger.sugar.Info(i...)
}
func Infof(format string, args ...interface{}) {
	appLogger.sugar.Infof(format, args...)
}

func Warn(i ...interface{}) {
	appLogger.sugar.Warn(i...)
}
func Warnf(format string, args ...interface{}) {
	appLogger.sugar.Warnf(format, args...)
}

func Error(i ...interface{}) {
	appLogger.sugar.Error(i...)
}

func Errorf(format string, args ...interface{}) {
	appLogger.sugar.Errorf(format, args...)
}

func Fatal(i ...interface{}) {
	appLogger.sugar.Fatal(i...)
}

func Fatalf(format string, args ...interface{}) {
	appLogger.sugar.Fatalf(format, args...)
}

func Panic(i ...interface{}) {
	appLogger.sugar.Panic(i...)
}
func Panicf(format string, args ...interface{}) {
	appLogger.sugar.Panicf(format, args...)
}
