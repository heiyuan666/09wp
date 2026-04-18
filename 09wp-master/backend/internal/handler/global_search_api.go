package handler

import (
	"context"
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AdminGlobalSearchSettingsGet 全网搜站点级配置（与「接口线路」表分离，避免误用整站 PUT 覆盖）
func AdminGlobalSearchSettingsGet(c *gin.Context) {
	cfg := readGlobalSearchConfig()
	response.OK(c, gin.H{
		"global_search_enabled":                      cfg.GlobalSearchEnabled,
		"global_search_link_check_enabled":           cfg.GlobalSearchLinkCheckEnabled,
		"global_search_api_url":                      strings.TrimSpace(cfg.GlobalSearchAPIURL),
		"global_search_cloud_types":                  strings.TrimSpace(cfg.GlobalSearchCloudTypes),
		"global_search_default_category_id":          cfg.GlobalSearchDefaultCategoryID,
		"global_search_auto_transfer":                cfg.GlobalSearchAutoTransfer,
		"global_search_cleanup_enabled":              cfg.GlobalSearchCleanupEnabled,
		"global_search_cleanup_days":                 cfg.GlobalSearchCleanupDays,
		"global_search_cleanup_minutes":              cfg.GlobalSearchCleanupMinutes,
		"global_search_cleanup_delete_netdisk_files": cfg.GlobalSearchCleanupDeleteNetdiskFiles,
		"global_search_url_sanitize_regex":           strings.TrimSpace(cfg.GlobalSearchURLSanitizeRegex),
	})
}

type globalSearchSettingsPut struct {
	GlobalSearchEnabled                   bool   `json:"global_search_enabled"`
	GlobalSearchLinkCheckEnabled          bool   `json:"global_search_link_check_enabled"`
	GlobalSearchAPIURL                    string `json:"global_search_api_url"`
	GlobalSearchCloudTypes                string `json:"global_search_cloud_types"`
	GlobalSearchDefaultCategoryID         uint64 `json:"global_search_default_category_id"`
	GlobalSearchAutoTransfer              bool   `json:"global_search_auto_transfer"`
	GlobalSearchCleanupEnabled            bool   `json:"global_search_cleanup_enabled"`
	GlobalSearchCleanupDays               int    `json:"global_search_cleanup_days"`
	GlobalSearchCleanupMinutes            int    `json:"global_search_cleanup_minutes"`
	GlobalSearchCleanupDeleteNetdiskFiles bool   `json:"global_search_cleanup_delete_netdisk_files"`
	GlobalSearchURLSanitizeRegex          string `json:"global_search_url_sanitize_regex"`
}

// AdminGlobalSearchSettingsPut 仅更新全网搜相关列
func AdminGlobalSearchSettingsPut(c *gin.Context) {
	var req globalSearchSettingsPut
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		response.Error(c, 404, "系统配置不存在")
		return
	}
	updates := map[string]interface{}{
		"global_search_enabled":                      req.GlobalSearchEnabled,
		"global_search_link_check_enabled":           req.GlobalSearchLinkCheckEnabled,
		"global_search_api_url":                      strings.TrimSpace(req.GlobalSearchAPIURL),
		"global_search_cloud_types":                  strings.TrimSpace(req.GlobalSearchCloudTypes),
		"global_search_default_category_id":          req.GlobalSearchDefaultCategoryID,
		"global_search_auto_transfer":                req.GlobalSearchAutoTransfer,
		"global_search_cleanup_enabled":              req.GlobalSearchCleanupEnabled,
		"global_search_cleanup_days":                 req.GlobalSearchCleanupDays,
		"global_search_cleanup_minutes":              req.GlobalSearchCleanupMinutes,
		"global_search_cleanup_delete_netdisk_files": req.GlobalSearchCleanupDeleteNetdiskFiles,
		"global_search_url_sanitize_regex":           strings.TrimSpace(req.GlobalSearchURLSanitizeRegex),
	}
	if err := database.DB().Model(&cfg).Updates(updates).Error; err != nil {
		response.Error(c, 500, "保存失败")
		return
	}
	service.DeleteSearchCache(context.Background(), "public:system-config:v3")
	service.DeleteSearchCache(context.Background(), "public:system-config:v4")
	service.DeleteSearchCache(context.Background(), "public:system-config:v5")
	// 全网搜结果缓存键含线路指纹，此处无法按前缀删；依赖 TTL 或用户改线路后新指纹自然未命中。
	response.OK(c, nil)
}

type globalSearchAPIReq struct {
	Name       string `json:"name" binding:"required"`
	APIURL     string `json:"api_url" binding:"required"`
	CloudTypes string `json:"cloud_types"`
	Enabled    *bool  `json:"enabled"`
	SortOrder  int    `json:"sort_order"`
}

func AdminGlobalSearchAPIList(c *gin.Context) {
	var list []model.GlobalSearchAPI
	if err := database.DB().Order("sort_order DESC, id ASC").Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OKPage(c, list, int64(len(list)))
}

func AdminGlobalSearchAPICreate(c *gin.Context) {
	var req globalSearchAPIReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	name := strings.TrimSpace(req.Name)
	apiURL := strings.TrimSpace(req.APIURL)
	if name == "" || apiURL == "" {
		response.Error(c, 400, "名称和接口地址不能为空")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	row := model.GlobalSearchAPI{
		Name:       name,
		APIURL:     apiURL,
		CloudTypes: strings.TrimSpace(req.CloudTypes),
		Enabled:    enabled,
		SortOrder:  req.SortOrder,
	}
	if err := database.DB().Create(&row).Error; err != nil {
		response.Error(c, 500, "创建失败")
		return
	}
	response.OK(c, row)
}

func AdminGlobalSearchAPIUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var req globalSearchAPIReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	name := strings.TrimSpace(req.Name)
	apiURL := strings.TrimSpace(req.APIURL)
	if name == "" || apiURL == "" {
		response.Error(c, 400, "名称和接口地址不能为空")
		return
	}
	var row model.GlobalSearchAPI
	if err := database.DB().First(&row, id).Error; err != nil {
		response.Error(c, 404, "记录不存在")
		return
	}
	enabled := row.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if err := database.DB().Model(&row).Updates(map[string]any{
		"name":        name,
		"api_url":     apiURL,
		"cloud_types": strings.TrimSpace(req.CloudTypes),
		"enabled":     enabled,
		"sort_order":  req.SortOrder,
	}).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

func AdminGlobalSearchAPIDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	if err := database.DB().Delete(&model.GlobalSearchAPI{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}
