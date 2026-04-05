package handler

import (
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Home 首页数据：最新资源、热门资源、分类导航
func Home(c *gin.Context) {
	db := database.DB()

	var latest []model.Resource
	_ = db.Where("status = 1").Order("created_at DESC").Limit(10).Find(&latest).Error

	var hot []model.Resource
	_ = db.Where("status = 1").Order("view_count DESC").Limit(10).Find(&hot).Error

	var categories []model.Category
	if err := db.Where("status = 1").Order("sort_order DESC, id DESC").Find(&categories).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	hotSearchOut := make([]gin.H, 0)
	var cfg model.SystemConfig
	showHotSearch := true
	if err := db.Order("id ASC").First(&cfg).Error; err == nil {
		showHotSearch = cfg.HotSearchEnabled
	}

	if showHotSearch {
		hotSearches, _ := service.ListHotSearchKeywords(24)
		hotSearchOut = make([]gin.H, 0, len(hotSearches))
		for _, r := range hotSearches {
			hotSearchOut = append(hotSearchOut, gin.H{
				"keyword":      r.Keyword,
				"search_count": r.SearchCount,
			})
		}
	}

	response.OK(c, gin.H{
		"latest":       latest,
		"hot":          hot,
		"categories":   categories,
		"hot_searches": hotSearchOut,
	})
}
