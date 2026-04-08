package handler

import (
	"net/http"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type gameFeedbackCreateReq struct {
	GameID       uint64 `json:"game_id" binding:"required"`
	GameResource uint64 `json:"game_resource_id" binding:"required"`
	DownloadURL  string `json:"download_url"`
	ExtractCode  string `json:"extract_code"`
	Type         string `json:"type" binding:"required"` // link_invalid / other
	Content      string `json:"content"`
	Contact      string `json:"contact"`
}

var allowedGameFeedbackTypes = map[string]struct{}{
	"link_invalid": {},
	"other":        {},
}

// GameFeedbackCreate 游戏详情页资源反馈（公开，不要求登录）
func GameFeedbackCreate(c *gin.Context) {
	var req gameFeedbackCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	req.Type = strings.TrimSpace(req.Type)
	req.Content = strings.TrimSpace(req.Content)
	req.Contact = strings.TrimSpace(req.Contact)
	req.DownloadURL = strings.TrimSpace(req.DownloadURL)
	req.ExtractCode = strings.TrimSpace(req.ExtractCode)

	if req.GameID == 0 || req.GameResource == 0 {
		response.Error(c, 400, "参数错误")
		return
	}
	if _, ok := allowedGameFeedbackTypes[req.Type]; !ok {
		response.Error(c, 400, "type 不合法")
		return
	}

	// 校验 game 与 game_resource 存在且匹配
	var gr model.GameResource
	if err := database.DB().
		Select("id", "game_id", "download_url", "extract_code").
		Where("id = ? AND game_id = ?", req.GameResource, req.GameID).
		First(&gr).Error; err != nil {
		response.Error(c, 404, "资源不存在")
		return
	}

	if req.Content == "" && req.Type == "link_invalid" {
		req.Content = "用户标记该下载链接可能已失效"
	}
	if req.Content == "" {
		response.Error(c, 400, "content 不能为空")
		return
	}
	if len(req.Content) > 2000 {
		req.Content = req.Content[:2000]
	}
	if len(req.Contact) > 255 {
		req.Contact = req.Contact[:255]
	}

	row := model.GameResourceFeedback{
		GameID:       req.GameID,
		GameResource: req.GameResource,
		DownloadURL:  req.DownloadURL,
		ExtractCode:  req.ExtractCode,
		Type:         req.Type,
		Content:      req.Content,
		Contact:      req.Contact,
		Status:       "pending",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if row.DownloadURL == "" {
		row.DownloadURL = strings.TrimSpace(gr.DownloadURL)
	}
	if row.ExtractCode == "" {
		row.ExtractCode = strings.TrimSpace(gr.ExtractCode)
	}

	if err := database.DB().Create(&row).Error; err != nil {
		response.Error(c, 500, "提交失败")
		return
	}
	response.OK(c, row)
}

