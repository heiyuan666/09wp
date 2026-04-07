package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var doubanURLRegex = regexp.MustCompile(`https?://movie\.douban\.com/subject/\d+/?`)

type doubanSearchCachePayload struct {
	Enabled bool           `json:"enabled"`
	Item    map[string]any `json:"item"`
}

func doubanNormalizeKeyword(q string) string {
	return strings.ToLower(strings.TrimSpace(q))
}

func doubanExtractURL(v any) string {
	switch t := v.(type) {
	case string:
		m := doubanURLRegex.FindString(t)
		if m != "" {
			return m
		}
	case []any:
		for _, it := range t {
			if u := doubanExtractURL(it); u != "" {
				return u
			}
		}
	case map[string]any:
		for _, it := range t {
			if u := doubanExtractURL(it); u != "" {
				return u
			}
		}
	}
	return ""
}

func doubanSearchURLByWpy(baseURL, keyword string) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "https://api.iyuns.com"
	}
	endpoint := base + "/api/wpysso?kw=" + url.QueryEscape(keyword)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("wpysso HTTP %d", resp.StatusCode)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}
	return doubanExtractURL(payload), nil
}

func doubanFetchDetail(baseURL, doubanURL string) (map[string]any, error) {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "https://api.iyuns.com"
	}
	endpoint := base + "/api/dbys?url=" + url.QueryEscape(doubanURL)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("dbys HTTP %d", resp.StatusCode)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// PublicDoubanSearch 按关键词检索豆瓣信息卡（带 DB 缓存）
func PublicDoubanSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		response.Error(c, 400, "缺少关键词")
		return
	}
	keyword := doubanNormalizeKeyword(q)
	cacheKey := "public:douban:search:" + keyword
	if b, ok := service.GetSearchCache(context.Background(), cacheKey); ok {
		var payload doubanSearchCachePayload
		if err := json.Unmarshal(b, &payload); err == nil {
			response.OK(c, payload)
			return
		}
	}
	apiBaseURL := "https://api.iyuns.com"
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil {
		if v := strings.TrimSpace(cfg.IYunsAPIBaseURL); v != "" {
			apiBaseURL = v
		}
	}

	var cached model.DoubanSearchCache
	if err := database.DB().Where("keyword = ?", keyword).First(&cached).Error; err == nil {
		if !cached.HasItem {
			payload := doubanSearchCachePayload{Enabled: true, Item: nil}
			if raw, err := json.Marshal(payload); err == nil {
				service.SetSearchCache(context.Background(), cacheKey, raw)
			}
			response.OK(c, payload)
			return
		}
		payload := doubanSearchCachePayload{
			Enabled: true,
			Item: map[string]any{
				"title":    cached.Title,
				"overview": cached.Overview,
				"poster":   cached.Poster,
				"year":     cached.Year,
				"rating":   cached.Rating,
				"url":      cached.DoubanURL,
			},
		}
		if raw, err := json.Marshal(payload); err == nil {
			service.SetSearchCache(context.Background(), cacheKey, raw)
		}
		response.OK(c, payload)
		return
	} else if err != gorm.ErrRecordNotFound {
		// ignore cache error
	}

	doubanURL, err := doubanSearchURLByWpy(apiBaseURL, q)
	if err != nil {
		response.Error(c, 500, "豆瓣检索失败")
		return
	}
	if doubanURL == "" {
		_ = database.DB().Where("keyword = ?", keyword).Assign(model.DoubanSearchCache{
			Keyword:   keyword,
			HasItem:   false,
			FetchedAt: time.Now(),
		}).FirstOrCreate(&model.DoubanSearchCache{}).Error
		payload := doubanSearchCachePayload{Enabled: true, Item: nil}
		if raw, err := json.Marshal(payload); err == nil {
			service.SetSearchCache(context.Background(), cacheKey, raw)
		}
		response.OK(c, payload)
		return
	}

	detail, err := doubanFetchDetail(apiBaseURL, doubanURL)
	if err != nil {
		response.Error(c, 500, "豆瓣详情获取失败")
		return
	}
	title := strings.TrimSpace(fmt.Sprintf("%v", detail["chinese_title"]))
	if title == "" {
		title = strings.TrimSpace(fmt.Sprintf("%v", detail["title"]))
	}
	overview := strings.TrimSpace(fmt.Sprintf("%v", detail["introduction"]))
	poster := strings.TrimSpace(fmt.Sprintf("%v", detail["poster"]))
	year := strings.TrimSpace(fmt.Sprintf("%v", detail["year"]))
	rating := strings.TrimSpace(fmt.Sprintf("%v", detail["douban_rating"]))
	if rating == "" {
		rating = strings.TrimSpace(fmt.Sprintf("%v", detail["douban_rating_average"]))
	}
	if rating == "<nil>" {
		rating = ""
	}

	_ = database.DB().Where("keyword = ?", keyword).Assign(model.DoubanSearchCache{
		Keyword:   keyword,
		HasItem:   title != "",
		DoubanURL: doubanURL,
		Title:     title,
		Overview:  overview,
		Poster:    poster,
		Year:      year,
		Rating:    rating,
		FetchedAt: time.Now(),
	}).FirstOrCreate(&model.DoubanSearchCache{}).Error

	if title == "" {
		payload := doubanSearchCachePayload{Enabled: true, Item: nil}
		if raw, err := json.Marshal(payload); err == nil {
			service.SetSearchCache(context.Background(), cacheKey, raw)
		}
		response.OK(c, payload)
		return
	}
	payload := doubanSearchCachePayload{
		Enabled: true,
		Item: map[string]any{
			"title":    title,
			"overview": overview,
			"poster":   poster,
			"year":     year,
			"rating":   rating,
			"url":      doubanURL,
		},
	}
	if raw, err := json.Marshal(payload); err == nil {
		service.SetSearchCache(context.Background(), cacheKey, raw)
	}
	response.OK(c, payload)
}

