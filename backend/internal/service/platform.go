package service

import "strings"

// DetectPlatformFromLink 统一从分享链接识别平台（用于搜索过滤/索引字段）。
// 注意：这里的返回值要与前端/接口约定保持一致。
func DetectPlatformFromLink(link string) string {
	u := strings.ToLower(strings.TrimSpace(link))
	switch {
	case strings.Contains(u, "pan.baidu.com"):
		return "baidu"
	case strings.Contains(u, "pan.quark.cn"):
		return "quark"
	case strings.Contains(u, "pan.xunlei.com"):
		return "xunlei"
	case strings.Contains(u, "aliyundrive.com"), strings.Contains(u, "alipan.com"):
		return "aliyun"
	case strings.Contains(u, "cloud.189.cn"), strings.Contains(u, "caiyun.189"), strings.Contains(u, "tianyi"):
		return "tianyi"
	case strings.Contains(u, "yun.139.com"), strings.Contains(u, "caiyun.139.com"):
		return "yidong"
	case strings.Contains(u, "drive-h.uc.cn"), strings.Contains(u, "drive.uc.cn"):
		return "uc"
	case strings.Contains(u, "115.com"), strings.Contains(u, "115cdn.com"):
		return "pan115"
	case strings.Contains(u, "123pan"), strings.Contains(u, "123684"), strings.Contains(u, "123685"),
		strings.Contains(u, "123912"), strings.Contains(u, "123592"), strings.Contains(u, "123865"), strings.Contains(u, "123.net"):
		return "pan123"
	default:
		return "other"
	}
}
