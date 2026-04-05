package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type DoubanSubject struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Cover string `json:"cover"`
	URL   string `json:"url"`
}

type doubanHotResp struct {
	// CloudSaver DoubanService.ts：response.data.subjects
	Subjects []DoubanSubject `json:"subjects"`
}

type DoubanHotItem struct {
	Title string `json:"title"`
	Cover string `json:"cover"`
	URL   string `json:"url"`
}

// GetDoubanHotItems 拉取豆瓣热门榜单（movie.douban.com/j/search_subjects）
func GetDoubanHotItems(doubanType, doubanTag string, pageLimit int) ([]DoubanHotItem, error) {
	if doubanType == "" {
		doubanType = "movie"
	}
	if doubanTag == "" {
		doubanTag = "热门"
	}
	if pageLimit < 1 {
		pageLimit = 16
	}
	if pageLimit > 50 {
		pageLimit = 50
	}

	q := url.Values{}
	q.Set("type", doubanType)
	q.Set("tag", doubanTag)
	q.Set("page_limit", strconv.Itoa(pageLimit))
	q.Set("page_start", "0")

	endpoint := "https://movie.douban.com/j/search_subjects?" + q.Encode()
	client := &http.Client{Timeout: 12 * time.Second}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	// 尽量模仿浏览器请求（避免被直接风控）
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", "https://movie.douban.com/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("豆瓣热门请求失败: http %d", resp.StatusCode)
	}

	var out doubanHotResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("豆瓣热门返回解析失败: %w", err)
	}
	var items []DoubanHotItem
	for _, s := range out.Subjects {
		if s.Title == "" {
			continue
		}
		// 后端屏蔽词：命中则不展示
		if IsKeywordBlockedText(s.Title) {
			continue
		}
		items = append(items, DoubanHotItem{
			Title: s.Title,
			Cover: s.Cover,
			URL:   s.URL,
		})
	}
	return items, nil
}

