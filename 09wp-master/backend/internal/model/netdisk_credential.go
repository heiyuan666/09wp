package model

import "time"

// NetdiskCredential 网盘转存凭证（单例 id=1），与系统配置分离便于独立维护 Cookie
type NetdiskCredential struct {
	ID     uint64 `gorm:"primaryKey" json:"id"`
	// 夸克
	QuarkCookie         string `gorm:"type:text" json:"quark_cookie"`
	QuarkCookieAccounts JSONCookieAccounts `gorm:"type:text;column:quark_cookie_accounts" json:"quark_cookie_accounts"`
	QuarkAutoSave       bool   `gorm:"default:false" json:"quark_auto_save"`
	QuarkTargetFolderID string `gorm:"size:64;default:'0'" json:"quark_target_folder_id"`
	QuarkAdFilterEnabled bool   `gorm:"default:false" json:"quark_ad_filter_enabled"`
	QuarkBannedKeywords  string `gorm:"type:text" json:"quark_banned_keywords"` // 逗号分隔
	// UC 网盘（与夸克同源接口）
	UcCookie         string `gorm:"type:text" json:"uc_cookie"`
	UcCookieAccounts JSONCookieAccounts `gorm:"type:text;column:uc_cookie_accounts" json:"uc_cookie_accounts"`
	UcAutoSave       bool   `gorm:"default:false" json:"uc_auto_save"`
	UcTargetFolderID string `gorm:"size:64;default:'0'" json:"uc_target_folder_id"`
	// 阿里云盘（开放平台 refresh_token）
	AliyunRefreshToken        string             `gorm:"type:text" json:"aliyun_refresh_token"`
	AliyunRefreshTokenAccounts JSONCookieAccounts `gorm:"type:text;column:aliyun_refresh_token_accounts" json:"aliyun_refresh_token_accounts"`
	// AliyunRenewAPIURL OpenList 等第三方提供的 token 续期接口地址（可选）。
	// 例如：https://api.oplist.org/alicloud/renewapi
	AliyunRenewAPIURL string `gorm:"size:500;default:''" json:"aliyun_renew_api_url"`
	AliyunAutoSave           bool   `gorm:"default:false" json:"aliyun_auto_save"`
	AliyunTargetParentFileID string `gorm:"size:64;default:'root'" json:"aliyun_target_parent_file_id"`
	// 115
	Pan115Cookie         string `gorm:"type:text" json:"pan115_cookie"`
	Pan115CookieAccounts JSONCookieAccounts `gorm:"type:text;column:pan115_cookie_accounts" json:"pan115_cookie_accounts"`
	Pan115AutoSave       bool   `gorm:"default:false" json:"pan115_auto_save"`
	Pan115TargetFolderID string `gorm:"size:64;default:''" json:"pan115_target_folder_id"`
	// 天翼
	TianyiCookie         string `gorm:"type:text" json:"tianyi_cookie"`
	TianyiCookieAccounts JSONCookieAccounts `gorm:"type:text;column:tianyi_cookie_accounts" json:"tianyi_cookie_accounts"`
	TianyiAutoSave       bool   `gorm:"default:false" json:"tianyi_auto_save"`
	TianyiTargetFolderID string `gorm:"size:32;default:'-11'" json:"tianyi_target_folder_id"`
	// 123
	Pan123Cookie         string `gorm:"type:text" json:"pan123_cookie"`
	Pan123CookieAccounts JSONCookieAccounts `gorm:"type:text;column:pan123_cookie_accounts" json:"pan123_cookie_accounts"`
	Pan123AutoSave       bool   `gorm:"default:false" json:"pan123_auto_save"`
	Pan123TargetFolderID string `gorm:"size:64;default:'0'" json:"pan123_target_folder_id"`
	// 百度
	BaiduCookie          string             `gorm:"type:text" json:"baidu_cookie"`
	BaiduCookieAccounts  JSONCookieAccounts `gorm:"type:text;column:baidu_cookie_accounts" json:"baidu_cookie_accounts"`
	BaiduAutoSave        bool               `gorm:"default:false" json:"baidu_auto_save"`
	BaiduTargetPath string `gorm:"size:500;default:'/'" json:"baidu_target_path"` // 转存到网盘中的路径，如 / 或 /我的资源
	// 迅雷（xunlei_cookie 为 refresh_token）
	XunleiCookie         string             `gorm:"type:text" json:"xunlei_cookie"`
	XunleiCookieAccounts JSONCookieAccounts `gorm:"type:text;column:xunlei_cookie_accounts" json:"xunlei_cookie_accounts"`
	XunleiAutoSave       bool               `gorm:"default:false" json:"xunlei_auto_save"`
	XunleiTargetFolderID string `gorm:"size:64;default:'0'" json:"xunlei_target_folder_id"`
	// 转存成功后是否将资源管理中的链接替换为您本人网盘生成的分享链接（各盘支持情况见说明）
	ReplaceLinkAfterTransfer bool `gorm:"default:false" json:"replace_link_after_transfer"`
	UpdatedBy                uint64 `gorm:"default:0" json:"updated_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (NetdiskCredential) TableName() string {
	return "netdisk_credentials"
}
