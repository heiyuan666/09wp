package handler

import (
	"context"
	"encoding/json"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
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
	HaokaOrderURL              string           `json:"haoka_order_url"`
	HaokaAgentRegURL           string           `json:"haoka_agent_reg_url"`
	FriendLinks                []FriendLinkItem `json:"friend_links"`
	DoubanHotNavEnabled        bool             `json:"douban_hot_nav_enabled"`
	HotSearchEnabled           bool             `json:"hot_search_enabled"`
	HomeRankBoardEnabled       bool             `json:"home_rank_board_enabled"`
	DoubanCoverProxyURL        string           `json:"douban_cover_proxy_url"`
}

// systemConfigOut 管理端返回：附带解析后的 friend_links
type systemConfigOut struct {
	model.SystemConfig
	FriendLinks []FriendLinkItem `json:"friend_links"`
}

// systemConfigPut 更新请求：friend_links 为数组，入库时序列化为 JSON 字符串
type systemConfigPut struct {
	model.SystemConfig
	FriendLinks []FriendLinkItem `json:"friend_links"`
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

// GetSystemConfig 获取全局系统配置
func GetSystemConfig(c *gin.Context) {
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		response.Error(c, 404, "系统配置不存在")
		return
	}
	links := parseFriendLinksJSON(cfg.FriendLinks)
	response.OK(c, systemConfigOut{SystemConfig: cfg, FriendLinks: links})
}

// GetPublicSystemConfig 获取前台可用系统配置（无需登录）
func GetPublicSystemConfig(c *gin.Context) {
	cacheKey := "public:system-config:v2"
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
		SubmissionAutoTransfer:     cfg.SubmissionAutoTransfer,
		ResourceDetailAutoTransfer: cfg.ResourceDetailAutoTransfer,
		HaokaOrderURL:              cfg.HaokaOrderURL,
		HaokaAgentRegURL:           cfg.HaokaAgentRegURL,
		FriendLinks:                links,
		DoubanHotNavEnabled:        cfg.DoubanHotNavEnabled,
		HotSearchEnabled:           cfg.HotSearchEnabled,
		HomeRankBoardEnabled:       cfg.HomeRankBoardEnabled,
		DoubanCoverProxyURL:        cfg.DoubanCoverProxyURL,
	})

	// 写入缓存（短 TTL）
	if raw, err := json.Marshal(publicSystemConfig{
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
		SubmissionAutoTransfer:     cfg.SubmissionAutoTransfer,
		ResourceDetailAutoTransfer: cfg.ResourceDetailAutoTransfer,
		HaokaOrderURL:              cfg.HaokaOrderURL,
		HaokaAgentRegURL:           cfg.HaokaAgentRegURL,
		FriendLinks:                links,
		DoubanHotNavEnabled:        cfg.DoubanHotNavEnabled,
		HotSearchEnabled:           cfg.HotSearchEnabled,
		HomeRankBoardEnabled:       cfg.HomeRankBoardEnabled,
		DoubanCoverProxyURL:        cfg.DoubanCoverProxyURL,
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

	updates := map[string]interface{}{
		"site_title":                    siteTitle,
		"admin_email":                   req.AdminEmail,
		"support_email":                 req.SupportEmail,
		"contact_phone":                 req.ContactPhone,
		"contact_qq":                    req.ContactQQ,
		"logo_url":                      req.LogoURL,
		"favicon_url":                   req.FaviconURL,
		"seo_keywords":                  req.SeoKeywords,
		"seo_description":               req.SeoDescription,
		"icp_record":                    req.IcpRecord,
		"footer_text":                   req.FooterText,
		"clarity_project_id":            strings.TrimSpace(req.ClarityProjectID),
		"clarity_enabled":               req.ClarityEnabled,
		"friend_links":                  friendLinksStr,
		"allow_register":                req.AllowRegister,
		"submission_need_review":        req.SubmissionNeedReview,
		"submission_auto_transfer":      req.SubmissionAutoTransfer,
		"resource_detail_auto_transfer": req.ResourceDetailAutoTransfer,
		"haoka_user_id":                 req.HaokaUserID,
		"haoka_secret":                  req.HaokaSecret,
		"haoka_sync_enabled":            req.HaokaSyncEnabled,
		"haoka_sync_interval":           req.HaokaSyncInterval,
		"haoka_order_url":               req.HaokaOrderURL,
		"haoka_agent_reg_url":           req.HaokaAgentRegURL,
		"smtp_host":                     req.SmtpHost,
		"smtp_port":                     req.SmtpPort,
		"smtp_user":                     req.SmtpUser,
		"smtp_pass":                     req.SmtpPass,
		"smtp_from":                     req.SmtpFrom,
		"tg_bot_token":                  req.TgBotToken,
		"tg_proxy_url":                  req.TgProxyURL,
		"tg_api_id":                     req.TgAPIID,
		"tg_api_hash":                   req.TgAPIHash,
		"tg_session":                    req.TgSession,
		"pancheck_base_url":             req.PanCheckBaseURL,
		"link_check_enabled":            req.LinkCheckEnabled,
		"link_check_interval":           req.LinkCheckInterval,
		"douban_hot_nav_enabled":        req.DoubanHotNavEnabled,
		"hot_search_enabled":            req.HotSearchEnabled,
		"home_rank_board_enabled":       req.HomeRankBoardEnabled,
		"douban_cover_proxy_url":        req.DoubanCoverProxyURL,
		"auto_delete_invalid_links":     req.AutoDeleteInvalidLinks,
		"hide_invalid_links_in_search":  req.HideInvalidLinksInSearch,
		"tg_channel_chat_id":            req.TgChannelChatID,
		"tg_sync_enabled":               req.TgSyncEnabled,
		"tg_sync_interval":              req.TgSyncInterval,
		"tg_default_cat_id":             req.TgDefaultCatID,
		"updated_by":                    uid,
	}

	if err := database.DB().Model(&cfg).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}

	// 配置更新后失效前台缓存，避免前台获取旧配置
	service.DeleteSearchCache(context.Background(), "public:system-config:v2")
	response.OK(c, nil)
}
