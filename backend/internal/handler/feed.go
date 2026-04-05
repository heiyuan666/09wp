package handler

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"github.com/gin-gonic/gin"
)

type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	LastBuild   string    `xml:"lastBuildDate"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description,omitempty"`
}

// RSS 输出最近资源（公开）
func RSS(c *gin.Context) {
	base := buildSiteBaseURL(c)
	var cfg model.SystemConfig
	_ = database.DB().Order("id ASC").First(&cfg).Error

	var list []model.Resource
	if err := database.DB().
		Model(&model.Resource{}).
		Where("status = 1").
		Order("created_at DESC, id DESC").
		Limit(20).
		Find(&list).Error; err != nil {
		// RSS 不应返回 JSON，这里兜底返回空 RSS
		list = nil
	}

	items := make([]rssItem, 0, len(list))
	for _, r := range list {
		u := fmt.Sprintf("%s/r/%d", base, r.ID)
		desc := strings.TrimSpace(r.Description)
		if len(desc) > 800 {
			desc = desc[:800]
		}
		items = append(items, rssItem{
			Title:       sanitizeXMLText(r.Title),
			Link:        u,
			GUID:        u,
			PubDate:     r.CreatedAt.Format(time.RFC1123Z),
			Description: sanitizeXMLText(desc),
		})
	}

	ch := rssChannel{
		Title:       sanitizeXMLText(firstNonEmpty(cfg.SiteTitle, "网盘资源导航")),
		Link:        base + "/",
		Description: sanitizeXMLText(firstNonEmpty(cfg.SeoDescription, "最新网盘资源")),
		Language:    "zh-CN",
		LastBuild:   time.Now().Format(time.RFC1123Z),
		Items:       items,
	}
	out := rss{Version: "2.0", Channel: ch}

	c.Header("Content-Type", "application/rss+xml; charset=utf-8")
	c.Status(http.StatusOK)
	enc := xml.NewEncoder(c.Writer)
	enc.Indent("", "  ")
	_ = enc.Encode(out)
}

// Sitemap 输出站点 sitemap.xml（公开）
func Sitemap(c *gin.Context) {
	base := buildSiteBaseURL(c)

	type urlEntry struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod,omitempty"`
	}
	type urlset struct {
		XMLName xml.Name   `xml:"urlset"`
		Xmlns   string     `xml:"xmlns,attr"`
		URLs    []urlEntry `xml:"url"`
	}

	urls := make([]urlEntry, 0, 2000)
	urls = append(urls,
		urlEntry{Loc: base + "/"},
		urlEntry{Loc: base + "/search"},
	)

	// 分类页（只输出启用的）
	var cats []model.Category
	_ = database.DB().Model(&model.Category{}).Where("status = 1").Order("id ASC").Find(&cats).Error
	for _, ccat := range cats {
		if strings.TrimSpace(ccat.Slug) == "" {
			continue
		}
		urls = append(urls, urlEntry{
			Loc:     fmt.Sprintf("%s/c/%s", base, strings.TrimSpace(ccat.Slug)),
			LastMod: ccat.UpdatedAt.Format("2006-01-02"),
		})
	}

	// 资源详情页（只输出显示的，避免过大：最多 5000 条）
	var res []model.Resource
	_ = database.DB().Model(&model.Resource{}).Where("status = 1").Order("id DESC").Limit(5000).Find(&res).Error
	for _, r := range res {
		urls = append(urls, urlEntry{
			Loc:     fmt.Sprintf("%s/r/%d", base, r.ID),
			LastMod: r.UpdatedAt.Format("2006-01-02"),
		})
	}

	// 号卡专区（公开）
	// 只输出前台展示的（status=1 & flag=true），避免过大：最多 5000 条
	var haokaProducts []model.HaokaProduct
	_ = database.DB().
		Model(&model.HaokaProduct{}).
		Where("status = 1").
		Where("flag = ?", true).
		Order("id DESC").
		Limit(5000).
		Find(&haokaProducts).Error
	if len(haokaProducts) > 0 {
		// /haoka 列表页：取最新更新时间
		urls = append(urls, urlEntry{
			Loc:     fmt.Sprintf("%s/haoka", base),
			LastMod: haokaProducts[0].UpdatedAt.Format("2006-01-02"),
		})
	} else {
		urls = append(urls, urlEntry{Loc: fmt.Sprintf("%s/haoka", base)})
	}
	for _, p := range haokaProducts {
		urls = append(urls, urlEntry{
			Loc:     fmt.Sprintf("%s/haoka/%d", base, p.ID),
			LastMod: p.UpdatedAt.Format("2006-01-02"),
		})
	}

	out := urlset{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Status(http.StatusOK)
	enc := xml.NewEncoder(c.Writer)
	enc.Indent("", "  ")
	_ = enc.Encode(out)
}

func buildSiteBaseURL(c *gin.Context) string {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin != "" {
		return strings.TrimRight(origin, "/")
	}
	host := strings.TrimSpace(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}
	proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if proto == "" {
		// 兜底：本地通常是 http
		proto = "http"
	}
	if host == "" {
		return "http://localhost:3007"
	}
	return strings.TrimRight(fmt.Sprintf("%s://%s", proto, host), "/")
}

func sanitizeXMLText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// 先 HTML escape，避免 RSS 里出现非法字符
	return html.EscapeString(s)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

