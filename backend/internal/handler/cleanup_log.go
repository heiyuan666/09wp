package handler

import (
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AdminCleanupLogList 管理端：清理任务日志列表
func AdminCleanupLogList(c *gin.Context) {
	page, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page", "1")))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page_size", "20")))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	task := strings.TrimSpace(c.DefaultQuery("task", "global_search_cleanup"))
	status := strings.TrimSpace(c.Query("status"))
	platform := strings.TrimSpace(c.Query("platform"))
	resourceID := strings.TrimSpace(c.Query("resource_id"))

	dbq := database.DB().Model(&model.CleanupTaskLog{}).Where("task = ?", task)
	if status != "" {
		dbq = dbq.Where("status = ?", status)
	}
	if platform != "" {
		dbq = dbq.Where("platform = ?", platform)
	}
	if resourceID != "" {
		if rid, err := strconv.ParseUint(resourceID, 10, 64); err == nil && rid > 0 {
			dbq = dbq.Where("resource_id = ?", rid)
		}
	}

	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	var list []model.CleanupTaskLog
	if err := dbq.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}
