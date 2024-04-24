package logger

import (
	"encoding/json"
	"ibanking-scraper/config"
	"io"
	"log"
	"os"
	"sync"

	glog "github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	config zap.Config
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

var (
	once sync.Once

	appLogger *Logger
)

func Setup() *Logger {
	conf := config.Load()
	return new(conf.Environment, conf.LogLevel)
}

func new(env, loglevel string) *Logger {
	once.Do(func() {
		cfg := getZapConfig(env)
		cfg.Level = zap.NewAtomicLevelAt(getLevel(loglevel))

		logger, err := cfg.Build()
		if err != nil {
			log.Fatalln(err)
		}

		appLogger = &Logger{
			config: cfg,
			logger: logger,
			sugar:  logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Sugar(),
		}

	})

	return appLogger
}

func getZapConfig(env string) zap.Config {
	switch env {
	case "production":
		return zap.Config{
			Encoding:         "json",
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig:    zap.NewProductionEncoderConfig(),
		}
	case "development":
		return zap.Config{
			Encoding:         "console",
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:    "message",
				StacktraceKey: "stackTrace",
				LevelKey:      "level",
				EncodeLevel:   zapcore.CapitalColorLevelEncoder,

				TimeKey:    "time",
				EncodeTime: zapcore.ISO8601TimeEncoder,

				CallerKey:    "caller",
				EncodeCaller: zapcore.ShortCallerEncoder,
			},
		}
	}

	return zap.NewDevelopmentConfig()
}

func getLevel(level string) zapcore.Level {
	switch level {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "CRITICAL":
		return zapcore.ErrorLevel
	default:
	}
	return zapcore.DebugLevel
}

func (l *Logger) Logger() *zap.Logger {
	return l.logger
}

func (l *Logger) SLogger() *zap.SugaredLogger {
	return l.sugar
}

func (l *Logger) Output() io.Writer {
	return os.Stdout
}

func (l *Logger) SetOutput(w io.Writer) {
	// do nothing
}

func (l *Logger) Prefix() string {
	// do nothing
	return ""
}
func (l *Logger) SetPrefix(p string) {
	// do nothing
}

func (l *Logger) Level() glog.Lvl {
	switch l.config.Level.Level() {
	case zapcore.DebugLevel:
		return glog.DEBUG
	case zapcore.InfoLevel:
		return glog.INFO
	case zapcore.WarnLevel:
		return glog.WARN
	case zapcore.ErrorLevel:
		return glog.ERROR
	default:
		return glog.OFF
	}
}

func (l *Logger) SetLevel(v glog.Lvl) {
	// do nothing
}
func (l *Logger) SetHeader(h string) {
	// do nothing
}

func (l *Logger) Print(i ...interface{}) {
	l.sugar.Infof("%v", i...)
}
func (l *Logger) Printf(format string, args ...interface{}) {
	l.sugar.Infof(format, args...)
}
func (l *Logger) Printj(j glog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		l.sugar.Panic(err)
	}
	l.sugar.Info(string(b))
}
func (l *Logger) Debug(i ...interface{}) {
	l.sugar.Debug(i...)
}
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.sugar.Debugf(format, args...)

}
func (l *Logger) Debugj(j glog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		l.sugar.Panic(err)
	}
	l.sugar.Debug(string(b))
}
func (l *Logger) Info(i ...interface{}) {
	l.sugar.Info(i...)
}
func (l *Logger) Infof(format string, args ...interface{}) {
	l.sugar.Infof(format, args...)
}
func (l *Logger) Infoj(j glog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		l.sugar.Panic(err)
	}
	l.sugar.Info(string(b))
}

func (l *Logger) Warn(i ...interface{}) {
	l.sugar.Warn(i...)
}
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.sugar.Warnf(format, args...)
}
func (l *Logger) Warnj(j glog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		l.sugar.Panic(err)
	}
	l.sugar.Warn(string(b))
}
func (l *Logger) Error(i ...interface{}) {
	l.sugar.Error(i...)
}
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.sugar.Errorf(format, args...)
}
func (l *Logger) Errorj(j glog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		l.sugar.Panic(err)
	}
	l.sugar.Error(string(b))
}
func (l *Logger) Fatal(i ...interface{}) {
	l.sugar.Fatal(i...)
}
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.sugar.Fatalf(format, args...)
}
func (l *Logger) Fatalj(j glog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		l.sugar.Panic(err)
	}
	l.sugar.Fatal(string(b))
}
func (l *Logger) Panic(i ...interface{}) {
	l.sugar.Panic(i...)
}
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.sugar.Panicf(format, args...)
}
func (l *Logger) Panicj(j glog.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		l.sugar.Panic(err)
	}
	l.sugar.Panic(string(b))
}
