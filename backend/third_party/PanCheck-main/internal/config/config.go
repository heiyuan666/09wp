package config

import (
	"fmt"
	"os"
	"strings"

	"PanCheck/pkg/database"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig    `yaml:"server"`
	Database database.Config `yaml:"database"`
	Checker  CheckerConfig   `yaml:"checker"`
	Redis    RedisConfig     `yaml:"redis"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port        int    `yaml:"port"`
	Mode        string `yaml:"mode"` // debug/release
	CORSOrigins string `yaml:"cors_origins"`
}

// CheckerConfig 检测器配置
type CheckerConfig struct {
	DefaultConcurrency int `yaml:"default_concurrency"` // 默认并发数
	Timeout            int `yaml:"timeout"`             // 超时时间（秒）
}

// PlatformRateConfig 平台频率控制配置
type PlatformRateConfig struct {
	Concurrency          int `json:"concurrency"`             // 并发数
	RequestDelayMs       int `json:"request_delay_ms"`        // 请求间隔（毫秒）
	MaxRequestsPerSecond int `json:"max_requests_per_second"` // 每秒最大请求数（0表示不限制）
	CacheTTLHours        int `json:"cache_ttl_hours"`         // 有效链接缓存过期时间（小时）
}

// RedisConfig Redis配置
type RedisConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`         // 是否启用Redis
	Host       string `yaml:"host" json:"host"`               // Redis地址
	Port       int    `yaml:"port" json:"port"`               // Redis端口
	Username   string `yaml:"username" json:"username"`       // Redis用户名（Redis 6.0+ ACL支持，留空则只使用密码）
	Password   string `yaml:"password" json:"password"`       // Redis密码
	InvalidTTL int    `yaml:"invalid_ttl" json:"invalid_ttl"` // 无效链接统一过期时间（小时）
}

var AppConfig *Config

// Load 加载配置文件
func Load(configPath string) error {
	// 设置默认值
	setDefaults()

	// 如果配置文件存在，先读取配置文件
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		AppConfig = &Config{}
		err = yaml.Unmarshal(data, AppConfig)
		if err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	} else {
		// 配置文件不存在，使用默认配置
		AppConfig = &Config{}
	}

	// 读取环境变量（环境变量会覆盖配置文件的值）
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 从环境变量读取服务器配置
	if port := viper.GetInt("SERVER_PORT"); port > 0 {
		AppConfig.Server.Port = port
	}
	if mode := viper.GetString("SERVER_MODE"); mode != "" {
		AppConfig.Server.Mode = mode
	}
	if corsOrigins := viper.GetString("SERVER_CORS_ORIGINS"); corsOrigins != "" {
		AppConfig.Server.CORSOrigins = corsOrigins
	}

	// 从环境变量读取数据库配置
	if dbType := viper.GetString("DATABASE_TYPE"); dbType != "" {
		AppConfig.Database.Type = dbType
	}
	if dbHost := viper.GetString("DATABASE_HOST"); dbHost != "" {
		AppConfig.Database.Host = dbHost
	}
	if dbPort := viper.GetInt("DATABASE_PORT"); dbPort > 0 {
		AppConfig.Database.Port = dbPort
	}
	if dbUser := viper.GetString("DATABASE_USER"); dbUser != "" {
		AppConfig.Database.User = dbUser
	}
	if dbPassword := viper.GetString("DATABASE_PASSWORD"); dbPassword != "" {
		AppConfig.Database.Password = dbPassword
	}
	if dbDatabase := viper.GetString("DATABASE_DATABASE"); dbDatabase != "" {
		AppConfig.Database.Database = dbDatabase
	}
	if dbCharset := viper.GetString("DATABASE_CHARSET"); dbCharset != "" {
		AppConfig.Database.Charset = dbCharset
	}

	// 从环境变量读取检测器配置
	if concurrency := viper.GetInt("CHECKER_DEFAULT_CONCURRENCY"); concurrency > 0 {
		AppConfig.Checker.DefaultConcurrency = concurrency
	}
	if timeout := viper.GetInt("CHECKER_TIMEOUT"); timeout > 0 {
		AppConfig.Checker.Timeout = timeout
	}

	// 从环境变量读取Redis配置
	if enabled := viper.GetBool("REDIS_ENABLED"); enabled {
		AppConfig.Redis.Enabled = enabled
	}
	if redisHost := viper.GetString("REDIS_HOST"); redisHost != "" {
		AppConfig.Redis.Host = redisHost
	}
	if redisPort := viper.GetInt("REDIS_PORT"); redisPort > 0 {
		AppConfig.Redis.Port = redisPort
	}
	if redisUsername := viper.GetString("REDIS_USERNAME"); redisUsername != "" {
		AppConfig.Redis.Username = redisUsername
	}
	if redisPassword := viper.GetString("REDIS_PASSWORD"); redisPassword != "" {
		AppConfig.Redis.Password = redisPassword
	}
	if invalidTTL := viper.GetInt("REDIS_INVALID_TTL"); invalidTTL > 0 {
		AppConfig.Redis.InvalidTTL = invalidTTL
	}

	// 设置默认值（如果仍然为空）
	if AppConfig.Server.Port == 0 {
		AppConfig.Server.Port = 6080
	}
	if AppConfig.Server.Mode == "" {
		AppConfig.Server.Mode = "debug"
	}
	if AppConfig.Server.CORSOrigins == "" {
		AppConfig.Server.CORSOrigins = "*"
	}
	if AppConfig.Checker.DefaultConcurrency == 0 {
		AppConfig.Checker.DefaultConcurrency = 5
	}
	if AppConfig.Checker.Timeout == 0 {
		AppConfig.Checker.Timeout = 30
	}
	if AppConfig.Redis.Host == "" {
		AppConfig.Redis.Host = "localhost"
	}
	if AppConfig.Redis.Port == 0 {
		AppConfig.Redis.Port = 6379
	}
	if AppConfig.Redis.InvalidTTL == 0 {
		AppConfig.Redis.InvalidTTL = 168 // 默认7天
	}

	return nil
}

// setDefaults 设置默认配置
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("SERVER_PORT", 6080)
	viper.SetDefault("SERVER_MODE", "debug")
	viper.SetDefault("SERVER_CORS_ORIGINS", "*")

	// 数据库默认配置
	viper.SetDefault("DATABASE_TYPE", "mysql")
	viper.SetDefault("DATABASE_HOST", "localhost")
	viper.SetDefault("DATABASE_PORT", 3306)
	viper.SetDefault("DATABASE_USER", "root")
	viper.SetDefault("DATABASE_PASSWORD", "")
	viper.SetDefault("DATABASE_DATABASE", "pancheck")
	viper.SetDefault("DATABASE_CHARSET", "utf8mb4")

	// 检测器默认配置
	viper.SetDefault("CHECKER_DEFAULT_CONCURRENCY", 5)
	viper.SetDefault("CHECKER_TIMEOUT", 30)

	// Redis默认配置
	viper.SetDefault("REDIS_ENABLED", false)
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_USERNAME", "")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_INVALID_TTL", 168) // 默认7天
}
