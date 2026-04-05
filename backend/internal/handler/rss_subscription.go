package handler

import (
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type rssSubscriptionReq struct {
	Name         string `json:"name" binding:"required"`
	FeedURL      string `json:"feed_url" binding:"required"`
	DefaultCatID uint64 `json:"default_cat_id"`
	Enabled      bool   `json:"enabled"`
	SyncInterval int    `json:"sync_interval"`
	MaxItems     int    `json:"max_items"`
}

// RSSSubscriptionList RSS 订阅列表
func RSSSubscriptionList(c *gin.Context) {
	var list []model.RSSSubscription
	if err := database.DB().Order("id DESC").Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, gin.H{
		"list":  list,
		"total": len(list),
	})
}

// RSSSubscriptionCreate 新增 RSS 订阅
func RSSSubscriptionCreate(c *gin.Context) {
	var req rssSubscriptionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.FeedURL = strings.TrimSpace(req.FeedURL)
	if req.SyncInterval < 60 {
		req.SyncInterval = 1800
	}
	if req.MaxItems <= 0 {
		req.MaxItems = 50
	}
	if req.MaxItems > 200 {
		req.MaxItems = 200
	}

	row := model.RSSSubscription{
		Name:         req.Name,
		FeedURL:      req.FeedURL,
		DefaultCatID: req.DefaultCatID,
		Enabled:      req.Enabled,
		SyncInterval: req.SyncInterval,
		MaxItems:     req.MaxItems,
	}

	if err := database.DB().Create(&row).Error; err != nil {
		response.Error(c, 500, "创建失败: "+err.Error())
		return
	}
	response.OK(c, row)
}

// RSSSubscriptionUpdate 编辑 RSS 订阅
func RSSSubscriptionUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var req rssSubscriptionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.FeedURL = strings.TrimSpace(req.FeedURL)
	if req.SyncInterval < 60 {
		req.SyncInterval = 1800
	}
	if req.MaxItems <= 0 {
		req.MaxItems = 50
	}
	if req.MaxItems > 200 {
		req.MaxItems = 200
	}

	updates := map[string]interface{}{
		"name":           req.Name,
		"feed_url":       req.FeedURL,
		"default_cat_id": req.DefaultCatID,
		"enabled":        req.Enabled,
		"sync_interval":  req.SyncInterval,
		"max_items":      req.MaxItems,
	}
	if err := database.DB().Model(&model.RSSSubscription{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败: "+err.Error())
		return
	}
	response.OK(c, nil)
}

// RSSSubscriptionDelete 删除 RSS 订阅
func RSSSubscriptionDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	if err := database.DB().Delete(&model.RSSSubscription{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

// RSSSubscriptionTest 测试 RSS 地址
func RSSSubscriptionTest(c *gin.Context) {
	var req struct {
		FeedURL string `json:"feed_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := service.TestRSSFeed(strings.TrimSpace(req.FeedURL)); err != nil {
		response.Error(c, 500, "测试失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true, "message": "RSS 测试成功"})
}

// RSSSubscriptionSync 手动同步一个 RSS 订阅
func RSSSubscriptionSync(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	added, skipped, err := service.SyncRSSSubscriptionByID(id)
	if err != nil {
		response.Error(c, 500, "同步失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"added": added, "skipped": skipped})
}

// RSSSubscriptionSyncAll 手动同步所有已启用 RSS 订阅
func RSSSubscriptionSyncAll(c *gin.Context) {
	synced, added, skipped, err := service.SyncAllEnabledRSSSubscriptions()
	if err != nil {
		response.Error(c, 500, "同步失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"synced": synced, "added": added, "skipped": skipped})
}
