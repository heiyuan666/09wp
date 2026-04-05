package database

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/config"
	"dfan-netdisk-backend/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAdmin 初始化管理员账号（若已存在则跳过）
//
// 默认用户名：admin账号通过环境变量覆盖：
// - ADMIN_USERNAME（默认 admin）
// - ADMIN_PASSWORD（默认 123456）
func SeedAdmin() error {
	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "123456"
	}

	var admin model.Admin
	err := DB().Where("username = ?", username).First(&admin).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin = model.Admin{
		Username:     username,
		PasswordHash: string(hash),
		Email:        "",
		Status:       1,
	}
	return DB().Create(&admin).Error
}

// SeedMenus 初始化默认菜单（若已有则跳过）
func SeedMenus() error {
	// 清理模板自带的无用菜单（避免侧边栏出现“扩展组件/功能演示”等）
	// 这些菜单来自模板初始数据或历史数据，不属于网盘系统。
	unwantedExactPaths := []string{
		"/dashboard/home",
		"/dashboard/analysis",
		"/dashboard/monitor",
		"/extended/button",
		"/extended/dialog",
		"/extended/hoverAnimation",
		"/extended/iconSelector",
		"/extended/textEllipsis",
		"/extended/transitionAnimation",
		"/demo/vxeTable",
		"/system/role",
		"/system/menu",
		// 历史号卡菜单（现在挪到“网盘资源”下）
		"/system/haoka",
	}
	_ = DB().Where("path IN ?", unwantedExactPaths).Delete(&model.Menu{}).Error
	_ = DB().Where("path LIKE ?", "/dashboard/%").Delete(&model.Menu{}).Error
	_ = DB().Where("path LIKE ?", "/extended/%").Delete(&model.Menu{}).Error
	_ = DB().Where("path LIKE ?", "/demo/%").Delete(&model.Menu{}).Error
	// 兼容某些历史数据 path 不带前导 /
	_ = DB().Where("path LIKE ?", "dashboard/%").Delete(&model.Menu{}).Error
	_ = DB().Where("path LIKE ?", "extended/%").Delete(&model.Menu{}).Error
	_ = DB().Where("path LIKE ?", "demo/%").Delete(&model.Menu{}).Error
	// 也可能存成相对路径片段（例如 extended/transitionAnimation）
	_ = DB().Where("path LIKE ?", "%extended/%").Delete(&model.Menu{}).Error
	_ = DB().Where("title IN ?", []string{"扩展组件", "功能演示"}).Delete(&model.Menu{}).Error
	// 清理系统角色/菜单按钮（避免侧边栏/权限列表出现“新增角色/新增菜单”等）
	_ = DB().Where("type = ? AND permission LIKE ?", "button", "role:%").Delete(&model.Menu{}).Error
	_ = DB().Where("type = ? AND permission LIKE ?", "button", "menu:%").Delete(&model.Menu{}).Error
	_ = DB().Where("title IN ?", []string{
		"新增角色", "编辑角色", "删除角色",
		"新增菜单", "编辑菜单", "删除菜单",
	}).Delete(&model.Menu{}).Error

	var netdiskDir model.Menu
	if err := DB().Where("type = ? AND title = ?", "directory", "网盘资源").First(&netdiskDir).Error; err != nil {
		netdiskDir = model.Menu{
			Type:   "directory",
			Path:   "",
			Title:  "网盘资源",
			Icon:   "HOutline:FolderIcon",
			Order:  1,
			Status: 1,
		}
		if err := DB().Create(&netdiskDir).Error; err != nil {
			return err
		}
	}

	ensureMenu := func(path, title, icon string, parentID *uint64, order int) error {
		var m model.Menu
		if err := DB().Where("path = ?", path).First(&m).Error; err == nil {
			return DB().Model(&m).Updates(map[string]interface{}{
				"title":     title,
				"icon":      icon,
				"parent_id": parentID,
				"order":     order,
				"status":    1,
			}).Error
		}
		return DB().Create(&model.Menu{
			Type:     "menu",
			Path:     path,
			Title:    title,
			Icon:     icon,
			ParentID: parentID,
			Order:    order,
			Status:   1,
		}).Error
	}

	if err := ensureMenu("/netdisk/categories", "分类管理", "HOutline:TagIcon", &netdiskDir.ID, 1); err != nil {
		return err
	}
	if err := ensureMenu("/netdisk/resources", "资源管理", "HOutline:LinkIcon", &netdiskDir.ID, 2); err != nil {
		return err
	}
	if err := ensureMenu("/netdisk/submissions", "用户提交审核", "HOutline:ClipboardDocumentCheckIcon", &netdiskDir.ID, 3); err != nil {
		return err
	}

	var gameDir model.Menu
	if err := DB().Where("type = ? AND title = ?", "directory", "游戏管理").First(&gameDir).Error; err != nil {
		gameDir = model.Menu{
			Type:   "directory",
			Path:   "",
			Title:  "游戏管理",
			Icon:   "HOutline:PuzzlePieceIcon",
			Order:  2,
			Status: 1,
		}
		if err := DB().Create(&gameDir).Error; err != nil {
			return err
		}
	}

	if err := ensureMenu("/game/categories", "游戏分类", "HOutline:TagIcon", &gameDir.ID, 1); err != nil {
		return err
	}
	if err := ensureMenu("/game/list", "游戏列表", "HOutline:ListBulletIcon", &gameDir.ID, 2); err != nil {
		return err
	}
	if err := ensureMenu("/game/resources", "下载资源", "HOutline:ArrowDownTrayIcon", &gameDir.ID, 3); err != nil {
		return err
	}

	var sysDir model.Menu
	if err := DB().Where("type = ? AND title = ?", "directory", "系统管理").First(&sysDir).Error; err != nil {
		sysDir = model.Menu{
			Type:   "directory",
			Path:   "",
			Title:  "系统管理",
			Icon:   "HOutline:Cog6ToothIcon",
			Order:  99,
			Status: 1,
		}
		if err := DB().Create(&sysDir).Error; err != nil {
			return err
		}
	}

	// 控制台（DAS）
	if err := ensureMenu("/dashboard", "控制台", "HOutline:Squares2X2Icon", &sysDir.ID, 0); err != nil {
		return err
	}

	if err := ensureMenu("/system/user", "用户管理", "HOutline:UserGroupIcon", &sysDir.ID, 1); err != nil {
		return err
	}
	if err := ensureMenu("/system/config", "系统配置", "HOutline:AdjustmentsHorizontalIcon", &sysDir.ID, 2); err != nil {
		return err
	}
	if err := ensureMenu("/system/tg-channels", "TG频道管理", "HOutline:CloudArrowDownIcon", &sysDir.ID, 3); err != nil {
		return err
	}
	if err := ensureMenu("/system/nav-menu", "导航菜单管理", "HOutline:Bars3BottomLeftIcon", &sysDir.ID, 4); err != nil {
		return err
	}
	if err := ensureMenu("/system/netdisk-credentials", "网盘凭证", "HOutline:KeyIcon", &sysDir.ID, 5); err != nil {
		return err
	}
	if err := ensureMenu("/system/keyword-blocks", "关键词屏蔽", "HOutline:ShieldCheckIcon", &sysDir.ID, 6); err != nil {
		return err
	}
	if err := ensureMenu("/system/feedbacks", "反馈管理", "HOutline:ChatBubbleLeftIcon", &sysDir.ID, 7); err != nil {
		return err
	}
	if err := ensureMenu("/system/rss-subscriptions", "RSS订阅抓取", "HOutline:RssIcon", &sysDir.ID, 8); err != nil {
		return err
	}

	// 号卡（外部套餐同步/管理）挪到“网盘资源”目录下
	if err := ensureMenu("/netdisk/haoka", "号卡", "HOutline:Squares2X2Icon", &netdiskDir.ID, 99); err != nil {
		return err
	}

	return nil
}

// SeedNetdiskCredential 初始化网盘凭证行（从 system_configs 复制，便于与旧数据兼容）
func SeedNetdiskCredential() error {
	var cnt int64
	if err := DB().Model(&model.NetdiskCredential{}).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	var sys model.SystemConfig
	if err := DB().Order("id ASC").First(&sys).Error; err != nil {
		return nil
	}
	n := model.NetdiskCredential{
		ID:                       1,
		QuarkCookie:              sys.QuarkCookie,
		QuarkAutoSave:            sys.QuarkAutoSave,
		QuarkTargetFolderID:      sys.QuarkTargetFolderID,
		QuarkAdFilterEnabled:     sys.QuarkAdFilterEnabled,
		QuarkBannedKeywords:      sys.QuarkBannedKeywords,
		Pan115Cookie:             sys.Pan115Cookie,
		Pan115AutoSave:           sys.Pan115AutoSave,
		Pan115TargetFolderID:     sys.Pan115TargetFolderID,
		TianyiCookie:             sys.TianyiCookie,
		TianyiAutoSave:           sys.TianyiAutoSave,
		TianyiTargetFolderID:     sys.TianyiTargetFolderID,
		Pan123Cookie:             sys.Pan123Cookie,
		Pan123AutoSave:           sys.Pan123AutoSave,
		Pan123TargetFolderID:     sys.Pan123TargetFolderID,
		BaiduCookie:              sys.BaiduCookie,
		BaiduAutoSave:            sys.BaiduAutoSave,
		BaiduTargetPath:          "/",
		XunleiCookie:             sys.XunleiCookie,
		XunleiAutoSave:           sys.XunleiAutoSave,
		XunleiTargetFolderID:     sys.XunleiTargetFolderID,
		UcCookie:                 sys.UcCookie,
		UcAutoSave:               sys.UcAutoSave,
		UcTargetFolderID:         sys.UcTargetFolderID,
		AliyunRefreshToken:       sys.AliyunRefreshToken,
		AliyunAutoSave:           sys.AliyunAutoSave,
		AliyunTargetParentFileID: sys.AliyunTargetParentFileID,
		ReplaceLinkAfterTransfer: sys.ReplaceLinkAfterTransfer,
		UpdatedBy:                0,
	}
	if strings.TrimSpace(n.BaiduTargetPath) == "" {
		n.BaiduTargetPath = "/"
	}
	return DB().Create(&n).Error
}

// SeedRoles 初始化默认角色（若不存在则创建）
func SeedRoles() error {
	var cnt int64
	if err := DB().Model(&model.Role{}).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}

	// 默认给超级管理员分配全部 menu（directory/menu/button 都可）
	var menus []model.Menu
	if err := DB().Order("id ASC").Find(&menus).Error; err != nil {
		return err
	}
	menuIDs := make([]string, 0, len(menus))
	for _, m := range menus {
		menuIDs = append(menuIDs, strconv.FormatUint(m.ID, 10))
	}
	raw, _ := json.Marshal(menuIDs)

	role := model.Role{
		Name:        "超级管理员",
		Code:        "SUPER_ADMIN",
		Description: "系统默认超级管理员角色",
		IsBuiltIn:   true,
		Status:      1,
		MenuIDs:     string(raw),
	}
	return DB().Create(&role).Error
}

// SeedSystemConfig 初始化全局配置（单例）
func SeedSystemConfig() error {
	var cnt int64
	if err := DB().Model(&model.SystemConfig{}).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}

	cfg := model.SystemConfig{
		SiteTitle:                  "网盘资源导航系统",
		AdminEmail:                 "admin@example.com",
		SupportEmail:               "support@example.com",
		ContactPhone:               "",
		ContactQQ:                  "",
		LogoURL:                    "",
		FaviconURL:                 "",
		SeoKeywords:                "网盘,资源,导航",
		SeoDescription:             "网盘资源导航管理系统",
		IcpRecord:                  "",
		FooterText:                 "©️零九cdn www.09cdn.com",
		ClarityProjectID:           "",
		ClarityEnabled:             false,
		FriendLinks:                "[]",
		AllowRegister:              true,
		HaokaUserID:                "",
		HaokaSecret:                "",
		HaokaSyncEnabled:           false,
		HaokaSyncInterval:          3600,
		HaokaOrderURL:              "",
		HaokaAgentRegURL:           "",
		SubmissionNeedReview:       true,
		SubmissionAutoTransfer:     false,
		ResourceDetailAutoTransfer: false,
		SmtpHost:                   "",
		SmtpPort:                   25,
		SmtpUser:                   "",
		SmtpPass:                   "",
		SmtpFrom:                   "",
		TgBotToken:                 "",
		TgProxyURL:                 "",
		TgAPIID:                    0,
		TgAPIHash:                  "",
		TgSession:                  "",
		PanCheckBaseURL:            config.DefaultPanCheckBaseURL,
		TgChannelChatID:            "",
		TgSyncEnabled:              false,
		TgSyncInterval:             300,
		TgDefaultCatID:             0,
		TgLastUpdateID:             0,
		QuarkCookie:                "",
		QuarkAutoSave:              false,
		QuarkTargetFolderID:        "0",
		DoubanHotNavEnabled:        false,
		HotSearchEnabled:           true,
		HomeRankBoardEnabled:       true,
		DoubanCoverProxyURL:        "",
		AutoDeleteInvalidLinks:     false,
		HideInvalidLinksInSearch:   false,
		UpdatedBy:                  0,
	}
	return DB().Create(&cfg).Error
}
