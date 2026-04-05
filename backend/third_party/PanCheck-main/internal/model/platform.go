package model

// Platform 网盘平台类型
type Platform string

const (
	PlatformUnknown Platform = "unknown" // 未知平台
	PlatformQuark   Platform = "quark"   // 夸克网盘
	PlatformUC      Platform = "uc"      // UC网盘
	PlatformBaidu   Platform = "baidu"   // 百度网盘
	PlatformTianyi  Platform = "tianyi"  // 天翼云盘
	PlatformPan123  Platform = "pan123"  // 123网盘
	PlatformPan115  Platform = "pan115"  // 115网盘
	PlatformAliyun  Platform = "aliyun"  // 阿里云盘
	PlatformXunlei  Platform = "xunlei"  // 迅雷云盘
	PlatformCMCC    Platform = "cmcc"    // 中国移动云盘
)

// String 返回平台名称
func (p Platform) String() string {
	return string(p)
}

// IsValid 检查平台是否有效
func (p Platform) IsValid() bool {
	switch p {
	case PlatformQuark, PlatformUC, PlatformBaidu, PlatformTianyi,
		PlatformPan123, PlatformPan115, PlatformAliyun, PlatformXunlei, PlatformCMCC:
		return true
	default:
		return false
	}
}

// AllPlatforms 返回所有支持的平台
func AllPlatforms() []Platform {
	return []Platform{
		PlatformQuark,
		PlatformUC,
		PlatformBaidu,
		PlatformTianyi,
		PlatformPan123,
		PlatformPan115,
		PlatformAliyun,
		PlatformXunlei,
		PlatformCMCC,
	}
}
