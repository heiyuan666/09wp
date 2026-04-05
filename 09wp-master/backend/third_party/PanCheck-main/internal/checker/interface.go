package checker

import "PanCheck/internal/model"

// CheckResult 检测结果
type CheckResult struct {
	Valid         bool   // 是否有效
	FailureReason string // 失败原因（如果无效）
	Duration      int64  // 检测耗时（毫秒）
	IsRateLimited bool   // 是否被平台限制
}

// LinkChecker 链接检测器接口
type LinkChecker interface {
	// Check 检测链接是否有效
	Check(link string) (*CheckResult, error)

	// GetPlatform 返回平台类型
	GetPlatform() model.Platform

	// GetConcurrencyLimit 返回并发限制数
	GetConcurrencyLimit() int
}

// CheckerFactory 检测器工厂
type CheckerFactory struct {
	checkers map[model.Platform]LinkChecker
}

// NewCheckerFactory 创建检测器工厂
func NewCheckerFactory() *CheckerFactory {
	return &CheckerFactory{
		checkers: make(map[model.Platform]LinkChecker),
	}
}

// Register 注册检测器
func (f *CheckerFactory) Register(checker LinkChecker) {
	f.checkers[checker.GetPlatform()] = checker
}

// GetChecker 获取指定平台的检测器
func (f *CheckerFactory) GetChecker(platform model.Platform) (LinkChecker, bool) {
	checker, ok := f.checkers[platform]
	return checker, ok
}

// GetAllCheckers 获取所有检测器
func (f *CheckerFactory) GetAllCheckers() map[model.Platform]LinkChecker {
	return f.checkers
}
