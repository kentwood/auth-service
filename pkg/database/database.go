package database

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"auth-service/internal/config"
)

// Connect 根据配置连接数据库
func Connect(cfg *config.DBConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	// 如果配置了DSN，优先使用DSN（向后兼容）
	if cfg.DSN != "" {
		// 根据DSN内容判断数据库类型
		if strings.Contains(cfg.DSN, "postgres") || strings.Contains(cfg.DSN, "host=") {
			dialector = postgres.Open(cfg.DSN)
		} else {
			dialector = mysql.Open(cfg.DSN)
		}
	} else {
		// 使用新的配置结构
		switch cfg.Type {
		case "postgres":
			dialector = postgres.Open(buildPostgresDSN(cfg))
		case "mysql":
			dialector = mysql.Open(buildMysqlDSN(cfg))
		default:
			return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.Type)
		}
	}

	// GORM 配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 可根据环境调整
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	return db, nil
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && (s[len(substr)] == ' ' || s[len(substr)] == '=' || s[len(substr)] == ':'))
}

// buildPostgresDSN 构建PostgreSQL连接字符串
func buildPostgresDSN(cfg *config.DBConfig) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
		cfg.SSLMode,
	)
}

// buildMysqlDSN 构建MySQL连接字符串
func buildMysqlDSN(cfg *config.DBConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
		cfg.ParseTime,
		cfg.Loc,
	)
}
