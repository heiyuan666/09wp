package validator

import (
	"PanCheck/internal/model"
	"regexp"
)

// LinkInfo 链接信息
type LinkInfo struct {
	Link     string
	Platform model.Platform
}

// ParseLink 解析链接，识别平台类型
func ParseLink(link string) LinkInfo {
	// 去除首尾空格
	link = trimSpace(link)
	if link == "" {
		return LinkInfo{Link: link, Platform: model.PlatformUnknown}
	}

	// 各平台链接匹配规则（支持查询参数和锚点）
	patterns := map[model.Platform]*regexp.Regexp{
		model.PlatformQuark:  regexp.MustCompile(`(?i)(?:https?://)?(?:pan\.quark\.cn|quark\.cn|pan\.qoark\.cn)/s/[a-zA-Z0-9]+`),
		model.PlatformUC:     regexp.MustCompile(`(?i)(?:https?://)?(?:drive\.uc\.cn|yun\.uc\.cn|uc\.cn)/s/[a-zA-Z0-9]+`),
		model.PlatformBaidu:  regexp.MustCompile(`(?i)(?:https?://)?(?:pan\.baidu\.com)/s/[a-zA-Z0-9_-]+`),
		model.PlatformTianyi: regexp.MustCompile(`(?i)(?:https?://)?(?:cloud\.189\.cn|h5\.cloud\.189\.cn)/(?:t/[a-zA-Z0-9]+|web/share\?code=[a-zA-Z0-9]+|share\.html#/t/[a-zA-Z0-9]+)`),
		model.PlatformPan123: regexp.MustCompile(`(?i)(?:https?://)?(?:123pan\.com|123pan\.cn|123684\.com|123685\.com|123912\.com|123592\.com|123865\.com)/s/[a-zA-Z0-9-]+`),
		model.PlatformPan115: regexp.MustCompile(`(?i)(?:https?://)?(?:115\.com|115cdn\.com|anxia\.com)/s/[a-zA-Z0-9]+`),
		model.PlatformAliyun: regexp.MustCompile(`(?i)(?:https?://)?(?:www\.aliyundrive\.com|aliyundrive\.com|www\.alipan\.com)/s/[a-zA-Z0-9]+`),
		model.PlatformXunlei: regexp.MustCompile(`(?i)(?:https?://)?(?:pan\.xunlei\.com)/s/[a-zA-Z0-9_-]+`),
		model.PlatformCMCC:   regexp.MustCompile(`(?i)(?:https?://)?(?:yun\.139\.com/shareweb/#/w/i/|caiyun\.139\.com/m/i\?)[a-zA-Z0-9]+`),
	}

	for platform, pattern := range patterns {
		if pattern.MatchString(link) {
			return LinkInfo{Link: normalizeLink(link), Platform: platform}
		}
	}

	return LinkInfo{Link: link, Platform: model.PlatformUnknown}
}

// ParseLinks 批量解析链接
func ParseLinks(links []string) []LinkInfo {
	result := make([]LinkInfo, 0, len(links))
	for _, link := range links {
		info := ParseLink(link)
		if info.Platform != model.PlatformUnknown {
			result = append(result, info)
		}
	}
	return result
}

// normalizeLink 规范化链接格式
func normalizeLink(link string) string {
	// 确保链接以http://或https://开头
	if !regexp.MustCompile(`^https?://`).MatchString(link) {
		return "https://" + link
	}
	return link
}

// trimSpace 去除首尾空格
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
