package handler

import (
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UserProfile 前台用户个人信息（占位）
func UserProfile(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)

	var user model.User
	if err := database.DB().First(&user, userID).Error; err != nil {
		response.Error(c, 404, "用户不存在")
		return
	}

	response.OK(c, toUserItem(user))
}

// AdminChangePassword 管理员修改密码
func AdminChangePassword(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	adminID, _ := userIDVal.(uint64)

	var req struct {
		OldPassword     string `json:"oldPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=6"`
		ConfirmPassword string `json:"confirmPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	var admin model.Admin
	if err := database.DB().First(&admin, adminID).Error; err != nil {
		response.Error(c, 404, "管理员不存在")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.OldPassword)); err != nil {
		response.Error(c, 400, "原密码错误")
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		response.Error(c, 400, "两次输入密码不一致")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, 500, "密码加密失败")
		return
	}

	if err := database.DB().Model(&admin).Update("password_hash", string(hash)).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}

	response.OK(c, nil)
}

// AdminUserList 后台用户列表（分页 + 搜索），映射到 DFAN IUserListResponse
func AdminUserList(c *gin.Context) {
	var list []model.User
	db := database.DB().Model(&model.User{})

	// DFAN 用户列表查询参数
	username := c.Query("username")
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	name := c.Query("name")
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
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
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	sortOrder := c.DefaultQuery("sortOrder", "desc")
	orderExpr := "id DESC"
	if sortOrder == "asc" {
		orderExpr = "id ASC"
	}

	if err := db.Order(orderExpr).
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	items := make([]map[string]interface{}, 0, len(list))
	for _, u := range list {
		items = append(items, toUserItem(u))
	}

	response.OK(c, gin.H{
		"list":     items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// AdminUserChangeStatus 后台启用/禁用用户
func AdminUserChangeStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status int8 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := database.DB().Model(&model.User{}).Where("id = ?", id).
		Update("status", req.Status).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// AdminUserDelete 后台删除用户
func AdminUserDelete(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB().Delete(&model.User{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

// UserChangePassword 前台用户修改密码
func UserChangePassword(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)

	var req struct {
		OldPassword     string `json:"oldPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=6"`
		ConfirmPassword string `json:"confirmPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	var user model.User
	if err := database.DB().First(&user, userID).Error; err != nil {
		response.Error(c, 404, "用户不存在")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		response.Error(c, 400, "原密码错误")
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		response.Error(c, 400, "两次输入密码不一致")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, 500, "密码加密失败")
		return
	}

	if err := database.DB().Model(&user).Update("password_hash", string(hash)).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}

	response.OK(c, nil)
}

// UserFavoriteList 查看用户收藏的资源
func UserFavoriteList(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)

	db := database.DB()

	var favs []model.UserFavorite
	if err := db.Where("user_id = ?", userID).Find(&favs).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	// 返回收藏对应的资源详情列表
	var resourceIDs []uint64
	for _, f := range favs {
		resourceIDs = append(resourceIDs, f.ResourceID)
	}
	if len(resourceIDs) == 0 {
		response.OK(c, []model.Resource{})
		return
	}

	var resources []model.Resource
	if err := db.Where("id IN ?", resourceIDs).Find(&resources).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	response.OK(c, resources)
}

// UserFavoriteAdd 收藏资源
func UserFavoriteAdd(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)
	resID := c.Param("resource_id")

	var fav model.UserFavorite
	if err := database.DB().Where("user_id = ? AND resource_id = ?", userID, resID).
		First(&fav).Error; err == nil {
		response.OK(c, nil)
		return
	}

	fav = model.UserFavorite{
		UserID:     userID,
		ResourceID: parseUint(resID),
	}
	if err := database.DB().Create(&fav).Error; err != nil {
		response.Error(c, 500, "收藏失败")
		return
	}
	response.OK(c, nil)
}

// UserFavoriteRemove 取消收藏
func UserFavoriteRemove(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)
	resID := c.Param("resource_id")

	if err := database.DB().Where("user_id = ? AND resource_id = ?", userID, resID).
		Delete(&model.UserFavorite{}).Error; err != nil {
		response.Error(c, 500, "取消失败")
		return
	}
	response.OK(c, nil)
}

func parseUint(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

// AdminCreateUser 创建用户（DFAN）
func AdminCreateUser(c *gin.Context) {
	var req struct {
		ID       string `json:"id"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		RoleID   string `json:"roleId"`
		Status   string `json:"status" binding:"required"` // active/inactive
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if req.Password == "" {
		response.Error(c, 400, "密码不能为空")
		return
	}

	var statusVal int8
	switch req.Status {
	case "inactive":
		statusVal = 0
	default:
		statusVal = 1
	}

	email := req.Email
	if email == "" {
		// 前端允许邮箱为空，这里给一个唯一占位值，避免唯一索引冲突
		email = req.Username + "@local.pan"
	}
	var dup int64
	_ = database.DB().Model(&model.User{}).Where("username = ? OR email = ?", req.Username, email).Count(&dup).Error
	if dup > 0 {
		response.Error(c, 409, "用户名或邮箱已存在")
		return
	}

	h, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, 500, "密码加密失败")
		return
	}
	hash := string(h)

	user := model.User{
		Username:     req.Username,
		Email:        email,
		Status:       statusVal,
		Name:         req.Name,
		PasswordHash: hash,
	}
	if req.RoleID != "" {
		rid := parseUint(req.RoleID)
		user.RoleID = &rid
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	if err := database.DB().Create(&user).Error; err != nil {
		response.Error(c, 500, "创建失败:"+err.Error())
		return
	}
	response.OK(c, nil)
}

// AdminUpdateUser 更新用户（DFAN）
func AdminUpdateUser(c *gin.Context) {
	var req struct {
		ID       string `json:"id" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		RoleID   string `json:"roleId"`
		Status   string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	updates := map[string]interface{}{
		"username": req.Username,
		"name":     req.Name,
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.RoleID != "" {
		updates["role_id"] = parseUint(req.RoleID)
	} else {
		updates["role_id"] = nil
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	switch req.Status {
	case "active":
		updates["status"] = 1
	default:
		updates["status"] = 0
	}

	if req.Password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			response.Error(c, 500, "密码加密失败")
			return
		}
		updates["password_hash"] = string(h)
	}

	if err := database.DB().Model(&model.User{}).Where("id = ?", req.ID).
		Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败:"+err.Error())
		return
	}

	response.OK(c, nil)
}

// AdminUserDetail 根据 ID 获取用户详情（DFAN）
func AdminUserDetail(c *gin.Context) {
	id := c.Param("id")
	var user model.User
	if err := database.DB().First(&user, id).Error; err != nil {
		response.Error(c, 404, "用户不存在")
		return
	}
	response.OK(c, toUserItem(user))
}

// AdminUserBatchDelete 批量删除用户（DFAN）
func AdminUserBatchDelete(c *gin.Context) {
	var ids []uint64
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := database.DB().Where("id IN ?", ids).Delete(&model.User{}).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

// UpdateProfile 修改当前用户个人信息（DFAN）
func UpdateProfile(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	adminID, _ := userIDVal.(uint64)

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数错误")
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		response.Error(c, 400, "username is required")
		return
	}

	var dup int64
	if err := database.DB().Model(&model.Admin{}).
		Where("username = ? AND id <> ?", username, adminID).
		Count(&dup).Error; err == nil && dup > 0 {
		response.Error(c, 400, "username already exists")
		return
	}

	updates := map[string]interface{}{
		"username": username,
		"email":    req.Email,
	}

	if err := database.DB().Model(&model.Admin{}).Where("id = ?", adminID).
		Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}

	response.OK(c, nil)
}

// UpdateAvatar 单独修改头像（DFAN）
func UpdateAvatar(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)

	var req struct {
		Avatar string `json:"avatar" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	if err := database.DB().Model(&model.User{}).Where("id = ?", userID).
		Update("avatar", req.Avatar).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// UserPermissions 返回菜单和按钮权限（这里只返回空，方便前端运行）
func UserPermissions(c *gin.Context) {
	// 从数据库读取真实菜单，并组装树返回
	var allMenus []model.Menu
	if err := database.DB().Where("status = 1").Order("`order` ASC, id ASC").Find(&allMenus).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	// directory/menu 用于侧边栏；button 单独下发到 buttonPermissions
	menus := make([]model.Menu, 0, len(allMenus))
	buttonPermissions := make([]string, 0, 16)
	for _, m := range allMenus {
		if m.Type == "button" {
			if m.Permission != "" {
				buttonPermissions = append(buttonPermissions, m.Permission)
			}
			continue
		}
		menus = append(menus, m)
	}

	type node map[string]interface{}
	nodes := make(map[uint64]node, len(menus))
	childrenMap := make(map[uint64][]node)
	var roots []node

	isUnwanted := func(path string, title string) bool {
		p := strings.TrimSpace(path)
		p = strings.TrimPrefix(p, "/")
		p = strings.ToLower(p)
		if strings.HasPrefix(p, "extended/") || strings.HasPrefix(p, "dashboard/") || strings.HasPrefix(p, "demo/") {
			return true
		}
		if p == "system/role" || p == "system/menu" {
			return true
		}
		if title == "扩展组件" || title == "功能演示" {
			return true
		}
		return false
	}

	for _, m := range menus {
		if isUnwanted(m.Path, m.Title) {
			continue
		}
		status := "inactive"
		if m.Status == 1 {
			status = "active"
		}
		n := node{
			"id":         strconv.FormatUint(m.ID, 10),
			"type":       m.Type,
			"path":       m.Path,
			"title":      m.Title,
			"icon":       m.Icon,
			"parentId":   nil,
			"order":      m.Order,
			"status":     status,
			"permission": m.Permission,
			"children":   []node{},
		}
		if m.ParentID != nil {
			n["parentId"] = strconv.FormatUint(*m.ParentID, 10)
		}
		nodes[m.ID] = n
	}

	for _, m := range menus {
		if isUnwanted(m.Path, m.Title) {
			continue
		}
		n := nodes[m.ID]
		if n == nil {
			continue
		}
		if m.ParentID == nil {
			roots = append(roots, n)
			continue
		}
		// 如果父节点被过滤掉，则把该节点提升为根节点
		if _, ok := nodes[*m.ParentID]; !ok {
			roots = append(roots, n)
			continue
		}
		childrenMap[*m.ParentID] = append(childrenMap[*m.ParentID], n)
	}

	for pid, kids := range childrenMap {
		if parent, ok := nodes[pid]; ok {
			parent["children"] = kids
		}
	}

	permSet := map[string]struct{}{
		"user:add":      {},
		"user:edit":     {},
		"user:delete":   {},
		"config:update": {},
	}
	for _, p := range buttonPermissions {
		// 显式排除你要下线的角色/菜单相关按钮权限
		if strings.HasPrefix(p, "role:") || strings.HasPrefix(p, "menu:") {
			continue
		}
		permSet[p] = struct{}{}
	}
	mergedPerms := make([]string, 0, len(permSet))
	for p := range permSet {
		mergedPerms = append(mergedPerms, p)
	}

	response.OK(c, gin.H{
		"menus":             roots,
		"buttonPermissions": mergedPerms,
	})
}

// CurrentUserInfo 返回当前登录用户信息（DFAN）
func CurrentUserInfo(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	adminID, _ := userIDVal.(uint64)

	var admin model.Admin
	if err := database.DB().First(&admin, adminID).Error; err != nil {
		response.Error(c, 404, "管理员不存在")
		return
	}
	response.OK(c, toAdminItem(admin))
}

// AddLoginLog 记录登录日志（这里简单丢弃，保证接口成功）
func AddLoginLog(c *gin.Context) {
	_ = c.Request.Body.Close()
	response.OK(c, nil)
}

// 将内部 User 映射为 DFAN IUserItem 结构
func toUserItem(u model.User) map[string]interface{} {
	statusStr := "inactive"
	if u.Status == 1 {
		statusStr = "active"
	}
	return map[string]interface{}{
		"id":       strconv.FormatUint(u.ID, 10),
		"username": u.Username,
		"password": "",
		"name":     u.Name,
		"avatar":   u.Avatar,
		"phone":    u.Phone,
		"email":    u.Email,
		"roleId": func() interface{} {
			if u.RoleID == nil {
				return nil
			}
			return strconv.FormatUint(*u.RoleID, 10)
		}(),
		"status":     statusStr,
		"createTime": u.CreatedAt.Format(time.RFC3339),
		"updateTime": u.UpdatedAt.Format(time.RFC3339),
		"bio":        u.Bio,
		"tags":       u.Tags,
		"loginLogs":  []interface{}{},
	}
}

func toAdminItem(a model.Admin) map[string]interface{} {
	statusStr := "inactive"
	if a.Status == 1 {
		statusStr = "active"
	}
	return map[string]interface{}{
		"id":         strconv.FormatUint(a.ID, 10),
		"username":   a.Username,
		"password":   "",
		"name":       a.Username,
		"avatar":     nil,
		"phone":      "",
		"email":      a.Email,
		"roleId":     nil,
		"status":     statusStr,
		"createTime": a.CreatedAt.Format(time.RFC3339),
		"updateTime": a.UpdatedAt.Format(time.RFC3339),
		"bio":        "",
		"tags":       "",
		"loginLogs":  []interface{}{},
	}
}
