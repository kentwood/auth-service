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
	OAuth2 OAuth2Config `mapstructure:"oauth2"` // 新增 OAuth2 配置
}

// ServerConfig 服务配置
type ServerConfig struct {
	Mode string `mapstructure:"mode"` // debug/release/test
	Port string `mapstructure:"port"` // 服务端口，如"8080"
}

// DBConfig 数据库配置
type DBConfig struct {
	Type      string `mapstructure:"type"` // 数据库类型：postgres, mysql
	Host      string `mapstructure:"host"`
	Port      string `mapstructure:"port"`
	User      string `mapstructure:"user"`
	Password  string `mapstructure:"password"`
	DBName    string `mapstructure:"dbname"`
	SSLMode   string `mapstructure:"sslmode"`   // PostgreSQL SSL模式
	Charset   string `mapstructure:"charset"`   // MySQL 字符集
	ParseTime bool   `mapstructure:"parsetime"` // MySQL 时间解析
	Loc       string `mapstructure:"loc"`       // MySQL 时区
	// 保持向后兼容的DSN字段（如果配置了DSN，优先使用DSN）
	DSN string `mapstructure:"dsn"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"` // JWT签名密钥
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `mapstructure:"level"` // debug/info/warn/error
}

// OAuth2Config OAuth2 配置
type OAuth2Config struct {
	GitHub GitHubOAuth2Config `mapstructure:"github"`
}

// GitHubOAuth2Config GitHub OAuth2 配置
type GitHubOAuth2Config struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
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
