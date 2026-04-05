package checker

import (
	"PanCheck/internal/config"
	"PanCheck/internal/model"
	"sync"
	"time"
)

// BaseChecker 基础检测器
type BaseChecker struct {
	platform         model.Platform
	concurrencyLimit int
	timeout          time.Duration
	rateConfig       *config.PlatformRateConfig

	// 频率控制相关
	lastRequestTime time.Time
	requestMutex    sync.Mutex
	tokenBucket     *TokenBucket
}

// TokenBucket 令牌桶，用于控制请求频率
type TokenBucket struct {
	capacity   int           // 桶容量
	tokens     int           // 当前令牌数
	refillRate time.Duration // 补充速率（每个令牌的间隔）
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(maxRequestsPerSecond int) *TokenBucket {
	if maxRequestsPerSecond <= 0 {
		return nil // 不限制
	}

	refillRate := time.Second / time.Duration(maxRequestsPerSecond)
	return &TokenBucket{
		capacity:   maxRequestsPerSecond,
		tokens:     maxRequestsPerSecond,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Take 尝试获取一个令牌，如果成功返回true
func (tb *TokenBucket) Take() bool {
	if tb == nil {
		return true // 无限制
	}

	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed / tb.refillRate)
	if tokensToAdd > 0 {
		tb.tokens = minInt(tb.tokens+tokensToAdd, tb.capacity)
		tb.lastRefill = now
	}

	// 尝试获取令牌
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// minInt 返回两个整数中的较小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewBaseChecker 创建基础检测器
func NewBaseChecker(platform model.Platform, concurrencyLimit int, timeout time.Duration) *BaseChecker {
	return &BaseChecker{
		platform:         platform,
		concurrencyLimit: concurrencyLimit,
		timeout:          timeout,
		lastRequestTime:  time.Now(),
	}
}

// SetRateConfig 设置频率配置
func (b *BaseChecker) SetRateConfig(config *config.PlatformRateConfig) {
	b.rateConfig = config
	if config != nil {
		if config.Concurrency > 0 {
			b.concurrencyLimit = config.Concurrency
		}
		if config.MaxRequestsPerSecond > 0 {
			b.tokenBucket = NewTokenBucket(config.MaxRequestsPerSecond)
		}
	}
}

// GetPlatform 返回平台类型
func (b *BaseChecker) GetPlatform() model.Platform {
	return b.platform
}

// GetConcurrencyLimit 返回并发限制数
func (b *BaseChecker) GetConcurrencyLimit() int {
	return b.concurrencyLimit
}

// GetTimeout 返回超时时间
func (b *BaseChecker) GetTimeout() time.Duration {
	return b.timeout
}

// GetRateConfig 返回频率配置
func (b *BaseChecker) GetRateConfig() *config.PlatformRateConfig {
	return b.rateConfig
}

// ApplyRateLimit 应用频率限制，确保请求间隔符合配置
func (b *BaseChecker) ApplyRateLimit() {
	if b.rateConfig == nil {
		return
	}

	b.requestMutex.Lock()
	defer b.requestMutex.Unlock()

	// 应用请求间隔
	if b.rateConfig.RequestDelayMs > 0 {
		elapsed := time.Since(b.lastRequestTime)
		requiredDelay := time.Duration(b.rateConfig.RequestDelayMs) * time.Millisecond
		if elapsed < requiredDelay {
			time.Sleep(requiredDelay - elapsed)
		}
	}

	// 应用令牌桶限制
	if b.tokenBucket != nil {
		for !b.tokenBucket.Take() {
			// 等待令牌可用
			time.Sleep(b.tokenBucket.refillRate)
		}
	}

	b.lastRequestTime = time.Now()
}
