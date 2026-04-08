package handler

import (
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type gameSiteConfigPut struct {
	SiteTitle      string `json:"site_title"`
	LogoURL        string `json:"logo_url"`
	FaviconURL     string `json:"favicon_url"`
	SeoKeywords    string `json:"seo_keywords"`
	SeoDescription string `json:"seo_description"`
}

func ensureGameSiteConfigRow() (model.GameSiteConfig, error) {
	var cfg model.GameSiteConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil {
		return cfg, nil
	}
	cfg = model.GameSiteConfig{
		SiteTitle: "游戏资源站",
	}
	if err := database.DB().Create(&cfg).Error; err != nil {
		return model.GameSiteConfig{}, err
	}
	return cfg, nil
}

// GetGameSiteConfig 管理端获取游戏站点配置（需管理员）
func GetGameSiteConfig(c *gin.Context) {
	cfg, err := ensureGameSiteConfigRow()
	if err != nil {
		response.Error(c, 500, "系统配置读取失败")
		return
	}
	response.OK(c, cfg)
}

// UpdateGameSiteConfig 管理端更新游戏站点配置（需管理员）
func UpdateGameSiteConfig(c *gin.Context) {
	var req gameSiteConfigPut
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	cfg, err := ensureGameSiteConfigRow()
	if err != nil {
		response.Error(c, 500, "系统配置读取失败")
		return
	}

	uidVal, _ := c.Get("user_id")
	uid, _ := uidVal.(uint64)

	updates := map[string]interface{}{
		"site_title":      strings.TrimSpace(req.SiteTitle),
		"logo_url":        strings.TrimSpace(req.LogoURL),
		"favicon_url":     strings.TrimSpace(req.FaviconURL),
		"seo_keywords":    strings.TrimSpace(req.SeoKeywords),
		"seo_description": strings.TrimSpace(req.SeoDescription),
		"updated_by":      uid,
	}
	if err := database.DB().Model(&cfg).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// GetPublicGameSiteConfig 游戏前台获取站点配置（无需登录，字段精简）
func GetPublicGameSiteConfig(c *gin.Context) {
	cfg, err := ensureGameSiteConfigRow()
	if err != nil {
		response.Error(c, 500, "系统配置读取失败")
		return
	}
	response.OK(c, gin.H{
		"site_title":      cfg.SiteTitle,
		"logo_url":        cfg.LogoURL,
		"favicon_url":     cfg.FaviconURL,
		"seo_keywords":    cfg.SeoKeywords,
		"seo_description": cfg.SeoDescription,
	})
}

