package handler

import (
	"encoding/json"
	"io"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetNetdiskCredentials 获取网盘转存凭证（单例）
func GetNetdiskCredentials(c *gin.Context) {
	n, err := service.LoadNetdiskCredentials()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, n)
}

// UpdateNetdiskCredentials 更新网盘转存凭证，并同步到 system_configs 对应字段
func UpdateNetdiskCredentials(c *gin.Context) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Error(c, 400, "读入失败")
		return
	}
	var req model.NetdiskCredential
	if err := json.Unmarshal(raw, &req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	var patch map[string]json.RawMessage
	_ = json.Unmarshal(raw, &patch)
	uidVal, _ := c.Get("user_id")
	uid, _ := uidVal.(uint64)

	if strings.TrimSpace(req.BaiduTargetPath) == "" {
		req.BaiduTargetPath = "/"
	}
	if strings.TrimSpace(req.UcTargetFolderID) == "" {
		req.UcTargetFolderID = "0"
	}
	if strings.TrimSpace(req.AliyunTargetParentFileID) == "" {
		req.AliyunTargetParentFileID = "root"
	}
	if strings.TrimSpace(req.XunleiTargetFolderID) == "" {
		req.XunleiTargetFolderID = "0"
	}

	n, err := service.LoadNetdiskCredentials()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	updates := map[string]any{
		"quark_cookie":            req.QuarkCookie,
		"quark_auto_save":         req.QuarkAutoSave,
		"quark_target_folder_id":  req.QuarkTargetFolderID,
		"quark_ad_filter_enabled": req.QuarkAdFilterEnabled,
		"quark_banned_keywords":   req.QuarkBannedKeywords,
		"pan115_cookie":           req.Pan115Cookie,
		"pan115_auto_save":        req.Pan115AutoSave,
		"pan115_target_folder_id": req.Pan115TargetFolderID,
		"tianyi_cookie":           req.TianyiCookie,
		"tianyi_auto_save":        req.TianyiAutoSave,
		"tianyi_target_folder_id": req.TianyiTargetFolderID,
		"pan123_cookie":           req.Pan123Cookie,
		"pan123_auto_save":        req.Pan123AutoSave,
		"pan123_target_folder_id": req.Pan123TargetFolderID,
		"baidu_cookie":                req.BaiduCookie,
		"baidu_auto_save":             req.BaiduAutoSave,
		"baidu_target_path":           req.BaiduTargetPath,
		"xunlei_cookie":               req.XunleiCookie,
		"xunlei_auto_save":            req.XunleiAutoSave,
		"xunlei_target_folder_id":     req.XunleiTargetFolderID,
		"uc_cookie":                   req.UcCookie,
		"uc_auto_save":                req.UcAutoSave,
		"uc_target_folder_id":         req.UcTargetFolderID,
		"aliyun_refresh_token":        req.AliyunRefreshToken,
		"aliyun_auto_save":            req.AliyunAutoSave,
		"aliyun_target_parent_file_id": req.AliyunTargetParentFileID,
		"replace_link_after_transfer":  req.ReplaceLinkAfterTransfer,
		"updated_by":                  uid,
	}
	if patch != nil {
		if _, ok := patch["quark_cookie_accounts"]; ok {
			updates["quark_cookie_accounts"] = req.QuarkCookieAccounts
		}
		if _, ok := patch["uc_cookie_accounts"]; ok {
			updates["uc_cookie_accounts"] = req.UcCookieAccounts
		}
		if _, ok := patch["pan115_cookie_accounts"]; ok {
			updates["pan115_cookie_accounts"] = req.Pan115CookieAccounts
		}
		if _, ok := patch["tianyi_cookie_accounts"]; ok {
			updates["tianyi_cookie_accounts"] = req.TianyiCookieAccounts
		}
		if _, ok := patch["pan123_cookie_accounts"]; ok {
			updates["pan123_cookie_accounts"] = req.Pan123CookieAccounts
		}
		if _, ok := patch["baidu_cookie_accounts"]; ok {
			updates["baidu_cookie_accounts"] = req.BaiduCookieAccounts
		}
		if _, ok := patch["aliyun_refresh_token_accounts"]; ok {
			updates["aliyun_refresh_token_accounts"] = req.AliyunRefreshTokenAccounts
		}
		if _, ok := patch["xunlei_cookie_accounts"]; ok {
			updates["xunlei_cookie_accounts"] = req.XunleiCookieAccounts
		}
	}
	if err := database.DB().Model(&model.NetdiskCredential{}).Where("id = ?", n.ID).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	var fresh model.NetdiskCredential
	_ = database.DB().First(&fresh, n.ID).Error
	_ = service.SyncNetdiskCredentialsToSystemConfig(fresh)
	response.OK(c, nil)
}
