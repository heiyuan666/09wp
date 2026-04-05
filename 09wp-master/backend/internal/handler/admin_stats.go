package handler

import (
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AdminStats 后台控制台统计（管理员）
func AdminStats(c *gin.Context) {
	var (
		resourceTotal int64
		resourceOn    int64
		userTotal     int64
		categoryTotal int64
		subPending    int64
		feedbackTotal int64
		searchTotal   int64
	)

	_ = database.DB().Model(&model.Resource{}).Count(&resourceTotal).Error
	_ = database.DB().Model(&model.Resource{}).Where("status = 1").Count(&resourceOn).Error
	_ = database.DB().Model(&model.User{}).Count(&userTotal).Error
	_ = database.DB().Model(&model.Category{}).Count(&categoryTotal).Error
	_ = database.DB().Model(&model.UserResourceSubmission{}).Where("status = ?", "pending").Count(&subPending).Error
	_ = database.DB().Model(&model.ResourceFeedback{}).Count(&feedbackTotal).Error
	_ = database.DB().Model(&model.SearchHotWord{}).Select("COALESCE(SUM(search_count), 0)").Scan(&searchTotal).Error

	response.OK(c, gin.H{
		"resources_total":     resourceTotal,
		"resources_online":    resourceOn,
		"users_total":         userTotal,
		"categories_total":    categoryTotal,
		"submissions_pending": subPending,
		"feedbacks_total":     feedbackTotal,
		"searches_total":      searchTotal,
	})
}

