package config

import (
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Log    LogConfig    `mapstructure:"log"`
	OAuth2 OAuth2Config `mapstructure:"oauth2"`
	UI     UIConfig     `mapstructure:"ui"` // 新增 UI 配置
}

// RedisConfig Redis 配置
type RedisConfig struct {
    Host     string `mapstructure:"host"`
    Port     string `mapstructure:"port"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
    Prefix   string `mapstructure:"prefix"` // key 前缀
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

// UIConfig 前端页面配置
type UIConfig struct {
	BaseURL          string `mapstructure:"base_url"`           // 前端基础URL
	LoginSuccessPath string `mapstructure:"login_success_path"` // 登录成功页面路径
	LoginErrorPath   string `mapstructure:"login_error_path"`   // 登录失败页面路径
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
