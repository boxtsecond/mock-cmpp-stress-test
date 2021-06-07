package log

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Dir       string `toml:"dir"`
	File      string `toml:"file"`
	Level     string `toml:"level"`
	LocalTime bool   `toml:"local_time"`
}

var Logger *zap.Logger

func Init(cfg *Config) {
	fileName := cfg.Dir + "/" + cfg.File
	hook := lumberjack.Logger{
		Filename:  fileName,
		LocalTime: cfg.LocalTime,
	}

	w := zapcore.AddSync(&hook)

	var zapLevel zapcore.Level
	switch cfg.Level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}
	encoderConfig := zap.NewProductionEncoderConfig()

	// 时间格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		zapLevel,
	)

	Logger = zap.New(core, zap.AddCaller(), zap.Development())
}
