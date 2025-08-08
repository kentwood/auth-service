package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	DB       DBConfig       `mapstructure:"db"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	OAuth2   OAuth2Config   `mapstructure:"oauth2"`
	UI       UIConfig       `mapstructure:"ui"`       // 新增 UI 配置
	HCaptcha HCaptchaConfig `mapstructure:"hcaptcha"` // 新增 hCaptcha 配置
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

// HCaptchaConfig hCaptcha 配置
type HCaptchaConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	SiteKey   string `mapstructure:"site_key"`
	Enabled   bool   `mapstructure:"enabled"` // 是否启用验证
}

// Load 加载配置文件
func Load(configPath ...string) (*Config, error) {
	var configFile string

	// 1. 优先使用传入的配置文件路径
	if len(configPath) > 0 && configPath[0] != "" {
		configFile = configPath[0]
	} else {
		// 2. 从环境变量获取配置文件路径
		if env := os.Getenv("CONFIG_FILE"); env != "" {
			configFile = env
		} else {
			// 3. 根据环境变量确定配置文件
			env := os.Getenv("APP_ENV")
			switch env {
			case "production", "prod":
				configFile = "configs/config.prod.yaml"
			case "test", "testing":
				configFile = "configs/config.test.yaml"
			case "development", "dev":
				configFile = "configs/config.dev.yaml"
			default:
				// 默认使用开发环境配置
				configFile = "configs/config.dev.yaml"
			}
		}
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configFile)
	}

	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 从环境变量覆盖敏感配置（安全考虑）
	overrideFromEnv(&cfg)

	fmt.Printf("✅ 成功加载配置文件: %s\n", configFile)
	return &cfg, nil
}

// overrideFromEnv 从环境变量覆盖敏感配置
func overrideFromEnv(cfg *Config) {
	// 数据库密码
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		cfg.DB.Password = dbPass
	}

	// Redis 密码
	if redisPass := os.Getenv("REDIS_PASSWORD"); redisPass != "" {
		cfg.Redis.Password = redisPass
	}

	// JWT 密钥
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWT.Secret = jwtSecret
	}

	// GitHub OAuth
	if githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET"); githubClientSecret != "" {
		cfg.OAuth2.GitHub.ClientSecret = githubClientSecret
	}

	// hCaptcha
	if hcaptchaSecret := os.Getenv("HCAPTCHA_SECRET_KEY"); hcaptchaSecret != "" {
		cfg.HCaptcha.SecretKey = hcaptchaSecret
	}
}
