package handler

import (
	"net/http"
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// HaokaCategoryListPublic 公开：号卡分类（电信/移动/联通）
func HaokaCategoryListPublic(c *gin.Context) {
	var cats []model.HaokaCategory
	if err := database.DB().Where("status = 1").Order("id ASC").Find(&cats).Error; err != nil {
		response.Error(c, 500, "分类查询失败")
		return
	}
	response.OK(c, cats)
}

// HaokaProductListPublic 公开：号卡商品列表（展示上架且启用）
func HaokaProductListPublic(c *gin.Context) {
	categoryID := strings.TrimSpace(c.Query("category_id"))
	operator := strings.TrimSpace(c.Query("operator")) // 可选：电信/移动/联通
	keyword := strings.TrimSpace(c.Query("q"))

	// 默认只展示上架
	flag := true
	if f := strings.TrimSpace(c.Query("flag")); f != "" {
		if f == "false" || f == "0" {
			flag = false
		}
	}

	page := 1
	pageSize := 20
	if v := strings.TrimSpace(c.DefaultQuery("page", "1")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := strings.TrimSpace(c.DefaultQuery("page_size", "20")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			if n < 1 {
				n = 1
			}
			if n > 100 {
				n = 100
			}
			pageSize = n
		}
	}

	query := database.DB().Model(&model.HaokaProduct{}).
		Where("status = 1").
		Where("flag = ?", flag)
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if operator != "" {
		query = query.Where("operator = ?", operator)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("product_name LIKE ?", like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, 500, "列表查询失败")
		return
	}

	var list []model.HaokaProduct
	if err := query.
		Order("id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error; err != nil {
		response.Error(c, 500, "列表查询失败")
		return
	}

	type item struct {
		model.HaokaProduct
		CategoryName string `json:"category_name"`
	}
	items := make([]item, 0, len(list))
	for _, p := range list {
		var cat model.HaokaCategory
		_ = database.DB().Where("id = ?", p.CategoryID).First(&cat).Error
		items = append(items, item{HaokaProduct: p, CategoryName: cat.Name})
	}

	response.OK(c, gin.H{
		"list":       items,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// HaokaProductDetailPublic 公开：号卡商品详情（含 skus）
func HaokaProductDetailPublic(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Error(c, 400, "id 不能为空")
		return
	}

	var p model.HaokaProduct
	if err := database.DB().Where("id = ? AND status = 1", id).First(&p).Error; err != nil {
		response.Error(c, 404, "号卡不存在")
		return
	}

	var cat model.HaokaCategory
	_ = database.DB().Where("id = ?", p.CategoryID).First(&cat).Error

	// 注意：skus 表里存的是外部 productID（p.ProductID），不是内部 id
	var skus []model.HaokaSku
	_ = database.DB().Where("product_id = ?", p.ProductID).Order("id ASC").Find(&skus).Error

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"product":       p,
			"category_name": cat.Name,
			"skus":          skus,
		},
	})
}

