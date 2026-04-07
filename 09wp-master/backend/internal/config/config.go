package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DefaultPanCheckBaseURL 失效检测服务（PanCheck）默认地址，与后台「盘查服务地址」一致。
const DefaultPanCheckBaseURL = "https://pancheck.116818.xyz"

// Config 基础配置结构体
type Config struct {
	HTTPPort        string
	MySQLDSN        string
	JWTSecret       string
	PanCheckBaseURL string
	Redis           RedisConfig
	Meili           MeiliConfig
}

// MeiliConfig Meilisearch 配置（可选）
type MeiliConfig struct {
	Enabled    bool
	URL        string
	APIKey     string
	Index      string
	TimeoutMS  int
	PrimaryKey string
}

func (c MeiliConfig) IsZero() bool {
	return !c.Enabled &&
		strings.TrimSpace(c.URL) == "" &&
		strings.TrimSpace(c.APIKey) == "" &&
		strings.TrimSpace(c.Index) == "" &&
		c.TimeoutMS == 0 &&
		strings.TrimSpace(c.PrimaryKey) == ""
}

// RedisConfig Redis 缓存配置
type RedisConfig struct {
	Enabled      bool
	Host         string
	Port         int
	Username     string
	Password     string
	SearchTTL    int // 搜索缓存 TTL（秒）
	PingTimeout  int // ping 超时（秒）
	ConnectTTLMS int // 连接超时（毫秒）
}

type fileConfig struct {
	HTTPPort        *string          `json:"http_port"`
	MySQLDSN        *string          `json:"mysql_dsn"`
	JWTSecret       *string          `json:"jwt_secret"`
	PanCheckBaseURL *string          `json:"pancheck_base_url"`
	Redis           *fileRedisConfig `json:"redis"`
	Meili           *fileMeiliConfig `json:"meilisearch"`
}

type fileRedisConfig struct {
	Enabled      *bool   `json:"enabled"`
	Host         *string `json:"host"`
	Port         *int    `json:"port"`
	Username     *string `json:"username"`
	Password     *string `json:"password"`
	SearchTTL    *int    `json:"search_cache_ttl"`   // 秒
	PingTimeout  *int    `json:"ping_timeout"`       // 秒
	ConnectTTLMS *int    `json:"connect_timeout_ms"` // 毫秒
}

type fileMeiliConfig struct {
	Enabled    *bool   `json:"enabled"`
	URL        *string `json:"url"`
	APIKey     *string `json:"api_key"`
	Index      *string `json:"index"`
	TimeoutMS  *int    `json:"timeout_ms"`
	PrimaryKey *string `json:"primary_key"`
}

func applyFileConfig(cfg *Config, fc *fileConfig) {
	if fc == nil {
		return
	}
	if fc.HTTPPort != nil {
		cfg.HTTPPort = *fc.HTTPPort
	}
	if fc.MySQLDSN != nil {
		cfg.MySQLDSN = *fc.MySQLDSN
	}
	if fc.JWTSecret != nil {
		cfg.JWTSecret = *fc.JWTSecret
	}
	if fc.PanCheckBaseURL != nil {
		cfg.PanCheckBaseURL = *fc.PanCheckBaseURL
	}
	if fc.Redis == nil {
		// continue
	} else {
		rc := fc.Redis
		if rc.Enabled != nil {
			cfg.Redis.Enabled = *rc.Enabled
		}
		if rc.Host != nil {
			cfg.Redis.Host = *rc.Host
		}
		if rc.Port != nil {
			cfg.Redis.Port = *rc.Port
		}
		if rc.Username != nil {
			cfg.Redis.Username = *rc.Username
		}
		if rc.Password != nil {
			cfg.Redis.Password = *rc.Password
		}
		if rc.SearchTTL != nil {
			cfg.Redis.SearchTTL = *rc.SearchTTL
		}
		if rc.PingTimeout != nil {
			cfg.Redis.PingTimeout = *rc.PingTimeout
		}
		if rc.ConnectTTLMS != nil {
			cfg.Redis.ConnectTTLMS = *rc.ConnectTTLMS
		}
	}

	if fc.Meili == nil {
		return
	}
	mc := fc.Meili
	if mc.Enabled != nil {
		cfg.Meili.Enabled = *mc.Enabled
	}
	if mc.URL != nil {
		cfg.Meili.URL = *mc.URL
	}
	if mc.APIKey != nil {
		cfg.Meili.APIKey = *mc.APIKey
	}
	if mc.Index != nil {
		cfg.Meili.Index = *mc.Index
	}
	if mc.TimeoutMS != nil {
		cfg.Meili.TimeoutMS = *mc.TimeoutMS
	}
	if mc.PrimaryKey != nil {
		cfg.Meili.PrimaryKey = *mc.PrimaryKey
	}
}

func loadFileConfig() *fileConfig {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath != "" {
		b, err := os.ReadFile(configPath)
		if err != nil {
			log.Printf("load config file failed: %v", err)
			return nil
		}
		var fc fileConfig
		if err := json.Unmarshal(b, &fc); err != nil {
			log.Printf("parse config file failed: %v", err)
			return nil
		}
		return &fc
	}

	// 默认尝试：backend 目录下的 config.json / configs/config.json
	// 注意：工作目录不同会影响相对路径，因此也尽量用 filepath.Join 做兼容。
	cwd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(cwd, "config.json"),
		filepath.Join(cwd, "configs", "config.json"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			log.Printf("load config file failed: %s: %v", p, err)
			continue
		}
		var fc fileConfig
		if err := json.Unmarshal(b, &fc); err != nil {
			log.Printf("parse config file failed: %s: %v", p, err)
			continue
		}
		return &fc
	}
	return nil
}

// Load 从环境变量加载配置，提供合理默认值，方便本地快速启动
func Load() Config {
	cfg := Config{
		HTTPPort:        "8080",
		MySQLDSN:        "pan:pan@tcp(111.170.19.100:3306/pan?charset=utf8mb4&parseTime=True&loc=Local",
		JWTSecret:       "dfan-netdisk-dev-secret",
		PanCheckBaseURL: DefaultPanCheckBaseURL,
		Redis: RedisConfig{
			Enabled:      false,
			Host:         "127.0.0.1",
			Port:         6379,
			Username:     "",
			Password:     "",
			SearchTTL:    60,
			PingTimeout:  3,
			ConnectTTLMS: 2000,
		},
		Meili: MeiliConfig{
			Enabled:    false,
			URL:        "http://127.0.0.1:7700",
			APIKey:     "",
			Index:      "resources",
			TimeoutMS:  2500,
			PrimaryKey: "id",
		},
	}

	// 1) 优先读取 config.json（可选）
	applyFileConfig(&cfg, loadFileConfig())

	// 2) 再用环境变量覆盖（兼容旧启动方式）
	if v := os.Getenv("HTTP_PORT"); v != "" {
		cfg.HTTPPort = v
	}
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		cfg.MySQLDSN = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}
	if v := os.Getenv("PANCHECK_BASE_URL"); v != "" {
		cfg.PanCheckBaseURL = v
	}

	if v := os.Getenv("REDIS_ENABLED"); v != "" {
		// 允许 true/1/yes
		cfg.Redis.Enabled = v == "1" || v == "true" || v == "yes"
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Redis.Port = n
		}
	}
	if v := os.Getenv("REDIS_USERNAME"); v != "" {
		cfg.Redis.Username = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("REDIS_SEARCH_CACHE_TTL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Redis.SearchTTL = n
		}
	}
	if v := os.Getenv("REDIS_PING_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Redis.PingTimeout = n
		}
	}
	// connect timeout（毫秒）主要给网络超时兜底
	if v := os.Getenv("REDIS_CONNECT_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Redis.ConnectTTLMS = n
		}
	}

	if v := os.Getenv("MEILI_ENABLED"); v != "" {
		cfg.Meili.Enabled = v == "1" || v == "true" || v == "yes" || v == "on"
	}
	if v := os.Getenv("MEILI_URL"); v != "" {
		cfg.Meili.URL = v
	}
	if v := os.Getenv("MEILI_API_KEY"); v != "" {
		cfg.Meili.APIKey = v
	}
	if v := os.Getenv("MEILI_INDEX"); v != "" {
		cfg.Meili.Index = v
	}
	if v := os.Getenv("MEILI_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Meili.TimeoutMS = n
		}
	}
	if v := os.Getenv("MEILI_PRIMARY_KEY"); v != "" {
		cfg.Meili.PrimaryKey = v
	}

	return cfg
}
