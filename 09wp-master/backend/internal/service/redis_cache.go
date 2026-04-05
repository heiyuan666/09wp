package service

import (
	"context"
	"strconv"
	"time"

	"dfan-netdisk-backend/internal/config"
	"github.com/redis/go-redis/v9"
)

var searchRedisClient *redis.Client
var searchRedisEnabled bool
var searchRedisTTL time.Duration

func InitSearchRedisCache(cfg config.RedisConfig) error {
	if !cfg.Enabled {
		searchRedisEnabled = false
		return nil
	}

	searchRedisTTL = time.Duration(cfg.SearchTTL) * time.Second
	searchRedisClient = redis.NewClient(&redis.Options{
		Addr:        cfg.Host + ":" + strconv.Itoa(cfg.Port),
		Username:    cfg.Username,
		Password:    cfg.Password,
		DialTimeout: time.Duration(cfg.ConnectTTLMS) * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.PingTimeout)*time.Second)
	defer cancel()

	if err := searchRedisClient.Ping(ctx).Err(); err != nil {
		// 兜底：redis 不可用则直接降级到无缓存
		searchRedisEnabled = false
		searchRedisClient = nil
		return err
	}
	searchRedisEnabled = true
	return nil
}

func isSearchRedisEnabled() bool {
	return searchRedisEnabled && searchRedisClient != nil
}

// GetSearchCache 读取搜索缓存（返回命中与否）
func GetSearchCache(ctx context.Context, key string) ([]byte, bool) {
	if !isSearchRedisEnabled() {
		return nil, false
	}
	b, err := searchRedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return b, true
}

// SetSearchCache 写入搜索缓存
func SetSearchCache(ctx context.Context, key string, value []byte) {
	if !isSearchRedisEnabled() {
		return
	}
	_ = searchRedisClient.Set(ctx, key, value, searchRedisTTL).Err()
}

// DeleteSearchCache 删除缓存（用于失效控制）
func DeleteSearchCache(ctx context.Context, key string) {
	if !isSearchRedisEnabled() {
		return
	}
	_ = searchRedisClient.Del(ctx, key).Err()
}
