package handler

import (
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AdminGameFeedbackList 管理端：游戏资源反馈列表
func AdminGameFeedbackList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	status := strings.TrimSpace(c.Query("status"))
	typeFilter := strings.TrimSpace(c.Query("type"))
	gameID := strings.TrimSpace(c.Query("game_id"))

	db := database.DB().Model(&model.GameResourceFeedback{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if typeFilter != "" {
		db = db.Where("type = ?", typeFilter)
	}
	if gameID != "" {
		if gid, err := strconv.ParseUint(gameID, 10, 64); err == nil && gid > 0 {
			db = db.Where("game_id = ?", gid)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "统计失败")
		return
	}

	var list []model.GameResourceFeedback
	if err := db.Order("created_at DESC, id DESC").
		Limit(pageSize).Offset((page-1)*pageSize).
		Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	response.OKPage(c, list, total)
}

type adminGameFeedbackStatusUpdateReq struct {
	Status string `json:"status" binding:"required"`
}

// AdminGameFeedbackUpdateStatus 管理端：更新游戏反馈状态
func AdminGameFeedbackUpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "反馈 ID 不合法")
		return
	}
	var req adminGameFeedbackStatusUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	req.Status = strings.TrimSpace(req.Status)
	if req.Status != "pending" && req.Status != "processed" {
		response.Error(c, 400, "status 不合法")
		return
	}

	tx := database.DB().Model(&model.GameResourceFeedback{}).Where("id = ?", id).Updates(map[string]any{
		"status":     req.Status,
		"updated_at": time.Now(),
	})
	if tx.Error != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	if tx.RowsAffected == 0 {
		response.Error(c, 404, "反馈不存在")
		return
	}
	response.OK(c, gin.H{"id": id, "status": req.Status})
}

