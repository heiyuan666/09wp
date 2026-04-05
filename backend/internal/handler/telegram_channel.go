package handler

import (
	"context"
	"errors"
	"strconv"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type tgChannelReq struct {
	Name          string `json:"name"`
	BotToken      string `json:"bot_token"`
	ChannelChatID string `json:"channel_chat_id"`
	ProxyURL      string `json:"proxy_url"`
	DefaultCatID  uint64 `json:"default_cat_id"`
	Enabled       bool   `json:"enabled"`
	SyncInterval  int    `json:"sync_interval"`
}

type tgChannelTestReq struct {
	BotToken      string `json:"bot_token"`
	ChannelChatID string `json:"channel_chat_id"`
	ProxyURL      string `json:"proxy_url"`
}

type tgChannelBackfillReq struct {
	Limit int `json:"limit"`
}

func TelegramChannelList(c *gin.Context) {
	var list []model.TelegramChannel
	if err := database.DB().Order("id DESC").Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func TelegramChannelCreate(c *gin.Context) {
	var req tgChannelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if req.Name == "" || req.ChannelChatID == "" {
		response.Error(c, 400, "名称、频道Chat ID必填")
		return
	}
	if req.SyncInterval < 30 {
		req.SyncInterval = 300
	}
	item := model.TelegramChannel{
		Name:          req.Name,
		BotToken:      req.BotToken,
		ChannelChatID: req.ChannelChatID,
		ProxyURL:      req.ProxyURL,
		DefaultCatID:  req.DefaultCatID,
		Enabled:       req.Enabled,
		SyncInterval:  req.SyncInterval,
	}
	if err := database.DB().Create(&item).Error; err != nil {
		response.Error(c, 500, "创建失败")
		return
	}
	response.OK(c, item)
}

func TelegramChannelUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "无效ID")
		return
	}
	var req tgChannelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if req.Name == "" || req.ChannelChatID == "" {
		response.Error(c, 400, "名称、频道Chat ID必填")
		return
	}
	if req.SyncInterval < 30 {
		req.SyncInterval = 300
	}

	updates := map[string]interface{}{
		"name":            req.Name,
		"bot_token":       req.BotToken,
		"channel_chat_id": req.ChannelChatID,
		"proxy_url":       req.ProxyURL,
		"default_cat_id":  req.DefaultCatID,
		"enabled":         req.Enabled,
		"sync_interval":   req.SyncInterval,
	}
	if err := database.DB().Model(&model.TelegramChannel{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

func TelegramChannelDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "无效ID")
		return
	}
	if err := database.DB().Delete(&model.TelegramChannel{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

func TelegramChannelSync(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "无效ID")
		return
	}
	added, skipped, err := service.SyncTelegramChannelByID(id)
	if err != nil {
		response.Error(c, 500, "同步失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"added": added, "skipped": skipped})
}

func TelegramChannelSyncAll(c *gin.Context) {
	synced, added, skipped, err := service.SyncAllEnabledTelegramChannels()
	if err != nil {
		response.Error(c, 500, "同步失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"synced": synced, "added": added, "skipped": skipped})
}

func TelegramChannelBackfill(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "无效ID")
		return
	}
	var req tgChannelBackfillReq
	_ = c.ShouldBindJSON(&req)
	added, skipped, scanned, err := service.BackfillTelegramChannelByID(id, req.Limit)
	if err != nil {
		response.Error(c, 500, "回溯同步失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"added": added, "skipped": skipped, "scanned": scanned})
}

func TelegramChannelTest(c *gin.Context) {
	var req tgChannelTestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := service.TestTelegramChannelConfig(req.BotToken, req.ChannelChatID, req.ProxyURL); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			response.Error(c, 400, "测试失败: 连接 Telegram 超时，请检查 tg_proxy_url 是否可达（建议 socks5://ip:port）")
			return
		}
		response.Error(c, 400, "测试失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true, "message": "连接测试成功"})
}

