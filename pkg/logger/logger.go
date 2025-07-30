package logger

import (
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
		Level:       zap.NewAtomicLevelAt(lvl),
		Encoding:    "json",
		OutputPaths: []string{"stdout"}, // 输出到控制台，生产环境可改为文件路径
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "time",
			LevelKey:     "level",
			MessageKey:   "msg",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger, _ := cfg.Build()
	return &ZapLogger{logger}
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
