package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetPublicDoubanHot 豆瓣热门榜单（前台公开）
func GetPublicDoubanHot(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "16"))
	if limit <= 0 {
		limit = 16
	}
	if limit > 50 {
		limit = 50
	}

	// 默认值：movie + 热门（可扩展为前端可选）
	doubanType := strings.TrimSpace(c.DefaultQuery("type", "movie"))
	doubanTag := strings.TrimSpace(c.DefaultQuery("tag", "热门"))

	cacheKey := fmt.Sprintf("public:douban-hot:v1:%s:%s:%d", doubanType, doubanTag, limit)
	if b, ok := service.GetSearchCache(context.Background(), cacheKey); ok {
		var cached []service.DoubanHotItem
		if err := json.Unmarshal(b, &cached); err == nil {
			response.OK(c, gin.H{"list": cached})
			return
		}
	}

	items, err := service.GetDoubanHotItems(doubanType, doubanTag, limit)
	if err != nil {
		// 公开接口：失败兜底返回空，避免前台整页失败
		response.OK(c, gin.H{"list": []any{}})
		return
	}

	// 写入缓存（短 TTL）
	if raw, err := json.Marshal(items); err == nil {
		service.SetSearchCache(context.Background(), cacheKey, raw)
	}

	// 返回结构尽量保持与热搜一致（前端同样用 tags）
	response.OK(c, gin.H{"list": items})
}
