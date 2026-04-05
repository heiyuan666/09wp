package service

import (
	"regexp"
	"strings"
)

// TransferPlatform 资源链接对应的转存通道
type TransferPlatform int

const (
	PlatformUnknown TransferPlatform = iota
	PlatformBaidu
	PlatformQuark
	PlatformUC
	PlatformPan115
	PlatformTianyi
	PlatformPan123
	PlatformAliyun
	PlatformXunlei
)

var (
	reBaidu   = regexp.MustCompile(`(?i)pan\.baidu\.com/(?:s/|share/)`)
	reQuark   = regexp.MustCompile(`(?i)pan\.quark\.cn/s/`)
	reUC      = regexp.MustCompile(`(?i)(?:drive|yun)\.uc\.cn/s/`)
	re115     = regexp.MustCompile(`(?i)(?:115\.com|115cdn\.com)/s/`)
	reTianyi  = regexp.MustCompile(`(?i)(?:cloud\.189\.cn|h5\.cloud\.189\.cn)`)
	re123Pan  = regexp.MustCompile(`(?i)(?:www\.)?(?:123pan|123684|123685|123912|123592|123865)\.(?:com|cn)/s/`)
	reAliyun  = regexp.MustCompile(`(?i)(?:www\.)?(?:aliyundrive\.com|alipan\.com)/s/`)
	reXunlei  = regexp.MustCompile(`(?i)pan\.xunlei\.com/s/`)
)

// DetectTransferPlatform 根据链接识别网盘类型
func DetectTransferPlatform(link string) TransferPlatform {
	u := strings.TrimSpace(link)
	if u == "" {
		return PlatformUnknown
	}
	if reBaidu.MatchString(u) {
		return PlatformBaidu
	}
	if reQuark.MatchString(u) {
		return PlatformQuark
	}
	if reUC.MatchString(u) {
		return PlatformUC
	}
	if re115.MatchString(u) {
		return PlatformPan115
	}
	if reTianyi.MatchString(u) {
		return PlatformTianyi
	}
	if re123Pan.MatchString(u) {
		return PlatformPan123
	}
	if reAliyun.MatchString(u) {
		return PlatformAliyun
	}
	if reXunlei.MatchString(u) {
		return PlatformXunlei
	}
	return PlatformUnknown
}

func (p TransferPlatform) String() string {
	switch p {
	case PlatformBaidu:
		return "baidu"
	case PlatformQuark:
		return "quark"
	case PlatformUC:
		return "uc"
	case PlatformPan115:
		return "pan115"
	case PlatformTianyi:
		return "tianyi"
	case PlatformPan123:
		return "pan123"
	case PlatformAliyun:
		return "aliyun"
	case PlatformXunlei:
		return "xunlei"
	default:
		return "unknown"
	}
}

// NormalizePan123Bearer 从配置中解析 Bearer（可粘贴整段 Cookie 或 authorization 行）
func NormalizePan123Bearer(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(s), "bearer ") {
		return s
	}
	// 有时用户粘贴多行 Cookie，尝试提取 authorization
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		low := strings.ToLower(line)
		if strings.HasPrefix(low, "authorization:") {
			v := strings.TrimSpace(line[len("authorization:"):])
			if v != "" && !strings.HasPrefix(strings.ToLower(v), "bearer ") {
				return "Bearer " + v
			}
			return v
		}
	}
	return "Bearer " + s
}
