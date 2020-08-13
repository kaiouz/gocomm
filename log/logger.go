package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.SugaredLogger

// NewStdLog 创建golang内建的logger
func NewStdLog() *log.Logger {
	return zap.NewStdLog(logger.Desugar())
}

// Info 输出info日志
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Warn 输出warn日志
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error 输出Error日志
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Debug 输出Debug日志
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Infof 输出info日志
func Infof(fmt string, args ...interface{}) {
	logger.Infof(fmt, args...)
}

// Warnf 输出warn日志
func Warnf(fmt string, args ...interface{}) {
	logger.Warnf(fmt, args...)
}

// Errorf 输出Error日志
func Errorf(fmt string, args ...interface{}) {
	logger.Errorf(fmt, args...)
}

// Debugf 输出Debug日志
func Debugf(fmt string, args ...interface{}) {
	logger.Debugf(fmt, args...)
}

// Init 初始化日志工具
func Init(debug bool, logPath, serviceName, address string) error {
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})
	middlePriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel && lvl < zapcore.ErrorLevel
	})

	var cores []zapcore.Core

	// debug下只输出到控制台, 非debug下输出到文件和日志服务
	if debug {
		consoleDebugging := zapcore.Lock(os.Stdout)
		consoleErrors := zapcore.Lock(os.Stderr)
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		cores = append(cores,
			zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
			zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
		)
	} else {
		// 日志文件
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		fileCommon := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Join(logPath, "common.log"),
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		})
		fileErrors := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Join(logPath, "error.log"),
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		})

		// 日志服务，只收集错误日志
		pbEncoder := newProtobufEncoder(zap.NewDevelopmentEncoderConfig(), serviceName)
		pbWs, err := newProtobufWriterSyncer(address)

		if err != nil {
			return fmt.Errorf("远程日志初始化错误, %v: %w", address, err)
		}

		cores = append(cores,
			zapcore.NewCore(consoleEncoder, fileErrors, highPriority),
			zapcore.NewCore(consoleEncoder, fileCommon, middlePriority),
			zapcore.NewCore(pbEncoder, pbWs, highPriority),
		)
	}

	core := zapcore.NewTee(cores...)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()

	return nil
}

// InitDev 初始化开发环境日志
func InitDev() error {
	return Init(true, "", "", "")
}

// InitProd 初始化生产环境日志
func InitProd(logPath, serviceName, address string) error {
	return Init(false, logPath, serviceName, address)
}

// Sync 同步日志, 刷新缓存
func Sync() {
	logger.Sync()
}
