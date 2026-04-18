// Package netdiskurl 识别常见网盘分享链接，用于公开接口归一化展示。
package netdiskurl

import (
	"regexp"
	"strings"
)

// IsNetdiskURL 判断是否为常见网盘分享 URL（与 RSS/TG 抓取规则对齐并略扩展）。
func IsNetdiskURL(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" || (!strings.HasPrefix(v, "http://") && !strings.HasPrefix(v, "https://")) {
		return false
	}
	return strings.Contains(v, "pan.quark.cn/") ||
		strings.Contains(v, "pan.baidu.com/") ||
		strings.Contains(v, "aliyundrive.com/") ||
		strings.Contains(v, "alipan.com/") ||
		strings.Contains(v, "pan.xunlei.com/") ||
		strings.Contains(v, "drive-h.uc.cn/") ||
		strings.Contains(v, "drive.uc.cn/") ||
		strings.Contains(v, "yun.uc.cn/") ||
		strings.Contains(v, "cloud.189.cn/") ||
		strings.Contains(v, "115.com/") ||
		strings.Contains(v, "115cdn.com/") ||
		strings.Contains(v, "123pan.") ||
		strings.Contains(v, "123684.com/") ||
		strings.Contains(v, "123685.com/") ||
		strings.Contains(v, "123.net/") ||
		strings.Contains(v, "yun.139.com/") ||
		strings.Contains(v, "caiyun.139.com/")
}

var urlInText = regexp.MustCompile(`https?://[^\s<>"']+`)

// ExtractFromText 从正文里提取网盘链接（去重保序）。
func ExtractFromText(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	for _, m := range urlInText.FindAllString(text, -1) {
		u := trimURLTail(m)
		if u == "" || !IsNetdiskURL(u) {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func trimURLTail(s string) string {
	s = strings.TrimSpace(s)
	for len(s) > 0 {
		last := s[len(s)-1]
		if last == '.' || last == ',' || last == ')' || last == ']' || last == '}' || last == '"' || last == '\'' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return strings.TrimSpace(s)
}

// MergeUnique 合并去重，保留 a 顺序再追加 b 中未出现项。
func MergeUnique(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	for _, u := range a {
		add(u)
	}
	for _, u := range b {
		add(u)
	}
	return out
}

// NormalizeSlices 将误填在直链里的网盘 URL 挪到网盘列表，并去重。
func NormalizeSlices(direct, pan []string) (outDirect, outPan []string) {
	seenPan := make(map[string]struct{})
	var panOut []string
	addPan := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" {
			return
		}
		if _, ok := seenPan[u]; ok {
			return
		}
		seenPan[u] = struct{}{}
		panOut = append(panOut, u)
	}
	var directOut []string
	seenDirect := make(map[string]struct{})
	addDirect := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" {
			return
		}
		if _, ok := seenDirect[u]; ok {
			return
		}
		seenDirect[u] = struct{}{}
		directOut = append(directOut, u)
	}

	for _, raw := range direct {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if IsNetdiskURL(raw) {
			addPan(raw)
		} else {
			addDirect(raw)
		}
	}
	for _, raw := range pan {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		addPan(raw)
	}
	return directOut, panOut
}
