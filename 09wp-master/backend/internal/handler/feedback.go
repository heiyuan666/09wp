package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type feedbackCreateReq struct {
	ResourceID uint64 `json:"resource_id" binding:"required"`
	Type       string `json:"type" binding:"required"` // link_invalid / password_error / content_error / other / report_feedback
	Content    string `json:"content"`
	Contact    string `json:"contact"`
}

var allowedFeedbackTypes = map[string]struct{}{
	"link_invalid":    {},
	"password_error":  {},
	"content_error":   {},
	"other":           {},
	"report_feedback": {},
}

var allowedFeedbackStatuses = map[string]struct{}{
	"pending":   {},
	"processed": {},
}

// FeedbackCreate 用户在详情页或搜索页提交反馈（前台公开，不要求登录）
func FeedbackCreate(c *gin.Context) {
	var req feedbackCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	req.Type = strings.TrimSpace(req.Type)
	req.Content = strings.TrimSpace(req.Content)
	req.Contact = strings.TrimSpace(req.Contact)

	if req.ResourceID == 0 {
		response.Error(c, 400, "resource_id 不能为空")
		return
	}
	if req.Type == "" {
		response.Error(c, 400, "type 不能为空")
		return
	}
	if _, ok := allowedFeedbackTypes[req.Type]; !ok {
		response.Error(c, 400, "type 不合法")
		return
	}

	var resource model.Resource
	if err := database.DB().
		Model(&model.Resource{}).
		Select("id").
		Where("id = ?", req.ResourceID).
		First(&resource).Error; err != nil {
		response.Error(c, 404, "资源不存在")
		return
	}

	// 失效举报支持快捷上报：前端可以不传 content。
	if req.Content == "" && req.Type == "link_invalid" {
		req.Content = "搜索页快捷失效举报"
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

	row := model.ResourceFeedback{
		ResourceID: req.ResourceID,
		Type:       req.Type,
		Content:    req.Content,
		Contact:    req.Contact,
		Status:     "pending",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := database.DB().Create(&row).Error; err != nil {
		response.Error(c, 500, "提交失败")
		return
	}

	response.OK(c, row)
}

// AdminFeedbackList 后台反馈管理（查看列表）
func AdminFeedbackList(c *gin.Context) {
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

	db := database.DB().Model(&model.ResourceFeedback{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if typeFilter != "" {
		db = db.Where("type = ?", typeFilter)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "统计失败")
		return
	}

	var list []model.ResourceFeedback
	if err := db.Order("created_at DESC, id DESC").
		Limit(pageSize).Offset((page-1)*pageSize).
		Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	response.OKPage(c, list, total)
}

type feedbackStatusUpdateReq struct {
	Status string `json:"status" binding:"required"`
}

// AdminFeedbackUpdateStatus 后台反馈管理：更新处理状态
func AdminFeedbackUpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "反馈 ID 不合法")
		return
	}

	var req feedbackStatusUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	req.Status = strings.TrimSpace(req.Status)
	if _, ok := allowedFeedbackStatuses[req.Status]; !ok {
		response.Error(c, 400, "status 不合法")
		return
	}

	tx := database.DB().Model(&model.ResourceFeedback{}).Where("id = ?", id).Updates(map[string]any{
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
