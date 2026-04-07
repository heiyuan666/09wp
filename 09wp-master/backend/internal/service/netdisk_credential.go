package service

import (
	"errors"
	"fmt"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"

	"gorm.io/gorm"
)

// LoadNetdiskCredentials 读取网盘凭证（单例）；不存在时从 system_configs 同步创建
func LoadNetdiskCredentials() (model.NetdiskCredential, error) {
	var n model.NetdiskCredential
	err := database.DB().First(&n, 1).Error
	if err == nil {
		return n, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.NetdiskCredential{}, err
	}
	var sys model.SystemConfig
	if e := database.DB().Order("id ASC").First(&sys).Error; e != nil {
		return model.NetdiskCredential{}, fmt.Errorf("系统配置不存在，无法初始化网盘凭证")
	}
	n = model.NetdiskCredential{
		ID:                   1,
		QuarkCookie:          sys.QuarkCookie,
		QuarkAutoSave:        sys.QuarkAutoSave,
		QuarkTargetFolderID:  sys.QuarkTargetFolderID,
		QuarkAdFilterEnabled: sys.QuarkAdFilterEnabled,
		QuarkBannedKeywords:  sys.QuarkBannedKeywords,
		Pan115Cookie:         sys.Pan115Cookie,
		Pan115AutoSave:       sys.Pan115AutoSave,
		Pan115TargetFolderID: sys.Pan115TargetFolderID,
		TianyiCookie:         sys.TianyiCookie,
		TianyiAutoSave:       sys.TianyiAutoSave,
		TianyiTargetFolderID: sys.TianyiTargetFolderID,
		Pan123Cookie:         sys.Pan123Cookie,
		Pan123AutoSave:       sys.Pan123AutoSave,
		Pan123TargetFolderID: sys.Pan123TargetFolderID,
		BaiduCookie:              sys.BaiduCookie,
		BaiduAutoSave:            sys.BaiduAutoSave,
		BaiduTargetPath:          sys.BaiduTargetPath,
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
	if err := database.DB().Create(&n).Error; err != nil {
		return model.NetdiskCredential{}, err
	}
	return n, nil
}

// SyncNetdiskCredentialsToSystemConfig 将凭证表同步到 system_configs，便于历史 SQL/备份一致
func SyncNetdiskCredentialsToSystemConfig(n model.NetdiskCredential) error {
	var sys model.SystemConfig
	if err := database.DB().Order("id ASC").First(&sys).Error; err != nil {
		return err
	}
	return database.DB().Model(&sys).Updates(map[string]any{
		"quark_cookie":            PrimaryQuarkCookie(n),
		"quark_auto_save":         n.QuarkAutoSave,
		"quark_target_folder_id":  n.QuarkTargetFolderID,
		"quark_ad_filter_enabled": n.QuarkAdFilterEnabled,
		"quark_banned_keywords":   n.QuarkBannedKeywords,
		"pan115_cookie":           PrimaryPan115Cookie(n),
		"pan115_auto_save":        n.Pan115AutoSave,
		"pan115_target_folder_id": n.Pan115TargetFolderID,
		"tianyi_cookie":           PrimaryTianyiCookie(n),
		"tianyi_auto_save":        n.TianyiAutoSave,
		"tianyi_target_folder_id": n.TianyiTargetFolderID,
		"pan123_cookie":           PrimaryPan123Cookie(n),
		"pan123_auto_save":        n.Pan123AutoSave,
		"pan123_target_folder_id": n.Pan123TargetFolderID,
		"baidu_cookie":                PrimaryBaiduCookie(n),
		"baidu_auto_save":             n.BaiduAutoSave,
		"baidu_target_path":           n.BaiduTargetPath,
		"xunlei_cookie":               PrimaryXunleiCookie(n),
		"xunlei_auto_save":            n.XunleiAutoSave,
		"xunlei_target_folder_id":     n.XunleiTargetFolderID,
		"uc_cookie":                   PrimaryUCCookie(n),
		"uc_auto_save":                n.UcAutoSave,
		"uc_target_folder_id":         n.UcTargetFolderID,
		"aliyun_refresh_token":        PrimaryAliyunRefreshToken(n),
		"aliyun_auto_save":            n.AliyunAutoSave,
		"aliyun_target_parent_file_id": n.AliyunTargetParentFileID,
		"replace_link_after_transfer": n.ReplaceLinkAfterTransfer,
	}).Error
}

// PersistXunleiRefreshToken 迅雷用 refresh_token 换 access 时若返回新 refresh_token，须写回库；
// 否则旧 token 已作废，下次请求会报 invalid_grant（4126）。
func PersistXunleiRefreshToken(oldRefresh, newRefresh string) error {
	oldTrim := strings.TrimSpace(oldRefresh)
	newTrim := strings.TrimSpace(newRefresh)
	if oldTrim == "" || newTrim == "" || oldTrim == newTrim {
		return nil
	}
	var n model.NetdiskCredential
	if err := database.DB().First(&n, 1).Error; err != nil {
		return err
	}
	changed := false
	if strings.TrimSpace(n.XunleiCookie) == oldTrim {
		n.XunleiCookie = newTrim
		changed = true
	}
	accs := append([]model.NetdiskCookieAccount(nil), n.XunleiCookieAccounts...)
	for i := range accs {
		if strings.TrimSpace(accs[i].Cookie) == oldTrim {
			accs[i].Cookie = newTrim
			changed = true
		}
	}
	if len(accs) > 0 {
		n.XunleiCookieAccounts = model.JSONCookieAccounts(accs)
	}
	if !changed {
		return nil
	}
	if err := database.DB().Save(&n).Error; err != nil {
		return err
	}
	return SyncNetdiskCredentialsToSystemConfig(n)
}
