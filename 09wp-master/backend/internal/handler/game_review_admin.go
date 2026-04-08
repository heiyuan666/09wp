package handler

import (
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AdminGameReviewList 管理端：游戏评论列表
func AdminGameReviewList(c *gin.Context) {
	gameID := parseUint64(c.DefaultQuery("game_id", "0"))
	keyword := strings.TrimSpace(c.DefaultQuery("keyword", ""))
	status := strings.TrimSpace(c.DefaultQuery("status", "")) // "", "0", "1"
	page := parsePage(c.DefaultQuery("page", "1"))
	pageSize := parsePageSize(c.DefaultQuery("page_size", "20"))

	db := database.DB().Model(&model.GameReview{})
	if gameID > 0 {
		db = db.Where("game_id = ?", gameID)
	}
	if status == "0" || status == "1" {
		db = db.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("content LIKE ?", like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "统计失败")
		return
	}

	var rows []model.GameReview
	if err := db.Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OKPage(c, rows, total)
}

// AdminGameReviewDelete 管理端：删除评论
func AdminGameReviewDelete(c *gin.Context) {
	id := parseUint64(c.Param("id"))
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	if err := database.DB().Delete(&model.GameReview{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	_ = database.DB().Where("review_id = ?", id).Delete(&model.GameReviewVote{}).Error
	response.OK(c, nil)
}

type adminGameReviewStatusReq struct {
	Status int8 `json:"status" binding:"required"` // 0/1
}

// AdminGameReviewSetStatus 管理端：隐藏/展示评论
func AdminGameReviewSetStatus(c *gin.Context) {
	id := parseUint64(c.Param("id"))
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var req adminGameReviewStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if req.Status != 0 && req.Status != 1 {
		response.Error(c, 400, "status 不合法")
		return
	}
	if err := database.DB().Model(&model.GameReview{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

