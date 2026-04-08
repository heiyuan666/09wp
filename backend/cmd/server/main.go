// @title           DFAN Netdisk Backend API
// @version         1.0
// @description     网盘后台与前台 HTTP API。完整路径以 /api/v1 为前缀；文档由 swag 从注释生成，可随开发逐步补全 @Router 注释。
// @BasePath        /api/v1
// @host            localhost:8080
// @schemes         http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description     登录后获得的 JWT；请求头 Authorization 值为 Bearer 加空格再加 token
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	_ "dfan-netdisk-backend/docs"

	"dfan-netdisk-backend/internal/config"
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/router"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/internal/version"
)

func envTruthy(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func main() {
	debug := flag.Bool("debug", false, "终端调试：输出网盘转存/UC 本人分享等详细日志（等价于设置 NETDISK_TRANSFER_DEBUG=1）")
	flag.Parse()
	if *debug || envTruthy("DEBUG") || envTruthy("APP_DEBUG") {
		_ = os.Setenv("NETDISK_TRANSFER_DEBUG", "1")
		_ = os.Setenv("UC_TRANSFER_DEBUG", "1")
		log.Println("[debug] 已开启：NETDISK_TRANSFER_DEBUG / UC_TRANSFER_DEBUG（转存与本人分享链路会打印 [UC-OWN-SHARE] 等日志）")
	}

	// 读取配置（从环境变量为主，后续可扩展成配置文件）
	cfg := config.Load()

	// 初始化数据库
	if err := database.InitMySQL(cfg.MySQLDSN); err != nil {
		log.Fatalf("init mysql failed: %v", err)
	}

	// 自动迁移核心表结构（开发环境方便快速起步）
	if err := database.AutoMigrate(); err != nil {
		log.Printf("auto migrate failed: %v", err)
	}

	// 初始化管理员账号（默认 admin/123456）
	if err := database.SeedAdmin(); err != nil {
		log.Printf("seed admin failed: %v", err)
	}

	// 初始化默认菜单
	if err := database.SeedMenus(); err != nil {
		log.Printf("seed menus failed: %v", err)
	}

	// 初始化默认角色
	if err := database.SeedRoles(); err != nil {
		log.Printf("seed roles failed: %v", err)
	}

	// 初始化全局系统配置
	if err := database.SeedSystemConfig(); err != nil {
		log.Printf("seed system config failed: %v", err)
	}

	if err := database.SeedNetdiskCredential(); err != nil {
		log.Printf("seed netdisk credential failed: %v", err)
	}
	if err := database.SeedGameCategories(); err != nil {
		log.Printf("seed game categories failed: %v", err)
	}

	// 初始化 Gin 路由
	r := router.SetupRouter(cfg.JWTSecret)
	service.SetPanCheckBaseURL(cfg.PanCheckBaseURL)

	// 初始化 Redis 搜索缓存（可选）
	if err := service.InitSearchRedisCache(cfg.Redis); err != nil {
		log.Printf("init redis cache disabled: %v", err)
	}

	// 启动 TG 自动同步后台任务
	service.StartTelegramSyncWorker()
	service.StartTelegramAutoTransferWorker()
	service.StartResourceCheckWorker()
	service.StartHaokaSyncWorker()
	service.StartRSSSyncWorker()

	addr := ":" + cfg.HTTPPort
	if env := os.Getenv("PORT"); env != "" {
		addr = ":" + env
	}

	log.Printf("dfan-netdisk-backend %s listening on %s", version.Version, addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
