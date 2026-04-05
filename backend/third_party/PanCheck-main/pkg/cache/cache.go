package cache

import (
	"PanCheck/internal/checker"
	"PanCheck/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheRepository 缓存仓库接口
type CacheRepository interface {
	// Get 从缓存获取检测结果
	Get(ctx context.Context, link string) (*checker.CheckResult, error)
	// Set 存入缓存，根据有效/无效和平台设置不同TTL
	Set(ctx context.Context, link string, result *checker.CheckResult, platform model.Platform, invalidTTL int, platformTTLMap map[model.Platform]int) error
	// Delete 删除缓存
	Delete(ctx context.Context, link string) error
	// Close 关闭连接
	Close() error
	// IsEnabled 检查缓存是否启用
	IsEnabled() bool
}

// CacheConfig Redis缓存配置
type CacheConfig struct {
	Enabled    bool
	Host       string
	Port       int
	Username   string // Redis用户名（Redis 6.0+ ACL支持，留空则只使用密码）
	Password   string
	InvalidTTL int // 无效链接过期时间（小时）
}

// redisCacheRepository Redis缓存仓库实现
type redisCacheRepository struct {
	client     *redis.Client
	enabled    bool
	invalidTTL int // 无效链接过期时间（小时）
}

// NewCacheRepository 创建缓存仓库
func NewCacheRepository(config CacheConfig) (CacheRepository, error) {
	if !config.Enabled {
		log.Println("Redis cache is disabled")
		return &redisCacheRepository{
			enabled: false,
		}, nil
	}

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       0, // 固定使用0号数据库
	}
	// Redis 6.0+ 支持ACL，可以使用用户名+密码
	if config.Username != "" {
		opts.Username = config.Username
	}
	rdb := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Failed to connect to Redis: %v, cache will be disabled", err)
		return &redisCacheRepository{
			enabled: false,
		}, nil
	}

	log.Println("Redis cache connected successfully")
	return &redisCacheRepository{
		client:     rdb,
		enabled:    true,
		invalidTTL: config.InvalidTTL,
	}, nil
}

// IsEnabled 检查缓存是否启用
func (r *redisCacheRepository) IsEnabled() bool {
	return r.enabled && r.client != nil
}

// Get 从缓存获取检测结果
func (r *redisCacheRepository) Get(ctx context.Context, link string) (*checker.CheckResult, error) {
	if !r.IsEnabled() {
		return nil, nil
	}

	key := fmt.Sprintf("link:check:%s", link)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// 缓存未命中
		return nil, nil
	}
	if err != nil {
		log.Printf("Failed to get cache for link %s: %v", link, err)
		return nil, err
	}

	var result checker.CheckResult
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		log.Printf("Failed to unmarshal cache result for link %s: %v", link, err)
		// 删除损坏的缓存
		r.client.Del(ctx, key)
		return nil, err
	}

	return &result, nil
}

// Set 存入缓存，根据有效/无效和平台设置不同TTL
func (r *redisCacheRepository) Set(ctx context.Context, link string, result *checker.CheckResult, platform model.Platform, invalidTTL int, platformTTLMap map[model.Platform]int) error {
	if !r.IsEnabled() {
		return nil
	}

	key := fmt.Sprintf("link:check:%s", link)
	val, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// 根据有效/无效和平台设置TTL
	var ttl time.Duration
	if result.Valid {
		// 有效链接：使用平台配置的TTL
		if ttlHours, ok := platformTTLMap[platform]; ok && ttlHours > 0 {
			ttl = time.Duration(ttlHours) * time.Hour
		} else {
			// 默认24小时
			ttl = 24 * time.Hour
		}
	} else {
		// 无效链接：使用统一配置的TTL
		if invalidTTL > 0 {
			ttl = time.Duration(invalidTTL) * time.Hour
		} else {
			// 默认7天
			ttl = 168 * time.Hour
		}
	}

	if err := r.client.Set(ctx, key, val, ttl).Err(); err != nil {
		log.Printf("Failed to set cache for link %s: %v", link, err)
		return err
	}

	return nil
}

// Delete 删除缓存
func (r *redisCacheRepository) Delete(ctx context.Context, link string) error {
	if !r.IsEnabled() {
		return nil
	}

	key := fmt.Sprintf("link:check:%s", link)
	return r.client.Del(ctx, key).Err()
}

// Close 关闭连接
func (r *redisCacheRepository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
