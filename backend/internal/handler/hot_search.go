package handler

import (
	"strconv"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// GetPublicHotSearch 热搜榜（无需登录）
func GetPublicHotSearch(c *gin.Context) {
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil && !cfg.HotSearchEnabled {
		response.OK(c, gin.H{"list": []gin.H{}})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	rows, err := service.ListHotSearchKeywords(limit)
	if err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		// 屏蔽词：前台不展示
		if service.IsKeywordBlockedText(r.Keyword) {
			continue
		}
		out = append(out, gin.H{
			"keyword":      r.Keyword,
			"search_count": r.SearchCount,
		})
	}
	response.OK(c, gin.H{"list": out})
}
