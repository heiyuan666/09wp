package handler

import (
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type softwareSiteConfigPut struct {
	SiteTitle      string `json:"site_title"`
	LogoURL        string `json:"logo_url"`
	FaviconURL     string `json:"favicon_url"`
	SeoKeywords    string `json:"seo_keywords"`
	SeoDescription string `json:"seo_description"`
}

func ensureSoftwareSiteConfigRow() (model.SoftwareSiteConfig, error) {
	var cfg model.SoftwareSiteConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil {
		return cfg, nil
	}
	cfg = model.SoftwareSiteConfig{
		SiteTitle:      "软件库",
		SeoKeywords:    "软件,下载,工具",
		SeoDescription: "精选多平台软件资源与下载指引",
	}
	if err := database.DB().Create(&cfg).Error; err != nil {
		return model.SoftwareSiteConfig{}, err
	}
	return cfg, nil
}

// GetSoftwareSiteConfig 管理端获取软件库站点配置（需管理员）
func GetSoftwareSiteConfig(c *gin.Context) {
	cfg, err := ensureSoftwareSiteConfigRow()
	if err != nil {
		response.Error(c, 500, "系统配置读取失败")
		return
	}
	response.OK(c, cfg)
}

// UpdateSoftwareSiteConfig 管理端更新软件库站点配置（需管理员）
func UpdateSoftwareSiteConfig(c *gin.Context) {
	var req softwareSiteConfigPut
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	cfg, err := ensureSoftwareSiteConfigRow()
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

// GetPublicSoftwareSiteConfig 软件库前台获取站点配置（无需登录）
func GetPublicSoftwareSiteConfig(c *gin.Context) {
	cfg, err := ensureSoftwareSiteConfigRow()
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
