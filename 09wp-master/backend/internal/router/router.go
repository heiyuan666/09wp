package router

import (
	"net/http"

	"dfan-netdisk-backend/internal/handler"
	"dfan-netdisk-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(jwtSecret string) *gin.Engine {
	r := gin.Default()
	r.Static("/public/covers", "./storage/covers")
	r.Static("/public/exports", "./storage/exports")

	// 允许前端开发服务器跨域访问（如 localhost:3007）
	r.Use(middleware.CORSMiddleware())

	// 把 secret 放进上下文，便于 handler 取用
	r.Use(func(c *gin.Context) {
		c.Set("jwt_secret", jwtSecret)
		c.Next()
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger UI（OpenAPI 2.0，由 swag 生成 docs 包后可用）
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 公开 XML（SEO / 订阅）
	r.GET("/sitemap.xml", handler.Sitemap)
	r.GET("/rss.xml", handler.RSS)

	api := r.Group("/api/v1")

	// 前台首页
	api.GET("/home", handler.Home)
	api.GET("/public/config", handler.GetPublicSystemConfig)
	api.GET("/public/version", handler.PublicVersion)
	api.GET("/public/hot-search", handler.GetPublicHotSearch)
	api.GET("/public/douban-hot", handler.GetPublicDoubanHot)
	api.GET("/public/douban/search", handler.PublicDoubanSearch)
	api.GET("/public/tmdb/search", handler.PublicTMDBSearch)
	api.GET("/public/haoka/categories", handler.HaokaCategoryListPublic)
	api.GET("/public/haoka/products", handler.HaokaProductListPublic)
	api.GET("/public/haoka/products/:id", handler.HaokaProductDetailPublic)

	// DFAN Admin 登录接口（系统后台）
	api.POST("/login", handler.AdminLogin(jwtSecret))

	// 前台用户认证（网盘前台）
	api.GET("/auth/captcha", handler.GetCaptcha)
	api.POST("/auth/register/send-code", handler.SendRegisterEmailCode)
	api.POST("/auth/register", handler.UserRegister)
	api.POST("/auth/login", handler.UserLogin)
	api.POST("/auth/qr/create", handler.QRLoginCreate)
	api.GET("/auth/qr/status/:sid", handler.QRLoginStatus)
	api.POST("/auth/qr/confirm", handler.QRLoginConfirm)
	api.POST("/auth/password/forgot", handler.UserPasswordForgot)
	api.POST("/auth/password/reset", handler.UserPasswordReset)

	// 用户信息 & 权限（DFAN Admin）
	adminMeta := api.Group("/users")
	adminMeta.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		adminMeta.GET("/permissions", handler.UserPermissions)
		adminMeta.GET("/info", handler.CurrentUserInfo)
		adminMeta.PUT("/log", handler.AddLoginLog)
	}

	// 前台资源浏览
	api.GET("/resources", handler.ResourceList)
	api.GET("/resources/:id", handler.ResourceDetail)
	api.POST("/resources/:id/access-link", handler.ResourceAccessLink)
	api.GET("/resources/:id/transfer/latest-log", handler.ResourceLatestTransferLog)
	api.GET("/open/netdisk/resources", handler.OpenNetdiskResourceList)
	api.GET("/open/netdisk/resources/:id", handler.OpenNetdiskResourceDetail)
	api.GET("/categories", handler.CategoryListPublic)
	api.GET("/search", handler.ResourceSearch)
	api.GET("/global-search", handler.PublicGlobalSearch)
	api.POST("/global-search/claim", handler.PublicGlobalSearchClaim)
	api.POST("/global-search/get-link", handler.PublicGlobalSearchGetLink)
	api.POST("/feedbacks", handler.FeedbackCreate)
	api.GET("/game/category/list", handler.GameCategoryList)
	api.GET("/game/list", handler.GameList)
	api.GET("/game/detail/:id", handler.GameDetail)
	api.GET("/game/resource/list", handler.GameResourceList)
	api.POST("/game/feedbacks", handler.GameFeedbackCreate)
	api.GET("/game/reviews", handler.GameReviewList)
	api.POST("/game/reviews", middleware.AuthMiddleware(jwtSecret, false), handler.GameReviewCreate)
	api.POST("/game/reviews/:id/vote", middleware.AuthMiddleware(jwtSecret, false), handler.GameReviewVote)
	api.GET("/game/public/config", handler.GetPublicGameSiteConfig)
	api.GET("/game/public/nav-menus", handler.PublicGameNavMenus)
	api.GET("/software/categories", handler.PublicSoftwareCategoryList)
	api.GET("/software/public/config", handler.GetPublicSoftwareSiteConfig)
	api.GET("/software/list", handler.PublicSoftwareList)
	api.GET("/software/detail/:id", handler.PublicSoftwareDetail)
	api.POST("/quark/transfer", middleware.AuthMiddleware(jwtSecret, true), handler.QuarkTransferByLink)
	api.POST("/netdisk/transfer", middleware.AuthMiddleware(jwtSecret, true), handler.NetdiskTransferByLink)
	api.POST("/netdisk/transfer/batch", middleware.AuthMiddleware(jwtSecret, true), handler.NetdiskTransferBatchByLinks)

	// 网盘前台用户登录后接口
	userGroup := api.Group("/user")
	userGroup.Use(middleware.AuthMiddleware(jwtSecret, false))
	{
		userGroup.GET("/profile", handler.UserProfile)
		userGroup.PUT("/password", handler.UserChangePassword)
		userGroup.GET("/favorites", handler.UserFavoriteList)
		userGroup.POST("/favorites/:resource_id", handler.UserFavoriteAdd)
		userGroup.DELETE("/favorites/:resource_id", handler.UserFavoriteRemove)
		userGroup.POST("/submissions", handler.UserSubmissionCreate)
		userGroup.GET("/submissions", handler.UserSubmissionMyList)
	}

	// 系统后台用户接口（DFAN Admin），需要管理员权限
	systemUser := api.Group("/users")
	systemUser.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		systemUser.GET("", handler.AdminUserList)
		systemUser.POST("", handler.AdminCreateUser)
		systemUser.GET("/:id", handler.AdminUserDetail)
		systemUser.PUT("", handler.AdminUpdateUser)
		systemUser.DELETE("", handler.AdminUserBatchDelete)

		systemUser.PUT("/profile", handler.UpdateProfile)
		systemUser.PUT("/password", handler.AdminChangePassword)
		systemUser.PUT("/avatar", handler.UpdateAvatar)
	}

	// 系统后台角色接口（DFAN Admin），需要管理员权限
	systemRole := api.Group("/roles")
	systemRole.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		systemRole.GET("", handler.RolePage)
		systemRole.GET("/:id", handler.RoleInfo)
		systemRole.PUT("", handler.RoleUpdate)
		systemRole.DELETE("", handler.RoleDelete)
	}

	// 系统配置接口（全局配置）
	systemConfig := api.Group("/system/config")
	systemConfig.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		systemConfig.GET("", handler.GetSystemConfig)
		systemConfig.PUT("", handler.UpdateSystemConfig)
		systemConfig.POST("/meili/test", handler.AdminMeiliTest)
		systemConfig.POST("/meili/reindex", handler.AdminMeiliReindex)
		systemConfig.POST("/global-search/test", handler.AdminGlobalSearchTest)
	}
	globalSearchAdmin := api.Group("/system/global-search")
	globalSearchAdmin.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		globalSearchAdmin.GET("/settings", handler.AdminGlobalSearchSettingsGet)
		globalSearchAdmin.PUT("/settings", handler.AdminGlobalSearchSettingsPut)
	}
	globalSearchAPIs := api.Group("/system/global-search/apis")
	globalSearchAPIs.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		globalSearchAPIs.GET("", handler.AdminGlobalSearchAPIList)
		globalSearchAPIs.POST("", handler.AdminGlobalSearchAPICreate)
		globalSearchAPIs.PUT("/:id", handler.AdminGlobalSearchAPIUpdate)
		globalSearchAPIs.DELETE("/:id", handler.AdminGlobalSearchAPIDelete)
	}

	netdiskCred := api.Group("/system/netdisk-credentials")
	netdiskCred.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		netdiskCred.GET("", handler.GetNetdiskCredentials)
		netdiskCred.PUT("", handler.UpdateNetdiskCredentials)
	}

	// 导航菜单管理（DFAN Admin）
	navMenus := api.Group("/nav-menus")
	navMenus.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		navMenus.GET("", handler.NavMenuList)
		navMenus.POST("", handler.NavMenuCreate)
		navMenus.PUT("/:id", handler.NavMenuUpdate)
		navMenus.DELETE("/:id", handler.NavMenuDelete)
	}

	// 导航菜单（前台公开）
	public := api.Group("/public")
	{
		public.GET("/nav-menus", handler.PublicNavMenus)
	}

	// 游戏站点设置（管理端，需管理员权限）
	gameConfig := api.Group("/game/config")
	gameConfig.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		gameConfig.GET("", handler.GetGameSiteConfig)
		gameConfig.PUT("", handler.UpdateGameSiteConfig)
	}

	// 游戏导航菜单管理（管理端，需管理员权限）
	gameNavMenus := api.Group("/game/nav-menus")
	gameNavMenus.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		gameNavMenus.GET("", handler.GameNavMenuList)
		gameNavMenus.POST("", handler.GameNavMenuCreate)
		gameNavMenus.PUT("/:id", handler.GameNavMenuUpdate)
		gameNavMenus.DELETE("/:id", handler.GameNavMenuDelete)
	}

	// TG 频道管理（独立模块）
	tgChannels := api.Group("/telegram/channels")
	tgChannels.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		tgChannels.GET("", handler.TelegramChannelList)
		tgChannels.POST("", handler.TelegramChannelCreate)
		tgChannels.POST("/test", handler.TelegramChannelTest)
		tgChannels.PUT("/:id", handler.TelegramChannelUpdate)
		tgChannels.DELETE("/:id", handler.TelegramChannelDelete)
		tgChannels.POST("/:id/sync", handler.TelegramChannelSync)
		tgChannels.POST("/:id/backfill", handler.TelegramChannelBackfill)
		tgChannels.POST("/sync-all", handler.TelegramChannelSyncAll)
	}

	tgSession := api.Group("/telegram/session")
	tgSession.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		tgSession.GET("/status", handler.TelegramSessionStatus)
		tgSession.POST("/send-code", handler.TelegramSessionSendCode)
		tgSession.POST("/sign-in", handler.TelegramSessionSignIn)
		tgSession.POST("/check-password", handler.TelegramSessionCheckPassword)
	}

	rssSubs := api.Group("/rss/subscriptions")
	rssSubs.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		rssSubs.GET("", handler.RSSSubscriptionList)
		rssSubs.POST("", handler.RSSSubscriptionCreate)
		rssSubs.POST("/test", handler.RSSSubscriptionTest)
		rssSubs.PUT("/:id", handler.RSSSubscriptionUpdate)
		rssSubs.DELETE("/:id", handler.RSSSubscriptionDelete)
		rssSubs.POST("/:id/sync", handler.RSSSubscriptionSync)
		rssSubs.POST("/sync-all", handler.RSSSubscriptionSyncAll)
	}

	// 后台资源与分类管理（仍然使用 admin 前缀）
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		// 控制台统计
		adminGroup.GET("/stats", handler.AdminStats)

		// 号卡（外部套餐同步/本地管理）
		adminGroup.GET("/haoka/categories", handler.AdminHaokaCategories)
		adminGroup.POST("/haoka/query-products", handler.AdminHaokaQueryProducts)
		adminGroup.POST("/haoka/sync", handler.AdminHaokaSync)
		adminGroup.GET("/haoka/products", handler.AdminHaokaProductList)
		adminGroup.POST("/haoka/products", handler.AdminHaokaProductCreate)
		adminGroup.POST("/haoka/products/upsert", handler.AdminHaokaProductUpsertFromExternal)
		adminGroup.GET("/haoka/products/:id", handler.AdminHaokaProductDetail)
		adminGroup.PUT("/haoka/products/:id", handler.AdminHaokaProductUpdate)
		adminGroup.PUT("/haoka/products/:id/flag", handler.AdminHaokaProductSetFlag)

		// 分类管理
		adminGroup.GET("/categories", handler.AdminCategoryList)
		adminGroup.POST("/categories", handler.AdminCategoryCreate)
		adminGroup.PUT("/categories/:id", handler.AdminCategoryUpdate)
		adminGroup.DELETE("/categories/:id", handler.AdminCategoryDelete)
		adminGroup.PUT("/categories/:id/status", handler.AdminCategoryChangeStatus)
		adminGroup.PUT("/categories/:id/sort", handler.AdminCategoryChangeSort)
		// 资源管理
		adminGroup.GET("/resources", handler.AdminResourceList)
		adminGroup.POST("/resources", handler.AdminResourceCreate)
		// 资源表格导入导出（CSV/XLSX）
		// multipart/form-data: file
		adminGroup.POST("/resources/import-table", handler.AdminResourceImportTable)
		// GET 导出表格；并创建一条“导出文件资源”记录到 resources 表
		adminGroup.GET("/resources/export-table", handler.AdminResourceExportTable)
		adminGroup.POST("/resources/sync-telegram", handler.AdminResourceSyncTelegram)
		adminGroup.PUT("/resources/:id", handler.AdminResourceUpdate)
		adminGroup.DELETE("/resources/:id", handler.AdminResourceDelete)
		adminGroup.POST("/resources/:id/retry-transfer", handler.AdminResourceRetryTransfer)
		adminGroup.GET("/resources/:id/transfer-logs", handler.AdminResourceTransferLogs)
		adminGroup.POST("/resources/batch-delete", handler.AdminResourceBatchDelete)
		adminGroup.POST("/resources/batch-status", handler.AdminResourceBatchStatus)
		adminGroup.POST("/resources/check-links", handler.AdminResourceCheckLinks)
		adminGroup.POST("/pancheck/check", handler.PanCheckLinks)

		// 关键词屏蔽管理
		adminGroup.GET("/keyword-blocks", handler.AdminKeywordBlockList)
		adminGroup.POST("/keyword-blocks", handler.AdminKeywordBlockCreate)
		adminGroup.PUT("/keyword-blocks/:id", handler.AdminKeywordBlockUpdate)
		adminGroup.DELETE("/keyword-blocks/:id", handler.AdminKeywordBlockDelete)

		// 反馈管理
		adminGroup.GET("/feedbacks", handler.AdminFeedbackList)
		adminGroup.PUT("/feedbacks/:id/status", handler.AdminFeedbackUpdateStatus)
		adminGroup.GET("/cleanup-logs", handler.AdminCleanupLogList)
		// 游戏资源反馈（标记失效等）
		adminGroup.GET("/game-feedbacks", handler.AdminGameFeedbackList)
		adminGroup.PUT("/game-feedbacks/:id/status", handler.AdminGameFeedbackUpdateStatus)

		// 用户提交资源（审核）
		adminGroup.GET("/submissions", handler.AdminSubmissionList)
		adminGroup.POST("/submissions/:id/approve", handler.AdminSubmissionApprove)
		adminGroup.POST("/submissions/:id/reject", handler.AdminSubmissionReject)
	}

	// 游戏后台管理（管理员权限）
	gameAdmin := api.Group("/game")
	gameAdmin.Use(middleware.AuthMiddleware(jwtSecret, true))
	{
		// 分类
		gameAdmin.POST("/category/create", handler.GameCategoryCreate)
		gameAdmin.PUT("/category/:id", handler.GameCategoryUpdate)
		gameAdmin.DELETE("/category/:id", handler.GameCategoryDelete)

		// 游戏
		gameAdmin.GET("/steam/search", handler.GameSteamSearch)
		gameAdmin.GET("/steam/app/:appid", handler.GameSteamAppDetail)
		gameAdmin.POST("/create", handler.GameCreate)
		gameAdmin.PUT("/:id", handler.GameUpdate)
		gameAdmin.DELETE("/:id", handler.GameDelete)

		// 下载资源
		gameAdmin.POST("/resource/create", handler.GameResourceCreate)
		gameAdmin.PUT("/resource/:id", handler.GameResourceUpdate)
		gameAdmin.DELETE("/resource/:id", handler.GameResourceDelete)

		// 封面/截图上传
		gameAdmin.POST("/upload", handler.GameUpload)
		gameAdmin.POST("/software/upload-cover", handler.SoftwareUploadCover)

		gameAdmin.GET("/software/site-config", handler.GetSoftwareSiteConfig)
		gameAdmin.PUT("/software/site-config", handler.UpdateSoftwareSiteConfig)

		// 软件分类
		gameAdmin.GET("/software/categories", handler.SoftwareCategoryList)
		gameAdmin.POST("/software/categories", handler.SoftwareCategoryCreate)
		gameAdmin.PUT("/software/categories/:id", handler.SoftwareCategoryUpdate)
		gameAdmin.DELETE("/software/categories/:id", handler.SoftwareCategoryDelete)
		gameAdmin.PUT("/software/categories/:id/sort", handler.SoftwareCategorySort)

		// 软件
		gameAdmin.GET("/software", handler.SoftwareList)
		gameAdmin.GET("/software/:id", handler.SoftwareDetail)
		gameAdmin.POST("/software", handler.SoftwareCreate)
		gameAdmin.PUT("/software/:id", handler.SoftwareUpdate)
		gameAdmin.DELETE("/software/:id", handler.SoftwareDelete)

		// 软件版本
		gameAdmin.GET("/software/:id/versions", handler.SoftwareVersionList)
		gameAdmin.POST("/software/:id/versions", handler.SoftwareVersionCreate)
		gameAdmin.PUT("/software/versions/:version_id", handler.SoftwareVersionUpdate)
		gameAdmin.DELETE("/software/versions/:version_id", handler.SoftwareVersionDelete)

		// 评论管理（避免与前台 /api/v1/game/reviews 冲突，统一挂到 /game/admin/reviews）
		gameAdmin.GET("/admin/reviews", handler.AdminGameReviewList)
		gameAdmin.PUT("/admin/reviews/:id/status", handler.AdminGameReviewSetStatus)
		gameAdmin.DELETE("/admin/reviews/:id", handler.AdminGameReviewDelete)
	}

	return r
}
