package handler

import (
	"encoding/json"
	"strconv"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// RolePage 获取角色分页
func RolePage(c *gin.Context) {
	db := database.DB().Model(&model.Role{})

	if name := c.Query("name"); name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if code := c.Query("code"); code != "" {
		db = db.Where("code LIKE ?", "%"+code+"%")
	}
	if status := c.Query("status"); status != "" {
		switch status {
		case "active":
			db = db.Where("status = 1")
		case "inactive":
			db = db.Where("status = 0")
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "统计失败")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 20
	}
	sortOrder := c.DefaultQuery("sortOrder", "desc")
	orderExpr := "id DESC"
	if sortOrder == "asc" {
		orderExpr = "id ASC"
	}

	var roles []model.Role
	if err := db.Order(orderExpr).Limit(pageSize).Offset((page-1)*pageSize).Find(&roles).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	items := make([]map[string]interface{}, 0, len(roles))
	for _, r := range roles {
		items = append(items, toRoleItem(r))
	}

	response.OK(c, gin.H{
		"list":     items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// RoleCreate 创建角色
func RoleCreate(c *gin.Context) {
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Code        string   `json:"code" binding:"required"`
		Description string   `json:"description"`
		Status      string   `json:"status" binding:"required"` // active/inactive
		MenuIDs     []string `json:"menuIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	var status int8
	switch req.Status {
	case "inactive":
		status = 0
	default:
		status = 1
	}
	raw, _ := json.Marshal(req.MenuIDs)
	role := model.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsBuiltIn:   false,
		Status:      status,
		MenuIDs:     string(raw),
	}
	if err := database.DB().Create(&role).Error; err != nil {
		response.Error(c, 500, "创建失败")
		return
	}
	response.OK(c, nil)
}

// RoleInfo 角色详情
func RoleInfo(c *gin.Context) {
	id := c.Param("id")
	var role model.Role
	if err := database.DB().First(&role, id).Error; err != nil {
		response.Error(c, 404, "角色不存在")
		return
	}
	response.OK(c, toRoleItem(role))
}

// RoleUpdate 更新角色
func RoleUpdate(c *gin.Context) {
	var req struct {
		ID          string   `json:"id" binding:"required"`
		Name        string   `json:"name" binding:"required"`
		Code        string   `json:"code" binding:"required"`
		Description string   `json:"description"`
		Status      string   `json:"status" binding:"required"`
		MenuIDs     []string `json:"menuIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	var role model.Role
	if err := database.DB().First(&role, req.ID).Error; err != nil {
		response.Error(c, 404, "角色不存在")
		return
	}
	if role.IsBuiltIn {
		response.Error(c, 403, "内置角色不允许修改")
		return
	}

	var status int8
	switch req.Status {
	case "inactive":
		status = 0
	default:
		status = 1
	}
	raw, _ := json.Marshal(req.MenuIDs)
	if err := database.DB().Model(&model.Role{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"name":        req.Name,
		"code":        req.Code,
		"description": req.Description,
		"status":      status,
		"menu_ids":    string(raw),
	}).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// RoleDelete 批量删除角色
func RoleDelete(c *gin.Context) {
	var ids []uint64
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		response.Error(c, 400, "参数错误")
		return
	}

	// 保护内置角色
	var builtInCnt int64
	if err := database.DB().Model(&model.Role{}).Where("id IN ? AND is_built_in = 1", ids).Count(&builtInCnt).Error; err == nil && builtInCnt > 0 {
		response.Error(c, 403, "内置角色不允许删除")
		return
	}

	if err := database.DB().Where("id IN ?", ids).Delete(&model.Role{}).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

func toRoleItem(r model.Role) map[string]interface{} {
	var status string
	switch r.Status {
	case 1:
		status = "active"
	default:
		status = "inactive"
	}
	menuIDs := []string{}
	_ = json.Unmarshal([]byte(r.MenuIDs), &menuIDs)

	return map[string]interface{}{
		"id":          strconv.FormatUint(r.ID, 10),
		"name":        r.Name,
		"code":        r.Code,
		"description": r.Description,
		"isBuiltIn":   r.IsBuiltIn,
		"status":      status,
		"menuIds":     menuIDs,
		"createTime":  r.CreatedAt.Format(time.RFC3339),
		"updateTime":  r.UpdatedAt.Format(time.RFC3339),
	}
}

