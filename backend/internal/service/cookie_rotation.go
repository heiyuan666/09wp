package service

import (
	"strings"
	"sync"

	"dfan-netdisk-backend/internal/model"
)

// PickedCookie 轮询选中的账号（单次转存全程应固定使用同一 cookie）
type PickedCookie struct {
	Cookie string
	Name   string
	// 非空时覆盖对应网盘全局转存目录 ID
	TargetFolderID string
	// 非空时覆盖百度全局转存路径
	TargetPath string
}

var cookieRotateMu sync.Mutex
var cookieRotateNext = map[string]int{}

func pickRotatingCookie(platform string, accounts model.JSONCookieAccounts, legacy string) PickedCookie {
	active := accounts.ActiveCookies()
	if len(active) == 0 {
		c := strings.TrimSpace(legacy)
		if c == "" {
			return PickedCookie{}
		}
		return PickedCookie{Cookie: c, Name: ""}
	}
	cookieRotateMu.Lock()
	i := cookieRotateNext[platform] % len(active)
	cookieRotateNext[platform]++
	cookieRotateMu.Unlock()
	a := active[i]
	name := strings.TrimSpace(a.Name)
	if name == "" {
		name = "账号"
	}
	return PickedCookie{
		Cookie:         strings.TrimSpace(a.Cookie),
		Name:           name,
		TargetFolderID: strings.TrimSpace(a.TargetFolderID),
		TargetPath:     strings.TrimSpace(a.TargetPath),
	}
}

// PickQuarkCookie 夸克：多账号顺序轮询；无多账号时回落到 quark_cookie
func PickQuarkCookie(cfg model.NetdiskCredential) PickedCookie {
	return pickRotatingCookie("quark", cfg.QuarkCookieAccounts, cfg.QuarkCookie)
}

// PickUCCookie UC 网盘
func PickUCCookie(cfg model.NetdiskCredential) PickedCookie {
	return pickRotatingCookie("uc", cfg.UcCookieAccounts, cfg.UcCookie)
}

// PickPan115Cookie 115
func PickPan115Cookie(cfg model.NetdiskCredential) PickedCookie {
	return pickRotatingCookie("pan115", cfg.Pan115CookieAccounts, cfg.Pan115Cookie)
}

// PickTianyiCookie 天翼
func PickTianyiCookie(cfg model.NetdiskCredential) PickedCookie {
	return pickRotatingCookie("tianyi", cfg.TianyiCookieAccounts, cfg.TianyiCookie)
}

// PickPan123Cookie 123 云盘（Bearer / 原文）
func PickPan123Cookie(cfg model.NetdiskCredential) PickedCookie {
	return pickRotatingCookie("pan123", cfg.Pan123CookieAccounts, cfg.Pan123Cookie)
}

// PickBaiduCookie 百度
func PickBaiduCookie(cfg model.NetdiskCredential) PickedCookie {
	return pickRotatingCookie("baidu", cfg.BaiduCookieAccounts, cfg.BaiduCookie)
}

// PickAliyunRefreshToken 阿里云盘 refresh_token 轮询（存于账号项的 cookie 字段）
func PickAliyunRefreshToken(cfg model.NetdiskCredential) PickedCookie {
	return pickRotatingCookie("aliyun_refresh", cfg.AliyunRefreshTokenAccounts, cfg.AliyunRefreshToken)
}

// PickXunleiRefreshToken 迅雷 refresh_token 轮询
func PickXunleiRefreshToken(cfg model.NetdiskCredential) PickedCookie {
	return pickRotatingCookie("xunlei_refresh", cfg.XunleiCookieAccounts, cfg.XunleiCookie)
}

// PrimaryQuarkCookie 同步到 system_configs 等：优先多账号列表首项，否则主 cookie
func PrimaryQuarkCookie(n model.NetdiskCredential) string {
	if p := n.QuarkCookieAccounts.PrimaryCookie(); p != "" {
		return p
	}
	return strings.TrimSpace(n.QuarkCookie)
}

func PrimaryUCCookie(n model.NetdiskCredential) string {
	if p := n.UcCookieAccounts.PrimaryCookie(); p != "" {
		return p
	}
	return strings.TrimSpace(n.UcCookie)
}

func PrimaryPan115Cookie(n model.NetdiskCredential) string {
	if p := n.Pan115CookieAccounts.PrimaryCookie(); p != "" {
		return p
	}
	return strings.TrimSpace(n.Pan115Cookie)
}

func PrimaryTianyiCookie(n model.NetdiskCredential) string {
	if p := n.TianyiCookieAccounts.PrimaryCookie(); p != "" {
		return p
	}
	return strings.TrimSpace(n.TianyiCookie)
}

func PrimaryPan123Cookie(n model.NetdiskCredential) string {
	if p := n.Pan123CookieAccounts.PrimaryCookie(); p != "" {
		return p
	}
	return strings.TrimSpace(n.Pan123Cookie)
}

func PrimaryBaiduCookie(n model.NetdiskCredential) string {
	if p := n.BaiduCookieAccounts.PrimaryCookie(); p != "" {
		return p
	}
	return strings.TrimSpace(n.BaiduCookie)
}

func PrimaryAliyunRefreshToken(n model.NetdiskCredential) string {
	if p := n.AliyunRefreshTokenAccounts.PrimaryCookie(); p != "" {
		return p
	}
	return strings.TrimSpace(n.AliyunRefreshToken)
}

func PrimaryXunleiCookie(n model.NetdiskCredential) string {
	if p := n.XunleiCookieAccounts.PrimaryCookie(); p != "" {
		return p
	}
	return strings.TrimSpace(n.XunleiCookie)
}

// 以下：多账号项里填了 target_folder_id / target_path 时优先，否则用全局配置与各盘默认

func effectiveQuarkUCFolderID(picked PickedCookie, global string) string {
	if s := strings.TrimSpace(picked.TargetFolderID); s != "" {
		return s
	}
	s := strings.TrimSpace(global)
	if s == "" {
		return "0"
	}
	return s
}

func effectivePan115Cid(picked PickedCookie, global string) string {
	if s := strings.TrimSpace(picked.TargetFolderID); s != "" {
		return s
	}
	return strings.TrimSpace(global)
}

func effectiveTianyiFolder(picked PickedCookie, global string) string {
	if s := strings.TrimSpace(picked.TargetFolderID); s != "" {
		return s
	}
	s := strings.TrimSpace(global)
	if s == "" {
		return "-11"
	}
	return s
}

func effectivePan123Parent(picked PickedCookie, global string) string {
	if s := strings.TrimSpace(picked.TargetFolderID); s != "" {
		return s
	}
	s := strings.TrimSpace(global)
	if s == "" {
		return "0"
	}
	return s
}

func effectiveAliyunParent(picked PickedCookie, global string) string {
	if s := strings.TrimSpace(picked.TargetFolderID); s != "" {
		return s
	}
	s := strings.TrimSpace(global)
	if s == "" {
		return "root"
	}
	return s
}

func effectiveXunleiFolder(picked PickedCookie, global string) string {
	if s := strings.TrimSpace(picked.TargetFolderID); s != "" {
		return s
	}
	s := strings.TrimSpace(global)
	if s == "" {
		return "0"
	}
	return s
}

func effectiveBaiduPath(picked PickedCookie, global string) string {
	if s := strings.TrimSpace(picked.TargetPath); s != "" {
		return s
	}
	s := strings.TrimSpace(global)
	if s == "" {
		return "/"
	}
	return s
}
