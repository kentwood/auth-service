package config

import (
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Log    LogConfig    `mapstructure:"log"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Mode string `mapstructure:"mode"` // debug/release/test
	Port string `mapstructure:"port"` // 服务端口，如"8080"
}

// DBConfig 数据库配置
type DBConfig struct {
	DSN string `mapstructure:"dsn"` // PostgreSQL连接串，如"host=localhost user=postgres password=1234 dbname=auth port=5432 sslmode=disable"
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"` // JWT签名密钥
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `mapstructure:"level"` // debug/info/warn/error
}

// Load 从配置文件加载配置
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
