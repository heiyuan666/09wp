package handler

import (
	"PanCheck/internal/config"
	"PanCheck/internal/model"
	"PanCheck/internal/repository"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SettingsHandler 设置处理器
type SettingsHandler struct {
	settingsRepo   *repository.SettingsRepository
	checkerFactory interface {
		GetAllCheckers() map[string]interface{}
	}
	reloadCheckerFunc func() error
}

// NewSettingsHandler 创建设置处理器
func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{
		settingsRepo: repository.NewSettingsRepository(),
	}
}

// SetReloadCheckerFunc 设置重新加载检测器的函数
func (h *SettingsHandler) SetReloadCheckerFunc(fn func() error) {
	h.reloadCheckerFunc = fn
}

// GetSettings 获取所有设置
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	category := c.Query("category")

	var settings []model.Setting
	var err error

	if category != "" {
		settings, err = h.settingsRepo.GetByCategory(category)
	} else {
		settings, err = h.settingsRepo.GetAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// GetSetting 获取单个设置
func (h *SettingsHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")

	setting, err := h.settingsRepo.GetByKey(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// UpdateSetting 更新设置
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	var req struct {
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.settingsRepo.GetByKey(key)
	if err != nil {
		// 不存在则创建
		setting = &model.Setting{
			Key:         key,
			Value:       req.Value,
			Description: req.Description,
			Category:    "checker",
		}
		err = h.settingsRepo.Update(setting)
	} else {
		setting.Value = req.Value
		if req.Description != "" {
			setting.Description = req.Description
		}
		err = h.settingsRepo.Update(setting)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// GetRateConfigSettings 获取所有平台的频率配置
func (h *SettingsHandler) GetRateConfigSettings(c *gin.Context) {
	platforms := []string{"quark", "uc", "baidu", "tianyi", "pan123", "pan115", "aliyun", "xunlei", "cmcc"}

	result := make(map[string]*config.PlatformRateConfig)
	defaultConfig := &config.PlatformRateConfig{
		Concurrency:          5,
		RequestDelayMs:       0,
		MaxRequestsPerSecond: 0,
		CacheTTLHours:        24, // 默认24小时
	}

	for _, platform := range platforms {
		key := "platform_rate_config_" + platform
		setting, err := h.settingsRepo.GetByKey(key)
		if err == nil && setting != nil {
			var rateConfig config.PlatformRateConfig
			if err := json.Unmarshal([]byte(setting.Value), &rateConfig); err == nil {
				result[platform] = &rateConfig
				continue
			}
		}
		// 使用默认配置
		result[platform] = defaultConfig
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// UpdateRateConfigSettings 批量更新平台频率配置
func (h *SettingsHandler) UpdateRateConfigSettings(c *gin.Context) {
	var req struct {
		Settings map[string]*config.PlatformRateConfig `json:"settings" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证配置范围
	for platform, rateConfig := range req.Settings {
		if rateConfig.Concurrency < 1 || rateConfig.Concurrency > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": platform + " 的并发数必须在1-100之间",
			})
			return
		}
		if rateConfig.RequestDelayMs < 0 || rateConfig.RequestDelayMs > 10000 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": platform + " 的请求间隔必须在0-10000毫秒之间",
			})
			return
		}
		if rateConfig.MaxRequestsPerSecond < 0 || rateConfig.MaxRequestsPerSecond > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": platform + " 的每秒最大请求数必须在0-100之间",
			})
			return
		}
		if rateConfig.CacheTTLHours < 0 || rateConfig.CacheTTLHours > 720 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": platform + " 的缓存有效期必须在0-720小时之间",
			})
			return
		}
	}

	// 批量更新
	updated := make(map[string]*config.PlatformRateConfig)
	for platform, rateConfig := range req.Settings {
		key := "platform_rate_config_" + platform

		// 序列化为JSON
		valueBytes, err := json.Marshal(rateConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "序列化 " + platform + " 配置失败: " + err.Error(),
			})
			return
		}

		setting, err := h.settingsRepo.GetByKey(key)
		if err != nil {
			// 不存在则创建
			setting = &model.Setting{
				Key:         key,
				Value:       string(valueBytes),
				Description: platform + "网盘频率控制配置",
				Category:    "checker",
			}
			err = h.settingsRepo.Update(setting)
		} else {
			setting.Value = string(valueBytes)
			err = h.settingsRepo.Update(setting)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "更新 " + platform + " 设置失败: " + err.Error(),
			})
			return
		}

		updated[platform] = rateConfig
	}

	// 重新加载检测器配置，使配置立即生效
	if h.reloadCheckerFunc != nil {
		if err := h.reloadCheckerFunc(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "配置已保存，但重新加载检测器失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "保存成功，配置已立即生效",
		"data":    updated,
	})
}

// GetRedisConfig 获取Redis配置
func (h *SettingsHandler) GetRedisConfig(c *gin.Context) {
	setting, err := h.settingsRepo.GetByKey("redis_config")
	if err != nil || setting == nil {
		// 返回默认配置
		defaultConfig := config.RedisConfig{
			Enabled:    false,
			Host:       "localhost",
			Port:       6379,
			Username:   "",
			Password:   "",
			InvalidTTL: 168, // 默认7天
		}
		c.JSON(http.StatusOK, gin.H{"data": defaultConfig})
		return
	}

	var redisConfig config.RedisConfig
	if err := json.Unmarshal([]byte(setting.Value), &redisConfig); err != nil {
		// 解析失败，返回默认配置
		defaultConfig := config.RedisConfig{
			Enabled:    false,
			Host:       "localhost",
			Port:       6379,
			Username:   "",
			Password:   "",
			InvalidTTL: 168,
		}
		c.JSON(http.StatusOK, gin.H{"data": defaultConfig})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": redisConfig})
}

// UpdateRedisConfig 更新Redis配置
func (h *SettingsHandler) UpdateRedisConfig(c *gin.Context) {
	var req struct {
		Config config.RedisConfig `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证配置
	if req.Config.Port < 1 || req.Config.Port > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Redis端口必须在1-65535之间",
		})
		return
	}
	if req.Config.InvalidTTL < 1 || req.Config.InvalidTTL > 720 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效链接缓存过期时间必须在1-720小时之间",
		})
		return
	}

	// 序列化为JSON
	valueBytes, err := json.Marshal(req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "序列化Redis配置失败: " + err.Error(),
		})
		return
	}

	setting, err := h.settingsRepo.GetByKey("redis_config")
	if err != nil {
		// 不存在则创建
		setting = &model.Setting{
			Key:         "redis_config",
			Value:       string(valueBytes),
			Description: "Redis缓存配置",
			Category:    "cache",
		}
		err = h.settingsRepo.Update(setting)
	} else {
		setting.Value = string(valueBytes)
		err = h.settingsRepo.Update(setting)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新Redis配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "保存成功，请重启服务使配置生效",
		"data":    req.Config,
	})
}
