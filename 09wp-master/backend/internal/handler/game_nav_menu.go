package handler

import (
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type gameNavMenuReq struct {
	Title     string `json:"title" binding:"required"`
	Path      string `json:"path"`
	Position  string `json:"position" binding:"required"` // top_nav / home_promo
	SortOrder int    `json:"sort_order"`
	Visible   *bool  `json:"visible"`
}

// GameNavMenuList 管理端：游戏导航菜单列表
func GameNavMenuList(c *gin.Context) {
	pos := c.Query("position")
	db := database.DB().Model(&model.GameNavigationMenu{})
	if pos != "" {
		db = db.Where("position = ?", pos)
	}
	var list []model.GameNavigationMenu
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

// GameNavMenuCreate 管理端：新增游戏导航菜单
func GameNavMenuCreate(c *gin.Context) {
	var req gameNavMenuReq
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
	row := model.GameNavigationMenu{
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

// GameNavMenuUpdate 管理端：更新游戏导航菜单
func GameNavMenuUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var req gameNavMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	var row model.GameNavigationMenu
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

// GameNavMenuDelete 管理端：删除游戏导航菜单
func GameNavMenuDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	if err := database.DB().Delete(&model.GameNavigationMenu{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

// PublicGameNavMenus 游戏前台公开导航菜单
func PublicGameNavMenus(c *gin.Context) {
	pos := c.DefaultQuery("position", "top_nav")
	var list []model.GameNavigationMenu
	if err := database.DB().
		Where("position = ? AND visible = 1", pos).
		Order("sort_order DESC, id ASC").
		Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list})
}

