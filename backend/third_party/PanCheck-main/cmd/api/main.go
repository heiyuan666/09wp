package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"PanCheck/internal/checker"
	"PanCheck/internal/config"
	"PanCheck/internal/handler"
	"PanCheck/internal/middleware"
	"PanCheck/internal/model"
	"PanCheck/internal/repository"
	"PanCheck/internal/scheduler"
	"PanCheck/internal/service"
	"PanCheck/pkg/cache"
	"PanCheck/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量文件
	if err := godotenv.Load(".env"); err != nil {
		// .env 文件不存在不是致命错误，只是警告
		// log.Printf("⚠️  未加载.env文件: %v\n", err)
	}

	// 加载配置
	configPath := "configs/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := config.Load(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := database.Init(config.AppConfig.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 初始化Redis缓存
	settingsRepo := repository.NewSettingsRepository()
	redisConfig := loadRedisConfig(settingsRepo)
	cacheRepo, err := cache.NewCacheRepository(cache.CacheConfig{
		Enabled:    redisConfig.Enabled,
		Host:       redisConfig.Host,
		Port:       redisConfig.Port,
		Username:   redisConfig.Username,
		Password:   redisConfig.Password,
		InvalidTTL: redisConfig.InvalidTTL,
	})
	if err != nil {
		log.Printf("Failed to initialize Redis cache: %v, continuing without cache", err)
	}
	if cacheRepo != nil {
		defer cacheRepo.Close()
	}

	// 初始化检测器工厂
	factory := checker.NewCheckerFactory()
	initCheckers(factory, config.AppConfig.Checker)

	// 加载平台缓存TTL配置
	platformTTLMap := loadPlatformTTLMap(settingsRepo)

	// 初始化服务
	linkService := service.NewLinkService()
	checkerService := service.NewCheckerService(factory, cacheRepo, redisConfig.InvalidTTL, platformTTLMap)
	taskService := service.NewScheduledTaskService()
	schedulerService := service.NewTaskSchedulerService(taskService, linkService, checkerService)

	// 初始化调度器
	taskScheduler := scheduler.NewScheduler(taskService, schedulerService)
	if err := taskScheduler.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer taskScheduler.Stop()

	// 初始化处理器
	linkHandler := handler.NewLinkHandler(linkService, checkerService)
	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler()
	settingsHandler := handler.NewSettingsHandler()
	// 设置重新加载检测器配置的函数
	settingsHandler.SetReloadCheckerFunc(func() error {
		// 重新初始化检测器
		initCheckers(factory, config.AppConfig.Checker)
		return nil
	})
	statisticsHandler := handler.NewStatisticsHandler()
	taskHandler := handler.NewScheduledTaskHandler(taskService, schedulerService, taskScheduler)

	// 设置Gin模式
	if config.AppConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.Default()

	// 中间件
	r.Use(middleware.CORS(config.AppConfig.Server.CORSOrigins))

	// 静态文件服务 - 提供前端文件
	r.Static("/assets", "./static/assets")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// 路由
	api := r.Group("/api/v1")
	{
		// 公开接口（不需要认证）
		api.POST("/auth/login", authHandler.Login)
		api.GET("/health", healthHandler.Health)
		api.POST("/links/check", linkHandler.CheckLinks)

		// 需要认证的接口组
		apiAuth := api.Group("")
		apiAuth.Use(middleware.AuthMiddleware())
		{
			// 链接相关接口
			apiAuth.GET("/links/submissions/:id", linkHandler.GetSubmission)
			apiAuth.GET("/links/submissions", linkHandler.ListSubmissions)
			apiAuth.GET("/links/rate-limited", linkHandler.ListRateLimitedLinks)
			apiAuth.DELETE("/links/rate-limited", linkHandler.ClearRateLimitedLinks)

			// 设置相关接口
			apiAuth.GET("/settings/rate-config", settingsHandler.GetRateConfigSettings)
			apiAuth.PUT("/settings/rate-config", settingsHandler.UpdateRateConfigSettings)
			apiAuth.GET("/settings/redis-config", settingsHandler.GetRedisConfig)
			apiAuth.PUT("/settings/redis-config", settingsHandler.UpdateRedisConfig)
			apiAuth.GET("/settings", settingsHandler.GetSettings)
			apiAuth.GET("/settings/:key", settingsHandler.GetSetting)
			apiAuth.PUT("/settings/:key", settingsHandler.UpdateSetting)

			// 统计相关接口
			apiAuth.GET("/statistics/overview", statisticsHandler.GetOverview)
			apiAuth.GET("/statistics/platform-invalid-counts", statisticsHandler.GetPlatformInvalidCounts)
			apiAuth.GET("/statistics/submission-time-series", statisticsHandler.GetSubmissionTimeSeries)

			// 任务计划相关接口
			apiAuth.GET("/scheduled-tasks", taskHandler.ListTasks)
			apiAuth.POST("/scheduled-tasks", taskHandler.CreateTask)
			apiAuth.GET("/scheduled-tasks/:id", taskHandler.GetTask)
			apiAuth.PUT("/scheduled-tasks/:id", taskHandler.UpdateTask)
			apiAuth.DELETE("/scheduled-tasks/:id", taskHandler.DeleteTask)
			apiAuth.POST("/scheduled-tasks/:id/test", taskHandler.TestTaskConfig)
			apiAuth.POST("/scheduled-tasks/test/test", taskHandler.TestTaskConfig) // 直接测试配置
			apiAuth.POST("/scheduled-tasks/:id/run", taskHandler.RunTask)
			apiAuth.POST("/scheduled-tasks/:id/enable", taskHandler.EnableTask)
			apiAuth.POST("/scheduled-tasks/:id/disable", taskHandler.DisableTask)
			apiAuth.GET("/scheduled-tasks/:id/executions", taskHandler.GetTaskExecutions)
			apiAuth.GET("/scheduled-tasks/tags", taskHandler.GetAllTags)
		}
	}

	// 所有非API路由都返回前端应用
	r.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API endpoint not found",
			})
			return
		}
		// 对于其他请求，返回前端应用的index.html
		c.File("./static/index.html")
	})
	r.Static("/images", "./static/images")

	// 启动后台检测服务
	go startBackgroundChecker(checkerService)

	// 启动HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.AppConfig.Server.Port),
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %d\n", config.AppConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// initCheckers 初始化所有检测器
func initCheckers(factory *checker.CheckerFactory, checkerConfig config.CheckerConfig) {
	timeout := time.Duration(checkerConfig.Timeout) * time.Second
	defaultConcurrency := checkerConfig.DefaultConcurrency

	// 从数据库读取各平台的频率配置
	settingsRepo := repository.NewSettingsRepository()

	// 获取频率配置
	getRateConfig := func(platform string) *config.PlatformRateConfig {
		key := fmt.Sprintf("platform_rate_config_%s", platform)
		setting, err := settingsRepo.GetByKey(key)
		if err == nil && setting != nil {
			var rateConfig config.PlatformRateConfig
			if err := json.Unmarshal([]byte(setting.Value), &rateConfig); err == nil {
				return &rateConfig
			}
		}
		// 返回默认配置
		return &config.PlatformRateConfig{
			Concurrency:          defaultConcurrency,
			RequestDelayMs:       0,
			MaxRequestsPerSecond: 0,
		}
	}

	// 创建并注册检测器
	registerChecker := func(platform string) {
		rateConfig := getRateConfig(platform)
		// 使用频率配置中的并发数创建检测器
		concurrency := rateConfig.Concurrency
		if concurrency <= 0 {
			concurrency = defaultConcurrency
		}

		var checkerInstance checker.LinkChecker
		switch platform {
		case "quark":
			checkerInstance = checker.NewQuarkChecker(concurrency, timeout)
		case "uc":
			checkerInstance = checker.NewUCChecker(concurrency, timeout)
		case "baidu":
			checkerInstance = checker.NewBaiduChecker(concurrency, timeout)
		case "tianyi":
			checkerInstance = checker.NewTianyiChecker(concurrency, timeout)
		case "pan123":
			checkerInstance = checker.NewPan123Checker(concurrency, timeout)
		case "pan115":
			checkerInstance = checker.NewPan115Checker(concurrency, timeout)
		case "aliyun":
			checkerInstance = checker.NewAliyunChecker(concurrency, timeout)
		case "xunlei":
			checkerInstance = checker.NewXunleiChecker(concurrency, timeout)
		case "cmcc":
			checkerInstance = checker.NewCMCCChecker(concurrency, timeout)
		default:
			return
		}

		// 设置频率配置
		if baseChecker, ok := checkerInstance.(interface {
			SetRateConfig(*config.PlatformRateConfig)
		}); ok {
			baseChecker.SetRateConfig(rateConfig)
		}
		factory.Register(checkerInstance)
	}

	registerChecker("quark")
	registerChecker("uc")
	registerChecker("baidu")
	registerChecker("tianyi")
	registerChecker("pan123")
	registerChecker("pan115")
	registerChecker("aliyun")
	registerChecker("xunlei")
	registerChecker("cmcc")
}

// loadRedisConfig 从数据库加载Redis配置，如果不存在则使用配置文件中的配置
func loadRedisConfig(settingsRepo *repository.SettingsRepository) config.RedisConfig {
	setting, err := settingsRepo.GetByKey("redis_config")
	if err == nil && setting != nil {
		var redisConfig config.RedisConfig
		if err := json.Unmarshal([]byte(setting.Value), &redisConfig); err == nil {
			return redisConfig
		}
	}
	// 使用配置文件中的配置
	return config.AppConfig.Redis
}

// loadPlatformTTLMap 从数据库加载各平台的缓存TTL配置
func loadPlatformTTLMap(settingsRepo *repository.SettingsRepository) map[model.Platform]int {
	platforms := []string{"quark", "uc", "baidu", "tianyi", "pan123", "pan115", "aliyun", "xunlei", "cmcc"}
	ttlMap := make(map[model.Platform]int)

	for _, platformStr := range platforms {
		key := fmt.Sprintf("platform_rate_config_%s", platformStr)
		setting, err := settingsRepo.GetByKey(key)
		if err == nil && setting != nil {
			var rateConfig config.PlatformRateConfig
			if err := json.Unmarshal([]byte(setting.Value), &rateConfig); err == nil {
				if rateConfig.CacheTTLHours > 0 {
					platform := model.Platform(platformStr)
					ttlMap[platform] = rateConfig.CacheTTLHours
				}
			}
		}
	}

	return ttlMap
}

// startBackgroundChecker 启动后台检测服务
func startBackgroundChecker(checkerService *service.CheckerService) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检测一次
	defer ticker.Stop()

	for range ticker.C {
		if err := checkerService.CheckPendingSubmissions(10); err != nil {
			log.Printf("Background checker error: %v\n", err)
		}
	}
}
