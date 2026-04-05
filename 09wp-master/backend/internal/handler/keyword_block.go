package handler

import (
	"net/http"
	"strconv"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminKeywordBlockList 关键词屏蔽词列表
func AdminKeywordBlockList(c *gin.Context) {
	var rows []model.KeywordBlock
	if err := database.DB().
		Model(&model.KeywordBlock{}).
		Order("enabled DESC, id DESC").
		Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": rows})
}

// AdminKeywordBlockCreate 新增关键词屏蔽词
func AdminKeywordBlockCreate(c *gin.Context) {
	var req struct {
		Keyword string `json:"keyword" binding:"required"`
		Enabled *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	norm := service.NormalizeKeywordBlockForStore(req.Keyword)
	if norm == "" {
		response.Error(c, 400, "keyword 不能为空")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	row := model.KeywordBlock{
		Keyword: norm,
		Enabled: enabled,
	}
	// upsert：keyword 唯一则更新
	if err := database.DB().
		Where("keyword = ?", norm).
		Assign(model.KeywordBlock{Enabled: enabled}).
		FirstOrCreate(&row).Error; err != nil {
		response.Error(c, 500, "新增失败")
		return
	}

	service.ClearKeywordBlockCache()
	response.OK(c, row)
}

// AdminKeywordBlockUpdate 更新关键词屏蔽词
func AdminKeywordBlockUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}

	var req struct {
		Keyword *string `json:"keyword"`
		Enabled *bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	var row model.KeywordBlock
	if err := database.DB().First(&row, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "记录不存在")
			return
		}
		response.Error(c, 500, "查询失败")
		return
	}

	updates := map[string]any{}
	if req.Keyword != nil {
		norm := service.NormalizeKeywordBlockForStore(*req.Keyword)
		if norm == "" {
			response.Error(c, 400, "keyword 不能为空")
			return
		}
		updates["keyword"] = norm
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if len(updates) == 0 {
		response.OK(c, row)
		return
	}

	if err := database.DB().Model(&row).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}

	service.ClearKeywordBlockCache()
	response.OK(c, row)
}

// AdminKeywordBlockDelete 删除关键词屏蔽词
func AdminKeywordBlockDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var row model.KeywordBlock
	if err := database.DB().First(&row, id).Error; err != nil {
		response.Error(c, 404, "记录不存在")
		return
	}
	if err := database.DB().Delete(&row).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	service.ClearKeywordBlockCache()
	response.OK(c, gin.H{"deleted": true})
}

