package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger 封装Zap日志
type ZapLogger struct {
	*zap.Logger
}

// NewZapLogger 创建日志实例
func NewZapLogger(level string) *ZapLogger {
	lvl := zapcore.InfoLevel
	switch level {
	case "debug":
		lvl = zapcore.DebugLevel
	case "warn":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	}

	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(lvl),
		Development:       false,
		DisableCaller:     true,  // 禁用调用者信息，正常日志不显示行号
		DisableStacktrace: false, // 启用堆栈跟踪，错误时显示
		Sampling:          nil,
		Encoding:          "json",
		OutputPaths:       []string{"stdout"}, // 输出到控制台，生产环境可改为文件路径
		ErrorOutputPaths:  []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      zapcore.OmitKey, // 不显示调用者键
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace", // 保留堆栈跟踪键，用于错误日志
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
	}

	logger, _ := cfg.Build(
		zap.AddStacktrace(zapcore.ErrorLevel), // 仅在错误级别时添加堆栈跟踪
	)
	return &ZapLogger{Logger: logger}
}

// NewZapLoggerWithOptions 创建带自定义选项的日志实例
func NewZapLoggerWithOptions(level string, development bool, outputPaths ...string) *ZapLogger {
	lvl := zapcore.InfoLevel
	switch level {
	case "debug":
		lvl = zapcore.DebugLevel
	case "warn":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	}

	// 默认输出路径
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}

	var cfg zap.Config
	if development {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder // 开发环境显示完整路径
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder // 生产环境显示短路径
	}

	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.OutputPaths = outputPaths
	cfg.DisableCaller = false
	cfg.EncoderConfig.CallerKey = "caller"

	logger, err := cfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		// 如果构建失败，返回一个基本的logger
		basicLogger, _ := zap.NewProduction()
		return &ZapLogger{Logger: basicLogger}
	}

	return &ZapLogger{Logger: logger}
}

// GinZapMiddleware Gin框架日志中间件
func GinZapMiddleware(logger *ZapLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		logger.Info("请求日志",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	}
}

// Error 创建error类型的字段
func Error(err error) zap.Field {
	return zap.Error(err)
}

// String 创建string类型的字段
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

// Int 创建int类型的字段
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Duration 创建duration类型的字段
func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

// Any 创建任意类型的字段（新增）
func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

// Info 记录信息级别日志
func (l *ZapLogger) Info(msg string, fields ...interface{}) {
	if len(fields)%2 != 0 {
		l.Logger.Info(msg)
		return
	}
	var zapFields []zap.Field
	for i := 0; i < len(fields); i += 2 {
		key := fmt.Sprintf("%v", fields[i])
		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}
	l.Logger.Info(msg, zapFields...)
}

// Debug 记录调试级别日志
func (l *ZapLogger) Debug(msg string, fields ...interface{}) {
	if len(fields)%2 != 0 {
		l.Logger.Debug(msg)
		return
	}
	var zapFields []zap.Field
	for i := 0; i < len(fields); i += 2 {
		key := fmt.Sprintf("%v", fields[i])
		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}
	l.Logger.Debug(msg, zapFields...)
}

// Warn 记录警告级别日志
func (l *ZapLogger) Warn(msg string, fields ...interface{}) {
	if len(fields)%2 != 0 {
		l.Logger.Warn(msg)
		return
	}
	var zapFields []zap.Field
	for i := 0; i < len(fields); i += 2 {
		key := fmt.Sprintf("%v", fields[i])
		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}
	l.Logger.Warn(msg, zapFields...)
}

// Error 记录错误级别日志
func (l *ZapLogger) Error(msg string, fields ...interface{}) {
	if len(fields)%2 != 0 {
		l.Logger.Error(msg)
		return
	}
	var zapFields []zap.Field
	for i := 0; i < len(fields); i += 2 {
		key := fmt.Sprintf("%v", fields[i])
		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}
	l.Logger.Error(msg, zapFields...)
}

// ErrorWithStack 记录错误级别日志并包含堆栈信息
func (l *ZapLogger) ErrorWithStack(msg string, err error, fields ...interface{}) {
	if len(fields)%2 != 0 {
		l.Logger.Error(msg, zap.Error(err), zap.Stack("stack"))
		return
	}
	var zapFields []zap.Field
	zapFields = append(zapFields, zap.Error(err), zap.Stack("stack"))
	for i := 0; i < len(fields); i += 2 {
		key := fmt.Sprintf("%v", fields[i])
		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}
	l.Logger.Error(msg, zapFields...)
}

// WithFields 创建带字段的日志上下文
func (l *ZapLogger) WithFields(fields ...interface{}) *ZapLogger {
	if len(fields)%2 != 0 {
		return l
	}
	var zapFields []zap.Field
	for i := 0; i < len(fields); i += 2 {
		key := fmt.Sprintf("%v", fields[i])
		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return &ZapLogger{Logger: l.Logger.With(zapFields...)}
}

// Fatal 记录致命错误日志并退出程序
func (l *ZapLogger) Fatal(msg string, fields ...interface{}) {
	if len(fields)%2 != 0 {
		l.Logger.Fatal(msg)
		return
	}
	var zapFields []zap.Field
	for i := 0; i < len(fields); i += 2 {
		key := fmt.Sprintf("%v", fields[i])
		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}
	l.Logger.Fatal(msg, zapFields...)
}
