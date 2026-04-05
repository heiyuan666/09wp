package utils

import (
	"regexp"
	"strings"
)

// DeviceInfo 设备信息
type DeviceInfo struct {
	Browser  string // 浏览器
	OS       string // 操作系统
	Device   string // 设备类型（desktop/mobile）
	Language string // 语言
}

// ParseDeviceInfo 从请求头解析设备信息
func ParseDeviceInfo(userAgent, acceptLanguage string) DeviceInfo {
	info := DeviceInfo{
		Browser:  parseBrowser(userAgent),
		OS:       parseOS(userAgent),
		Device:   parseDevice(userAgent),
		Language: parseLanguage(acceptLanguage),
	}
	return info
}

// parseBrowser 解析浏览器类型
func parseBrowser(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	// Chrome
	if strings.Contains(userAgent, "chrome") && !strings.Contains(userAgent, "edg") && !strings.Contains(userAgent, "opr") {
		if strings.Contains(userAgent, "chromium") {
			return "chrome"
		}
		return "chrome"
	}

	// Edge
	if strings.Contains(userAgent, "edg") {
		return "edge-chromium"
	}

	// Firefox
	if strings.Contains(userAgent, "firefox") {
		return "firefox"
	}

	// Safari
	if strings.Contains(userAgent, "safari") && !strings.Contains(userAgent, "chrome") {
		return "safari"
	}

	// Opera
	if strings.Contains(userAgent, "opr") || strings.Contains(userAgent, "opera") {
		return "opera"
	}

	// iOS Safari
	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		return "ios"
	}

	// Android
	if strings.Contains(userAgent, "android") {
		return "android"
	}

	return "unknown"
}

// parseOS 解析操作系统
func parseOS(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	// Windows
	if strings.Contains(userAgent, "windows") {
		if strings.Contains(userAgent, "windows nt 10.0") {
			return "Windows 10"
		}
		if strings.Contains(userAgent, "windows nt 6.3") {
			return "Windows 8.1"
		}
		if strings.Contains(userAgent, "windows nt 6.2") {
			return "Windows 8"
		}
		if strings.Contains(userAgent, "windows nt 6.1") {
			return "Windows 7"
		}
		return "Windows"
	}

	// macOS
	if strings.Contains(userAgent, "mac os x") || strings.Contains(userAgent, "macintosh") {
		return "macOS"
	}

	// iOS
	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		return "iOS"
	}

	// Android
	if strings.Contains(userAgent, "android") {
		// 尝试提取 Android 版本
		re := regexp.MustCompile(`android\s+([\d.]+)`)
		matches := re.FindStringSubmatch(userAgent)
		if len(matches) > 1 {
			return "Android " + matches[1]
		}
		return "Android"
	}

	// Linux
	if strings.Contains(userAgent, "linux") {
		return "Linux"
	}

	return "unknown"
}

// parseDevice 解析设备类型
func parseDevice(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	// 移动设备
	if strings.Contains(userAgent, "mobile") ||
		strings.Contains(userAgent, "iphone") ||
		strings.Contains(userAgent, "ipad") ||
		strings.Contains(userAgent, "android") {
		return "mobile"
	}

	// 桌面设备
	return "desktop"
}

// parseLanguage 解析语言
func parseLanguage(acceptLanguage string) string {
	if acceptLanguage == "" {
		return "zh-CN"
	}

	// 取第一个语言代码
	parts := strings.Split(acceptLanguage, ",")
	if len(parts) > 0 {
		lang := strings.TrimSpace(parts[0])
		// 移除质量值（如 zh-CN;q=0.9）
		if idx := strings.Index(lang, ";"); idx != -1 {
			lang = lang[:idx]
		}
		return lang
	}

	return "zh-CN"
}
