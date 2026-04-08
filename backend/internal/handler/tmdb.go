package handler

import (
	"context"
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type tmdbItem struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Name        string  `json:"name"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	Backdrop    string  `json:"backdrop_path"`
	ReleaseDate string  `json:"release_date"`
	FirstAir    string  `json:"first_air_date"`
	VoteAverage float64 `json:"vote_average"`
	MediaType   string  `json:"media_type"`
}

type tmdbSearchCachePayload struct {
	Enabled bool          `json:"enabled"`
	Item    map[string]any `json:"item"`
}

func tmdbSearchWithToken(token, q, proxyURL string) (*tmdbItem, error) {
	transport := &http.Transport{}
	if strings.TrimSpace(proxyURL) != "" {
		pu, err := url.Parse(strings.TrimSpace(proxyURL))
		if err != nil {
			return nil, fmt.Errorf("tmdb 代理地址无效: %w", err)
		}
		transport.Proxy = http.ProxyURL(pu)
	}
	client := &http.Client{
		Timeout:   12 * time.Second,
		Transport: transport,
	}
	paths := []struct {
		path      string
		mediaType string
	}{
		{path: "movie", mediaType: "movie"},
		{path: "tv", mediaType: "tv"},
	}
	for _, p := range paths {
		endpoint := "https://api.themoviedb.org/3/search/" + p.path + "?language=zh-CN&include_adult=false&page=1&query=" + url.QueryEscape(q)
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("tmdb 请求构建失败: %w", err)
		}
		req.Header.Set("accept", "application/json")
		req.Header.Set("authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("tmdb 请求失败: %w", err)
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode >= 400 {
			body := strings.TrimSpace(string(raw))
			if len(body) > 300 {
				body = body[:300]
			}
			if body == "" {
				body = "(empty body)"
			}
			return nil, fmt.Errorf("tmdb 返回异常: HTTP %d %s", resp.StatusCode, body)
		}
		var payload struct {
			Results []tmdbItem `json:"results"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("tmdb 响应解析失败: %w", err)
		}
		for _, item := range payload.Results {
			title := strings.TrimSpace(item.Title)
			if title == "" {
				title = strings.TrimSpace(item.Name)
			}
			if title == "" {
				continue
			}
			item.MediaType = p.mediaType
			return &item, nil
		}
	}
	return nil, nil
}

func tmdbNormalizeKeyword(q string) string {
	return strings.ToLower(strings.TrimSpace(q))
}

// PublicTMDBSearch 按关键词搜索 TMDB（电影/剧集）
func PublicTMDBSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		response.Error(c, 400, "缺少关键词")
		return
	}
	token := ""
	proxyURL := ""
	cacheTTL := time.Duration(0)
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil {
		token = strings.TrimSpace(cfg.TMDBBearerToken)
		proxyURL = strings.TrimSpace(cfg.TMDBProxyURL)
		if cfg.TMDBSearchCacheTTL > 0 {
			cacheTTL = time.Duration(cfg.TMDBSearchCacheTTL) * time.Second
		}
	}
	if token == "" {
		token = strings.TrimSpace(os.Getenv("TMDB_BEARER_TOKEN"))
	}
	if token == "" {
		response.OK(c, gin.H{"enabled": false, "item": nil})
		return
	}
	keyword := tmdbNormalizeKeyword(q)
	cacheKey := "public:tmdb:search:" + keyword
	if b, ok := service.GetSearchCache(context.Background(), cacheKey); ok {
		var payload tmdbSearchCachePayload
		if err := json.Unmarshal(b, &payload); err == nil {
			response.OK(c, payload)
			return
		}
	}
	var cached model.TMDBSearchCache
	if err := database.DB().Where("keyword = ?", keyword).First(&cached).Error; err == nil {
		// 仅命中「有结果」时短路；HasItem=false 不短路，避免旧逻辑写死的「无结果」永久生效。
		if cached.HasItem {
			payload := tmdbSearchCachePayload{
				Enabled: true,
				Item: map[string]any{
					"id":           cached.ItemID,
					"title":        cached.Title,
					"overview":     strings.TrimSpace(cached.Overview),
					"poster":       cached.Poster,
					"backdrop":     cached.Backdrop,
					"release_date": cached.ReleaseDate,
					"rating":       cached.Rating,
					"media_type":   cached.MediaType,
					"url":          cached.URL,
				},
			}
			if raw, err := json.Marshal(payload); err == nil {
				service.SetSearchCacheWithTTL(context.Background(), cacheKey, raw, cacheTTL)
			}
			response.OK(c, payload)
			return
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 缓存异常不阻断主流程，继续直连 TMDB
	}
	item, err := tmdbSearchWithToken(token, q, proxyURL)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	if item != nil {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = strings.TrimSpace(item.Name)
		}
		poster := ""
		if p := strings.TrimSpace(item.PosterPath); p != "" {
			poster = "https://image.tmdb.org/t/p/w500" + p
		}
		backdrop := ""
		if p := strings.TrimSpace(item.Backdrop); p != "" {
			backdrop = "https://image.tmdb.org/t/p/w780" + p
		}
		releaseDate := strings.TrimSpace(item.ReleaseDate)
		if releaseDate == "" {
			releaseDate = strings.TrimSpace(item.FirstAir)
		}
		payload := tmdbSearchCachePayload{
			Enabled: true,
			Item: map[string]any{
				"id":           item.ID,
				"title":        title,
				"overview":     strings.TrimSpace(item.Overview),
				"poster":       poster,
				"backdrop":     backdrop,
				"release_date": releaseDate,
				"rating":       item.VoteAverage,
				"media_type":   item.MediaType,
				"url":          fmt.Sprintf("https://www.themoviedb.org/%s/%d", item.MediaType, item.ID),
			},
		}
		if raw, err := json.Marshal(payload); err == nil {
			service.SetSearchCacheWithTTL(context.Background(), cacheKey, raw, cacheTTL)
		}
		response.OK(c, payload)
		_ = database.DB().Where("keyword = ?", keyword).Assign(model.TMDBSearchCache{
			Keyword:     keyword,
			HasItem:     true,
			ItemID:      item.ID,
			Title:       title,
			Overview:    strings.TrimSpace(item.Overview),
			Poster:      poster,
			Backdrop:    backdrop,
			ReleaseDate: releaseDate,
			Rating:      item.VoteAverage,
			MediaType:   item.MediaType,
			URL:         fmt.Sprintf("https://www.themoviedb.org/%s/%d", item.MediaType, item.ID),
			FetchedAt:   time.Now(),
		}).FirstOrCreate(&model.TMDBSearchCache{}).Error
		return
	}
	_ = database.DB().Where("keyword = ?", keyword).Assign(model.TMDBSearchCache{
		Keyword:   keyword,
		HasItem:   false,
		FetchedAt: time.Now(),
	}).FirstOrCreate(&model.TMDBSearchCache{}).Error
	payload := tmdbSearchCachePayload{Enabled: true, Item: nil}
	if raw, err := json.Marshal(payload); err == nil {
		service.SetSearchCacheWithTTL(context.Background(), cacheKey, raw, cacheTTL)
	}
	response.OK(c, payload)
}

