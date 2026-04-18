package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/config"
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/meilisearch/meilisearch-go"
)

// FriendLinkItem 友情链接（标题 + 链接）
type FriendLinkItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type publicSystemConfig struct {
	SiteTitle                  string           `json:"site_title"`
	SupportEmail               string           `json:"support_email"`
	ContactPhone               string           `json:"contact_phone"`
	ContactQQ                  string           `json:"contact_qq"`
	LogoURL                    string           `json:"logo_url"`
	FaviconURL                 string           `json:"favicon_url"`
	SeoKeywords                string           `json:"seo_keywords"`
	SeoDescription             string           `json:"seo_description"`
	IcpRecord                  string           `json:"icp_record"`
	FooterText                 string           `json:"footer_text"`
	ClarityProjectID           string           `json:"clarity_project_id"`
	ClarityEnabled             bool             `json:"clarity_enabled"`
	AllowRegister              bool             `json:"allow_register"`
	SubmissionNeedReview       bool             `json:"submission_need_review"`
	SubmissionAutoTransfer     bool             `json:"submission_auto_transfer"`
	ResourceDetailAutoTransfer bool             `json:"resource_detail_auto_transfer"`
	ResourceDetailEachClickFreshShare bool      `json:"resource_detail_each_click_fresh_share"`
	HaokaOrderURL              string           `json:"haoka_order_url"`
	HaokaAgentRegURL           string           `json:"haoka_agent_reg_url"`
	FriendLinks                []FriendLinkItem `json:"friend_links"`
	FooterQuickLinks           []FriendLinkItem `json:"footer_quick_links"`
	FooterHotPlatforms         []string         `json:"footer_hot_platforms"`
	FooterSocialLinks          []FriendLinkItem `json:"footer_social_links"`
	FooterWechat               string           `json:"footer_wechat"`
	DoubanHotNavEnabled        bool             `json:"douban_hot_nav_enabled"`
	HotSearchEnabled           bool             `json:"hot_search_enabled"`
	ShowSiteTitle              bool             `json:"show_site_title"`
	HomeRankBoardEnabled       bool             `json:"home_rank_board_enabled"`
	DoubanCoverProxyURL        string           `json:"douban_cover_proxy_url"`
	TgImageProxyURL            string           `json:"tg_image_proxy_url"`
	ThunderDownloadEnabled     bool             `json:"thunder_download_enabled"`
	// GlobalSearchCloudTypes 全网搜默认网盘筛选（逗号分隔），前台搜索页与未传 cloud_types 时一致
	GlobalSearchCloudTypes string `json:"global_search_cloud_types"`
}

// systemConfigOut 管理端返回：附带解析后的 friend_links
type systemConfigOut struct {
	model.SystemConfig
	FriendLinks        []FriendLinkItem `json:"friend_links"`
	FooterQuickLinks   []FriendLinkItem `json:"footer_quick_links"`
	FooterHotPlatforms []string         `json:"footer_hot_platforms"`
	FooterSocialLinks  []FriendLinkItem `json:"footer_social_links"`
}

// systemConfigPut 更新请求：friend_links 为数组，入库时序列化为 JSON 字符串
type systemConfigPut struct {
	model.SystemConfig
	FriendLinks        []FriendLinkItem `json:"friend_links"`
	FooterQuickLinks   []FriendLinkItem `json:"footer_quick_links"`
	FooterHotPlatforms []string         `json:"footer_hot_platforms"`
	FooterSocialLinks  []FriendLinkItem `json:"footer_social_links"`
}

func parseFriendLinksJSON(s string) []FriendLinkItem {
	s = strings.TrimSpace(s)
	if s == "" {
		return []FriendLinkItem{}
	}
	var out []FriendLinkItem
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return []FriendLinkItem{}
	}
	return out
}

func normalizeFriendLinks(items []FriendLinkItem) ([]FriendLinkItem, string) {
	const maxN = 50
	if len(items) > maxN {
		items = items[:maxN]
	}
	out := make([]FriendLinkItem, 0, len(items))
	for _, it := range items {
		title := strings.TrimSpace(it.Title)
		url := strings.TrimSpace(it.URL)
		if title == "" && url == "" {
			continue
		}
		if len(title) > 120 {
			title = title[:120]
		}
		if len(url) > 500 {
			url = url[:500]
		}
		out = append(out, FriendLinkItem{Title: title, URL: url})
	}
	raw, _ := json.Marshal(out)
	return out, string(raw)
}

func parseStringListJSON(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return []string{}
	}
	return out
}

func normalizeStringList(items []string) ([]string, string) {
	const maxN = 50
	if len(items) > maxN {
		items = items[:maxN]
	}
	out := make([]string, 0, len(items))
	for _, it := range items {
		v := strings.TrimSpace(it)
		if v == "" {
			continue
		}
		if len(v) > 120 {
			v = v[:120]
		}
		out = append(out, v)
	}
	raw, _ := json.Marshal(out)
	return out, string(raw)
}

// GetSystemConfig 获取全局系统配置
func GetSystemConfig(c *gin.Context) {
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		response.Error(c, 404, "系统配置不存在")
		return
	}
	links := parseFriendLinksJSON(cfg.FriendLinks)
	response.OK(c, systemConfigOut{
		SystemConfig:       cfg,
		FriendLinks:        links,
		FooterQuickLinks:   parseFriendLinksJSON(cfg.FooterQuickLinks),
		FooterHotPlatforms: parseStringListJSON(cfg.FooterHotPlatforms),
		FooterSocialLinks:  parseFriendLinksJSON(cfg.FooterSocialLinks),
	})
}

// GetPublicSystemConfig 获取前台可用系统配置（无需登录）
func GetPublicSystemConfig(c *gin.Context) {
	cacheKey := "public:system-config:v5"
	if b, ok := service.GetSearchCache(context.Background(), cacheKey); ok {
		var cached publicSystemConfig
		if err := json.Unmarshal(b, &cached); err == nil {
			response.OK(c, cached)
			return
		}
	}

	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		response.Error(c, 404, "系统配置不存在")
		return
	}
	links := parseFriendLinksJSON(cfg.FriendLinks)
	footerQuickLinks := parseFriendLinksJSON(cfg.FooterQuickLinks)
	footerHotPlatforms := parseStringListJSON(cfg.FooterHotPlatforms)
	footerSocialLinks := parseFriendLinksJSON(cfg.FooterSocialLinks)
	response.OK(c, publicSystemConfig{
		SiteTitle:                  cfg.SiteTitle,
		SupportEmail:               cfg.SupportEmail,
		ContactPhone:               cfg.ContactPhone,
		ContactQQ:                  cfg.ContactQQ,
		LogoURL:                    cfg.LogoURL,
		FaviconURL:                 cfg.FaviconURL,
		SeoKeywords:                cfg.SeoKeywords,
		SeoDescription:             cfg.SeoDescription,
		IcpRecord:                  cfg.IcpRecord,
		FooterText:                 cfg.FooterText,
		ClarityProjectID:           cfg.ClarityProjectID,
		ClarityEnabled:             cfg.ClarityEnabled,
		AllowRegister:              cfg.AllowRegister,
		SubmissionNeedReview:       cfg.SubmissionNeedReview,
		SubmissionAutoTransfer:            cfg.SubmissionAutoTransfer,
		ResourceDetailAutoTransfer:        cfg.ResourceDetailAutoTransfer,
		ResourceDetailEachClickFreshShare: cfg.ResourceDetailEachClickFreshShare,
		HaokaOrderURL:                     cfg.HaokaOrderURL,
		HaokaAgentRegURL:                  cfg.HaokaAgentRegURL,
		FriendLinks:                       links,
		FooterQuickLinks:                  footerQuickLinks,
		FooterHotPlatforms:                footerHotPlatforms,
		FooterSocialLinks:                 footerSocialLinks,
		FooterWechat:                      cfg.FooterWechat,
		DoubanHotNavEnabled:               cfg.DoubanHotNavEnabled,
		HotSearchEnabled:                  cfg.HotSearchEnabled,
		ShowSiteTitle:                     cfg.ShowSiteTitle,
		HomeRankBoardEnabled:              cfg.HomeRankBoardEnabled,
		DoubanCoverProxyURL:               cfg.DoubanCoverProxyURL,
		TgImageProxyURL:                   cfg.TgImageProxyURL,
		ThunderDownloadEnabled:            cfg.ThunderDownloadEnabled,
		GlobalSearchCloudTypes:            strings.TrimSpace(cfg.GlobalSearchCloudTypes),
	})

	// 写入缓存（短 TTL）
	if raw, err := json.Marshal(publicSystemConfig{
		SiteTitle:                         cfg.SiteTitle,
		SupportEmail:                      cfg.SupportEmail,
		ContactPhone:                      cfg.ContactPhone,
		ContactQQ:                         cfg.ContactQQ,
		LogoURL:                           cfg.LogoURL,
		FaviconURL:                        cfg.FaviconURL,
		SeoKeywords:                       cfg.SeoKeywords,
		SeoDescription:                    cfg.SeoDescription,
		IcpRecord:                         cfg.IcpRecord,
		FooterText:                        cfg.FooterText,
		ClarityProjectID:                  cfg.ClarityProjectID,
		ClarityEnabled:                    cfg.ClarityEnabled,
		AllowRegister:                     cfg.AllowRegister,
		SubmissionNeedReview:              cfg.SubmissionNeedReview,
		SubmissionAutoTransfer:            cfg.SubmissionAutoTransfer,
		ResourceDetailAutoTransfer:        cfg.ResourceDetailAutoTransfer,
		ResourceDetailEachClickFreshShare: cfg.ResourceDetailEachClickFreshShare,
		HaokaOrderURL:                     cfg.HaokaOrderURL,
		HaokaAgentRegURL:           cfg.HaokaAgentRegURL,
		FriendLinks:                links,
		FooterQuickLinks:           footerQuickLinks,
		FooterHotPlatforms:         footerHotPlatforms,
		FooterSocialLinks:          footerSocialLinks,
		FooterWechat:               cfg.FooterWechat,
		DoubanHotNavEnabled:        cfg.DoubanHotNavEnabled,
		HotSearchEnabled:           cfg.HotSearchEnabled,
		ShowSiteTitle:              cfg.ShowSiteTitle,
		HomeRankBoardEnabled:       cfg.HomeRankBoardEnabled,
		DoubanCoverProxyURL:        cfg.DoubanCoverProxyURL,
		TgImageProxyURL:            cfg.TgImageProxyURL,
		ThunderDownloadEnabled:     cfg.ThunderDownloadEnabled,
		GlobalSearchCloudTypes:     strings.TrimSpace(cfg.GlobalSearchCloudTypes),
	}); err == nil {
		service.SetSearchCache(context.Background(), cacheKey, raw)
	}
}

// UpdateSystemConfig 更新全局系统配置
func UpdateSystemConfig(c *gin.Context) {
	var req systemConfigPut
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		response.Error(c, 404, "系统配置不存在")
		return
	}

	uidVal, _ := c.Get("user_id")
	uid, _ := uidVal.(uint64)

	siteTitle := req.SiteTitle
	if siteTitle == "" {
		siteTitle = "网盘资源导航系统"
	}
	if req.TgSyncInterval < 30 {
		req.TgSyncInterval = 30
	}
	if req.HaokaSyncInterval < 300 {
		req.HaokaSyncInterval = 300
	}
	if req.LinkCheckInterval < 60 {
		req.LinkCheckInterval = 3600
	}

	_, friendLinksStr := normalizeFriendLinks(req.FriendLinks)
	_, footerQuickLinksStr := normalizeFriendLinks(req.FooterQuickLinks)
	_, footerHotPlatformsStr := normalizeStringList(req.FooterHotPlatforms)
	_, footerSocialLinksStr := normalizeFriendLinks(req.FooterSocialLinks)

	updates := map[string]interface{}{
		"site_title":                                 siteTitle,
		"admin_email":                                req.AdminEmail,
		"support_email":                              req.SupportEmail,
		"contact_phone":                              req.ContactPhone,
		"contact_qq":                                 req.ContactQQ,
		"logo_url":                                   req.LogoURL,
		"favicon_url":                                req.FaviconURL,
		"seo_keywords":                               req.SeoKeywords,
		"seo_description":                            req.SeoDescription,
		"icp_record":                                 req.IcpRecord,
		"footer_text":                                req.FooterText,
		"clarity_project_id":                         strings.TrimSpace(req.ClarityProjectID),
		"clarity_enabled":                            req.ClarityEnabled,
		"friend_links":                               friendLinksStr,
		"footer_quick_links":                         footerQuickLinksStr,
		"footer_hot_platforms":                       footerHotPlatformsStr,
		"footer_social_links":                        footerSocialLinksStr,
		"footer_wechat":                              strings.TrimSpace(req.FooterWechat),
		"allow_register":                             req.AllowRegister,
		"submission_need_review":                     req.SubmissionNeedReview,
		"submission_auto_transfer":                   req.SubmissionAutoTransfer,
		"resource_detail_auto_transfer":              req.ResourceDetailAutoTransfer,
		"resource_detail_each_click_fresh_share":    req.ResourceDetailEachClickFreshShare,
		"haoka_user_id":                              req.HaokaUserID,
		"haoka_secret":                               req.HaokaSecret,
		"haoka_sync_enabled":                         req.HaokaSyncEnabled,
		"haoka_sync_interval":                        req.HaokaSyncInterval,
		"haoka_order_url":                            req.HaokaOrderURL,
		"haoka_agent_reg_url":                        req.HaokaAgentRegURL,
		"smtp_host":                                  req.SmtpHost,
		"smtp_port":                                  req.SmtpPort,
		"smtp_user":                                  req.SmtpUser,
		"smtp_pass":                                  req.SmtpPass,
		"smtp_from":                                  req.SmtpFrom,
		"tg_bot_token":                               req.TgBotToken,
		"tg_proxy_url":                               req.TgProxyURL,
		"tg_api_id":                                  req.TgAPIID,
		"tg_api_hash":                                req.TgAPIHash,
		"tg_session":                                 req.TgSession,
		"pancheck_base_url":                          req.PanCheckBaseURL,
		"link_check_enabled":                         req.LinkCheckEnabled,
		"link_check_interval":                        req.LinkCheckInterval,
		"douban_hot_nav_enabled":                     req.DoubanHotNavEnabled,
		"hot_search_enabled":                         req.HotSearchEnabled,
		"show_site_title":                            req.ShowSiteTitle,
		"home_rank_board_enabled":                    req.HomeRankBoardEnabled,
		"douban_cover_proxy_url":                     req.DoubanCoverProxyURL,
		"tg_image_proxy_url":                         req.TgImageProxyURL,
		"douban_search_cache_ttl":                    req.DoubanSearchCacheTTL,
		"douban_search_enabled":                      req.DoubanSearchEnabled,
		"tmdb_bearer_token":                          strings.TrimSpace(req.TMDBBearerToken),
		"tmdb_search_cache_ttl":                      req.TMDBSearchCacheTTL,
		"tmdb_proxy_url":                             strings.TrimSpace(req.TMDBProxyURL),
		"tmdb_enabled":                               req.TMDBEnabled,
		"iyuns_api_base_url": strings.TrimSpace(req.IYunsAPIBaseURL),
		// 全网搜 * 仅通过 PUT /system/global-search/settings 维护，避免系统配置页保存时覆盖
		"auto_delete_invalid_links": req.AutoDeleteInvalidLinks,
		"hide_invalid_links_in_search":               req.HideInvalidLinksInSearch,
		"thunder_download_enabled":                   req.ThunderDownloadEnabled,
		"quark_cleanup_enabled":                      req.QuarkCleanupEnabled,
		"quark_cleanup_folder_id":                    strings.TrimSpace(req.QuarkCleanupFolderID),
		"quark_cleanup_older_than_minutes":           req.QuarkCleanupOlderThanMinutes,
		"quark_cleanup_interval_minutes":             req.QuarkCleanupIntervalMinutes,
		"meili_enabled":                              req.MeiliEnabled,
		"meili_url":                                  strings.TrimSpace(req.MeiliURL),
		"meili_api_key":                              strings.TrimSpace(req.MeiliAPIKey),
		"meili_index_name":                           strings.TrimSpace(req.MeiliIndexName),
		"tg_channel_chat_id":                         req.TgChannelChatID,
		"tg_sync_enabled":                            req.TgSyncEnabled,
		"tg_sync_interval":                           req.TgSyncInterval,
		"tg_default_cat_id":                          req.TgDefaultCatID,
		"updated_by":                                 uid,
	}

	if err := database.DB().Model(&cfg).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}

	// 配置更新后失效前台缓存，避免前台获取旧配置
	service.DeleteSearchCache(context.Background(), "public:system-config:v3")
	service.DeleteSearchCache(context.Background(), "public:system-config:v4")
	service.DeleteSearchCache(context.Background(), "public:system-config:v5")

	// 尝试热更新 Meilisearch 客户端（失败不影响保存）
	_ = service.InitMeili(config.MeiliConfig{
		Enabled:    req.MeiliEnabled,
		URL:        strings.TrimSpace(req.MeiliURL),
		APIKey:     strings.TrimSpace(req.MeiliAPIKey),
		Index:      strings.TrimSpace(req.MeiliIndexName),
		TimeoutMS:  2500,
		PrimaryKey: "id",
	})
	response.OK(c, nil)
}

// AdminMeiliTest 测试 Meilisearch 连接与索引可用性
func AdminMeiliTest(c *gin.Context) {
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		response.Error(c, 404, "系统配置不存在")
		return
	}
	if !cfg.MeiliEnabled {
		response.OK(c, gin.H{"enabled": false, "ok": false, "message": "Meilisearch 未开启"})
		return
	}
	host := strings.TrimSpace(cfg.MeiliURL)
	if host == "" {
		response.OK(c, gin.H{"enabled": true, "ok": false, "message": "Meilisearch URL 为空"})
		return
	}

	client := meilisearch.New(host,
		meilisearch.WithAPIKey(strings.TrimSpace(cfg.MeiliAPIKey)),
		meilisearch.WithCustomClient(&http.Client{Timeout: 3 * time.Second}),
	)
	health, err := client.Health()
	if err != nil {
		response.OK(c, gin.H{"enabled": true, "ok": false, "message": "连接失败: " + err.Error()})
		return
	}
	idxName := strings.TrimSpace(cfg.MeiliIndexName)
	if idxName == "" {
		idxName = "resources"
	}
	_, err = client.GetIndex(idxName)
	if err != nil {
		// 索引不存在也算“可连接”，但提示用户需要重建
		response.OK(c, gin.H{
			"enabled": true,
			"ok":      true,
			"health":  health,
			"index":   idxName,
			"message": "连接正常，但索引不存在（可点击重建索引）",
		})
		return
	}
	response.OK(c, gin.H{
		"enabled": true,
		"ok":      true,
		"health":  health,
		"index":   idxName,
		"message": "连接正常",
	})
}

// AdminMeiliReindex 全量重建 Meili 索引（从 MySQL 导入）
func AdminMeiliReindex(c *gin.Context) {
	// 每次重建前按 DB 系统配置尝试初始化 Meili，避免后端启动时 meili=disabled 导致无法重建
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		response.Error(c, 404, "系统配置不存在")
		return
	}
	_ = service.InitMeili(config.MeiliConfig{
		Enabled:    cfg.MeiliEnabled,
		URL:        strings.TrimSpace(cfg.MeiliURL),
		APIKey:     strings.TrimSpace(cfg.MeiliAPIKey),
		Index:      strings.TrimSpace(cfg.MeiliIndexName),
		TimeoutMS:  2500,
		PrimaryKey: "id",
	})

	batchSize := service.ParseBatchSize(c.DefaultQuery("batch_size", "500"))
	target := strings.TrimSpace(c.DefaultQuery("target", "resources")) // resources | games
	var (
		out service.MeiliReindexResult
		err error
	)
	if target == "games" {
		out, err = service.MeiliReindexGames(c.Request.Context(), batchSize)
	} else {
		out, err = service.MeiliReindexAll(c.Request.Context(), batchSize)
	}
	if err != nil {
		response.Error(c, 500, "重建索引失败: "+err.Error())
		return
	}
	response.OK(c, out)
}
