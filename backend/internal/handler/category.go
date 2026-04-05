package handler

import (
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AdminCategoryCreate 后台添加分类（简版）
func AdminCategoryCreate(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Slug      string `json:"slug" binding:"required"`
		SortOrder int    `json:"sort_order"`
		Status    int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	cat := model.Category{
		Name:      req.Name,
		Slug:      req.Slug,
		SortOrder: req.SortOrder,
		Status:    req.Status,
	}
	if cat.Status == 0 {
		cat.Status = 1
	}

	if err := database.DB().Create(&cat).Error; err != nil {
		response.Error(c, 500, "创建失败")
		return
	}

	response.OK(c, cat)
}

// AdminCategoryList 后台分类列表（简版）
func AdminCategoryList(c *gin.Context) {
	var list []model.Category
	if err := database.DB().Order("sort_order DESC, id DESC").Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OKPage(c, list, int64(len(list)))
}

// AdminCategoryUpdate 编辑分类
func AdminCategoryUpdate(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name      string `json:"name" binding:"required"`
		Slug      string `json:"slug" binding:"required"`
		SortOrder int    `json:"sort_order"`
		Status    int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := database.DB().Model(&model.Category{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":       req.Name,
			"slug":       req.Slug,
			"sort_order": req.SortOrder,
			"status":     req.Status,
		}).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// AdminCategoryDelete 删除分类
func AdminCategoryDelete(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB().Delete(&model.Category{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

// AdminCategoryChangeStatus 修改分类状态
func AdminCategoryChangeStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status int8 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := database.DB().Model(&model.Category{}).Where("id = ?", id).
		Update("status", req.Status).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// AdminCategoryChangeSort 修改分类排序
func AdminCategoryChangeSort(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		SortOrder int `json:"sort_order" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := database.DB().Model(&model.Category{}).Where("id = ?", id).
		Update("sort_order", req.SortOrder).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// CategoryListPublic 前台分类列表（只返回启用的）
func CategoryListPublic(c *gin.Context) {
	var list []model.Category
	if err := database.DB().Where("status = 1").Order("sort_order DESC, id DESC").Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, list)
}

