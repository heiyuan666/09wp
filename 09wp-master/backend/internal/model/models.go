package model

import "time"

// Admin 管理员表
type Admin struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	Email        string    `gorm:"size:100;default:''" json:"email"`
	Status       int8      `gorm:"default:1" json:"status"` // 1=正常 0=禁用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// User 普通用户表
type User struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"size:50;uniqueIndex" json:"username"`
	Email        string     `gorm:"size:100;uniqueIndex" json:"email"`
	Phone        *string    `gorm:"size:20" json:"phone,omitempty"`
	RoleID       *uint64    `gorm:"index" json:"role_id,omitempty"`
	Name         string     `gorm:"size:50" json:"name"`  // 显示名称
	Bio          string     `gorm:"size:255" json:"bio"`  // 个人简介
	Tags         string     `gorm:"size:255" json:"tags"` // 个人标签
	PasswordHash string     `gorm:"size:255" json:"-"`
	Avatar       *string    `gorm:"size:255" json:"avatar,omitempty"`
	Status       int8       `gorm:"default:1" json:"status"` // 1=正常 0=禁用
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Category 分类表
type Category struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Slug      string    `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	Status    int8      `gorm:"default:1" json:"status"` // 1=显示 0=隐藏
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Resource 网盘资源表
type Resource struct {
	ID                 uint64     `gorm:"primaryKey" json:"id"`
	Title              string     `gorm:"size:200;not null" json:"title"`
	Link               string     `gorm:"size:500;not null" json:"link"`
	ExtraLinks         JSONStringList `gorm:"type:text;column:extra_links" json:"extra_links"` // 其它网盘分享链接（JSON 数组），主链接见 link
	CategoryID         uint64     `gorm:"index;not null" json:"category_id"`
	Source             string     `gorm:"size:30;default:'';index" json:"source"`             // 来源：manual/telegram
	ExternalID         string     `gorm:"size:120;default:'';index" json:"external_id"` // 外部 ID（TG/RSS 等）去重在业务层校验；库内勿对 '' 做唯一索引
	Description        string     `gorm:"type:text" json:"description"`
	ExtractCode        string     `gorm:"size:50" json:"extract_code"`
	Cover              string     `gorm:"size:2048" json:"cover"` // 外链封面（如 telesco.pe）可能极长
	Tags               string     `gorm:"size:255" json:"tags"`                 // 逗号分隔
	LinkValid          bool       `gorm:"default:true;index" json:"link_valid"` // 链接检测是否有效
	LinkCheckMsg       string     `gorm:"size:255;default:''" json:"link_check_msg"`
	LinkCheckedAt      *time.Time `json:"link_checked_at,omitempty"`
	TransferStatus     string     `gorm:"size:20;default:'';index" json:"transfer_status"` // pending/success/failed
	TransferMsg        string     `gorm:"size:255;default:''" json:"transfer_msg"`
	TransferRetryCount int        `gorm:"default:0" json:"transfer_retry_count"`
	TransferLastAt     *time.Time `json:"transfer_last_at,omitempty"`
	// idx_res_pub_hot: 前台 WHERE status=1 ORDER BY view_count（InnoDB 二级索引叶子含主键 id）
	ViewCount uint64 `gorm:"default:0;index:idx_res_pub_hot,priority:2" json:"view_count"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	// idx_res_pub_latest: 前台 WHERE status=1 ORDER BY created_at
	Status    int8      `gorm:"default:1;index;index:idx_res_pub_hot,priority:1;index:idx_res_pub_latest,priority:1" json:"status"` // 1=显示 0=隐藏
	CreatedAt time.Time `gorm:"index:idx_res_pub_latest,priority:2" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResourceTransferLog 网盘转存尝试日志
type ResourceTransferLog struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	ResourceID uint64 `gorm:"index;not null" json:"resource_id"`
	Attempt    int    `gorm:"default:1" json:"attempt"` // 第几次尝试（重试从 1 开始）

	Platform    string `gorm:"size:32;default:'';index" json:"platform"`
	Status      string `gorm:"size:20;default:'';index" json:"status"` // pending/success/failed
	Message     string `gorm:"size:255;default:''" json:"message"`
	ErrorDetail string `gorm:"type:text" json:"error_detail"`

	OldLink string `gorm:"size:500;default:''" json:"old_link"`
	NewLink string `gorm:"size:500;default:''" json:"new_link"`

	// OwnShareURL 是“由网盘生成的本人分享链接”（用于替换）
	OwnShareURL string `gorm:"size:500;default:''" json:"own_share_url"`

	// FilterLog 记录转存后的过滤日志（如广告关键词过滤），JSON 字符串
	FilterLog string `gorm:"type:text" json:"filter_log,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserFavorite 用户收藏表
type UserFavorite struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `gorm:"index;not null" json:"user_id"`
	ResourceID uint64    `gorm:"index;not null" json:"resource_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// Menu 系统菜单（对齐前端 IMenuItem）
type Menu struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	Type       string    `gorm:"size:20;not null;index" json:"type"` // directory/menu/button
	Path       string    `gorm:"size:255" json:"path"`
	Title      string    `gorm:"size:100;not null" json:"title"`
	Icon       string    `gorm:"size:100" json:"icon"`
	ParentID   *uint64   `gorm:"index" json:"parent_id"`
	Order      int       `gorm:"default:0" json:"order"`
	Status     int8      `gorm:"default:1;index" json:"status"` // 1=active 0=inactive
	Permission string    `gorm:"size:100" json:"permission"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Role 系统角色（对齐前端 IRoleItem）
type Role struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Code        string    `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Description string    `gorm:"size:255" json:"description"`
	IsBuiltIn   bool      `gorm:"default:false" json:"is_built_in"`
	Status      int8      `gorm:"default:1;index" json:"status"` // 1=active 0=inactive
	MenuIDs     string    `gorm:"type:text" json:"menu_ids"`     // JSON string: ["1","2"]
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SystemConfig 全局系统配置（单例）
type SystemConfig struct {
	ID               uint64 `gorm:"primaryKey" json:"id"`
	SiteTitle        string `gorm:"size:120" json:"site_title"`
	AdminEmail       string `gorm:"size:120" json:"admin_email"`
	SupportEmail     string `gorm:"size:120" json:"support_email"`
	ContactPhone     string `gorm:"size:30" json:"contact_phone"`
	ContactQQ        string `gorm:"size:30" json:"contact_qq"`
	LogoURL          string `gorm:"size:255" json:"logo_url"`
	FaviconURL       string `gorm:"size:255" json:"favicon_url"`
	SeoKeywords      string `gorm:"size:255" json:"seo_keywords"`
	SeoDescription   string `gorm:"size:500" json:"seo_description"`
	IcpRecord        string `gorm:"size:120" json:"icp_record"`
	FooterText       string `gorm:"size:255" json:"footer_text"`
	ClarityProjectID string `gorm:"size:64;default:''" json:"clarity_project_id"`
	ClarityEnabled   bool   `gorm:"default:false" json:"clarity_enabled"`
	// FriendLinks JSON 数组：[{"title":"名称","url":"https://..."}]，API 层用 json:"-" 由 handler 解析/序列化
	FriendLinks                string `gorm:"type:text" json:"-"`
	AllowRegister              bool   `gorm:"default:true" json:"allow_register"`
	SubmissionNeedReview       bool   `gorm:"default:true" json:"submission_need_review"`
	SubmissionAutoTransfer     bool   `gorm:"default:false" json:"submission_auto_transfer"`
	ResourceDetailAutoTransfer bool   `gorm:"default:false" json:"resource_detail_auto_transfer"`
	HaokaUserID                string `gorm:"size:120;default:''" json:"haoka_user_id"`
	HaokaSecret                string `gorm:"size:255;default:''" json:"haoka_secret"`
	HaokaSyncEnabled           bool   `gorm:"default:false" json:"haoka_sync_enabled"`
	HaokaSyncInterval          int    `gorm:"default:3600" json:"haoka_sync_interval"` // 秒
	// Haoka 店铺下单/代理注册链接（前台可配置按钮跳转）
	HaokaOrderURL            string `gorm:"size:500;default:''" json:"haoka_order_url"`
	HaokaAgentRegURL         string `gorm:"size:500;default:''" json:"haoka_agent_reg_url"`
	SmtpHost                 string `gorm:"size:120" json:"smtp_host"`
	SmtpPort                 int    `gorm:"default:25" json:"smtp_port"`
	SmtpUser                 string `gorm:"size:120" json:"smtp_user"`
	SmtpPass                 string `gorm:"size:120" json:"smtp_pass"`
	SmtpFrom                 string `gorm:"size:120" json:"smtp_from"`
	TgBotToken               string `gorm:"size:255" json:"tg_bot_token"`
	TgProxyURL               string `gorm:"size:255;default:''" json:"tg_proxy_url"` // TG全局默认代理
	TgAPIID                  int    `gorm:"default:0" json:"tg_api_id"`
	TgAPIHash                string `gorm:"size:120;default:''" json:"tg_api_hash"`
	TgSession                string `gorm:"type:text" json:"tg_session"` // MTProto 用户会话（base64）
	PanCheckBaseURL          string `gorm:"column:pancheck_base_url;size:255;default:''" json:"pancheck_base_url"`
	LinkCheckEnabled         bool   `gorm:"default:false" json:"link_check_enabled"`
	LinkCheckInterval        int    `gorm:"default:3600" json:"link_check_interval"` // 秒
	TgChannelChatID          string `gorm:"size:120" json:"tg_channel_chat_id"`      // 如 -1001234567890
	TgSyncEnabled            bool   `gorm:"default:false" json:"tg_sync_enabled"`
	TgSyncInterval           int    `gorm:"default:300" json:"tg_sync_interval"` // 秒
	TgDefaultCatID           uint64 `gorm:"default:0" json:"tg_default_cat_id"`
	TgLastUpdateID           int64  `gorm:"default:0" json:"tg_last_update_id"`
	QuarkCookie              string `gorm:"type:text" json:"quark_cookie"`
	QuarkAutoSave            bool   `gorm:"default:false" json:"quark_auto_save"`
	QuarkTargetFolderID      string `gorm:"size:64;default:'0'" json:"quark_target_folder_id"` // 转存目标目录fid，默认根目录
	QuarkAdFilterEnabled     bool   `gorm:"default:false" json:"quark_ad_filter_enabled"`
	QuarkBannedKeywords      string `gorm:"type:text" json:"quark_banned_keywords"`
	Pan115Cookie             string `gorm:"type:text" json:"pan115_cookie"`
	Pan115AutoSave           bool   `gorm:"default:false" json:"pan115_auto_save"`
	Pan115TargetFolderID     string `gorm:"size:64;default:''" json:"pan115_target_folder_id"` // 115 转存目标目录 cid，空为根目录
	TianyiCookie             string `gorm:"type:text" json:"tianyi_cookie"`
	TianyiAutoSave           bool   `gorm:"default:false" json:"tianyi_auto_save"`
	TianyiTargetFolderID     string `gorm:"size:32;default:'-11'" json:"tianyi_target_folder_id"` // 天翼个人云目标目录，默认 -11
	Pan123Cookie             string `gorm:"type:text" json:"pan123_cookie"`                       // Bearer Token 或含 authorization 的文本
	Pan123AutoSave           bool   `gorm:"default:false" json:"pan123_auto_save"`
	Pan123TargetFolderID     string `gorm:"size:64;default:'0'" json:"pan123_target_folder_id"` // 123 转存目标 parentFileId
	BaiduCookie              string `gorm:"type:text" json:"baidu_cookie"`
	BaiduAutoSave            bool   `gorm:"default:false" json:"baidu_auto_save"`
	BaiduTargetPath          string `gorm:"size:500;default:'/'" json:"baidu_target_path"`
	XunleiCookie             string `gorm:"type:text" json:"xunlei_cookie"` // 迅雷 refresh_token
	XunleiAutoSave           bool   `gorm:"default:false" json:"xunlei_auto_save"`
	XunleiTargetFolderID     string `gorm:"size:64;default:'0'" json:"xunlei_target_folder_id"`
	UcCookie                 string `gorm:"type:text" json:"uc_cookie"`
	UcAutoSave               bool   `gorm:"default:false" json:"uc_auto_save"`
	UcTargetFolderID         string `gorm:"size:64;default:'0'" json:"uc_target_folder_id"`
	AliyunRefreshToken       string `gorm:"type:text" json:"aliyun_refresh_token"`
	AliyunAutoSave           bool   `gorm:"default:false" json:"aliyun_auto_save"`
	AliyunTargetParentFileID string `gorm:"size:64;default:'root'" json:"aliyun_target_parent_file_id"`
	ReplaceLinkAfterTransfer bool   `gorm:"default:false" json:"replace_link_after_transfer"`
	// DoubanHotNavEnabled 是否在前台导航栏展示豆瓣热门
	DoubanHotNavEnabled bool `gorm:"default:false" json:"douban_hot_nav_enabled"`
	HotSearchEnabled    bool `gorm:"default:true" json:"hot_search_enabled"`
	// HomeRankBoardEnabled 前台首页是否展示排行榜（热门/最新/豆瓣）
	HomeRankBoardEnabled bool `gorm:"default:true" json:"home_rank_board_enabled"`
	// DoubanCoverProxyURL 豆瓣封面返代接口模板，如：
	// https://image.baidu.com/search/down?url=
	// 或支持模板：...{url}
	DoubanCoverProxyURL string `gorm:"size:500;default:''" json:"douban_cover_proxy_url"`
	// TgImageProxyURL TG 同步等资源的外链封面返代模板，如：https://wsrv.nl/?url=
	// 仅对 source=telegram 且封面为 http(s) 外链时由前端拼接；本地 /public/covers 不经过此代理。
	TgImageProxyURL string `gorm:"size:500;default:''" json:"tg_image_proxy_url"`

	// AutoDeleteInvalidLinks 是否对失效链接资源自动“删除”
	// 物理删除 resources，并清理对应的 user_favorites 记录（尽力兜底）。
	AutoDeleteInvalidLinks bool `gorm:"default:false" json:"auto_delete_invalid_links"`

	// HideInvalidLinksInSearch 是否在前台搜索中隐藏失效链接资源
	// 若开启且请求未显式传 link_valid 参数，则强制 link_valid = true
	HideInvalidLinksInSearch bool      `gorm:"default:false" json:"hide_invalid_links_in_search"`
	UpdatedBy                uint64    `gorm:"default:0" json:"updated_by"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// NavigationMenu 前台导航菜单（用于顶部导航 / 首页推荐按钮）
type NavigationMenu struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:100;not null" json:"title"`
	Path      string    `gorm:"size:255;not null;default:''" json:"path"`
	Position  string    `gorm:"size:32;not null;index" json:"position"` // top_nav / home_promo
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	Visible   bool      `gorm:"default:true" json:"visible"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResourceFeedback 用户在资源详情页提交的反馈
type ResourceFeedback struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	ResourceID uint64    `gorm:"index;not null" json:"resource_id"`
	Type       string    `gorm:"size:32;default:''" json:"type"` // link_invalid / password_error / content_error / other
	Content    string    `gorm:"type:text;not null" json:"content"`
	Contact    string    `gorm:"size:255;default:''" json:"contact"`
	Status     string    `gorm:"size:20;default:'pending';index" json:"status"` // pending / processed
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SearchHotWord 用户搜索关键词统计（用于热搜榜）
type SearchHotWord struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	Keyword        string    `gorm:"size:200;uniqueIndex;not null" json:"keyword"`
	SearchCount    uint64    `gorm:"default:0;index" json:"search_count"`
	LastSearchedAt time.Time `json:"last_searched_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// KeywordBlock 关键词屏蔽（用于前台搜索/热搜过滤）
type KeywordBlock struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Keyword   string    `gorm:"size:200;uniqueIndex;not null" json:"keyword"`
	Enabled   bool      `gorm:"default:true;index" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TelegramAuthState MTProto 登录过程临时状态（单例）
type TelegramAuthState struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Phone         string    `gorm:"size:30;default:''" json:"phone"`
	PhoneCodeHash string    `gorm:"size:255;default:''" json:"phone_code_hash"`
	TempSession   string    `gorm:"type:text" json:"temp_session"` // base64
	NeedPassword  bool      `gorm:"default:false" json:"need_password"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TelegramChannel TG 频道采集配置（独立模块）
type TelegramChannel struct {
	ID             uint64     `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"size:120;not null" json:"name"`
	BotToken       string     `gorm:"size:255;not null" json:"bot_token"`
	ChannelChatID  string     `gorm:"size:120;not null" json:"channel_chat_id"`
	ProxyURL       string     `gorm:"size:255;default:''" json:"proxy_url"` // 支持 http:// 或 socks5://
	DefaultCatID   uint64     `gorm:"default:0" json:"default_cat_id"`
	Enabled        bool       `gorm:"default:true;index" json:"enabled"`
	SyncInterval   int        `gorm:"default:300" json:"sync_interval"` // 秒
	LastUpdateID   int64      `gorm:"default:0" json:"last_update_id"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	LastSyncStatus string     `gorm:"size:30;default:''" json:"last_sync_status"` // success/failed
	LastSyncMsg    string     `gorm:"size:255;default:''" json:"last_sync_msg"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// RSSSubscription RSS 订阅抓取配置
type RSSSubscription struct {
	ID             uint64     `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"size:120;not null" json:"name"`
	FeedURL        string     `gorm:"size:500;not null;uniqueIndex" json:"feed_url"`
	DefaultCatID   uint64     `gorm:"default:0" json:"default_cat_id"`
	Enabled        bool       `gorm:"default:true;index" json:"enabled"`
	SyncInterval   int        `gorm:"default:1800" json:"sync_interval"` // 秒
	MaxItems       int        `gorm:"default:50" json:"max_items"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	LastSyncStatus string     `gorm:"size:30;default:''" json:"last_sync_status"` // success/failed
	LastSyncMsg    string     `gorm:"size:255;default:''" json:"last_sync_msg"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
