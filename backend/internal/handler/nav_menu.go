package handler

import (
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type navMenuReq struct {
	Title     string `json:"title" binding:"required"`
	Path      string `json:"path"`
	Position  string `json:"position" binding:"required"` // top_nav / home_promo
	SortOrder int    `json:"sort_order"`
	Visible   *bool  `json:"visible"`
}

// NavMenuList 管理端导航菜单列表
func NavMenuList(c *gin.Context) {
	pos := c.Query("position")
	db := database.DB().Model(&model.NavigationMenu{})
	if pos != "" {
		db = db.Where("position = ?", pos)
	}
	var list []model.NavigationMenu
	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "统计失败")
		return
	}
	if err := db.Order("sort_order DESC, id ASC").Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OKPage(c, list, total)
}

// NavMenuCreate 新增导航菜单
func NavMenuCreate(c *gin.Context) {
	var req navMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		response.Error(c, 400, "标题不能为空")
		return
	}
	path := strings.TrimSpace(req.Path)
	if req.Position != "top_nav" && req.Position != "home_promo" {
		response.Error(c, 400, "position 不合法")
		return
	}
	visible := true
	if req.Visible != nil {
		visible = *req.Visible
	}
	row := model.NavigationMenu{
		Title:     title,
		Path:      path,
		Position:  req.Position,
		SortOrder: req.SortOrder,
		Visible:   visible,
	}
	if err := database.DB().Create(&row).Error; err != nil {
		response.Error(c, 500, "创建失败")
		return
	}
	response.OK(c, row)
}

// NavMenuUpdate 更新导航菜单
func NavMenuUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var req navMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	var row model.NavigationMenu
	if err := database.DB().First(&row, id).Error; err != nil {
		response.Error(c, 404, "记录不存在")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		response.Error(c, 400, "标题不能为空")
		return
	}
	path := strings.TrimSpace(req.Path)
	if req.Position != "top_nav" && req.Position != "home_promo" {
		response.Error(c, 400, "position 不合法")
		return
	}
	visible := true
	if req.Visible != nil {
		visible = *req.Visible
	}
	updates := map[string]interface{}{
		"title":      title,
		"path":       path,
		"position":   req.Position,
		"sort_order": req.SortOrder,
		"visible":    visible,
	}
	if err := database.DB().Model(&row).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// NavMenuDelete 删除导航菜单
func NavMenuDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	if err := database.DB().Delete(&model.NavigationMenu{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

// PublicNavMenus 前台公开导航菜单
func PublicNavMenus(c *gin.Context) {
	pos := c.DefaultQuery("position", "top_nav")
	var list []model.NavigationMenu
	if err := database.DB().
		Where("position = ? AND visible = 1", pos).
		Order("sort_order DESC, id ASC").
		Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list})
}

