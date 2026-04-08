package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func gameOK(c *gin.Context, data interface{}) {
	response.OK(c, data)
}

func gameErr(c *gin.Context, msg string) {
	response.Error(c, 400, msg)
}

func parseFlexibleDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &t, nil
		}
	}
	return nil, errors.New("invalid date format")
}

func parseGalleryString(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	b, err := json.Marshal(arr)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func decodeGalleryString(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return []string{}
	}
	return arr
}

var splitDownloadSep = regexp.MustCompile(`[\r\n\t ,;，；]+`)

func normalizeDownloadLinks(single string, multi []string) []string {
	parts := make([]string, 0, len(multi)+4)
	if strings.TrimSpace(single) != "" {
		parts = append(parts, splitDownloadSep.Split(single, -1)...)
	}
	for _, item := range multi {
		if strings.TrimSpace(item) == "" {
			continue
		}
		parts = append(parts, splitDownloadSep.Split(item, -1)...)
	}

	seen := make(map[string]struct{}, len(parts))
	links := make([]string, 0, len(parts))
	for _, raw := range parts {
		link := strings.TrimSpace(strings.TrimRight(raw, ".,;，；)）]】"))
		if link == "" {
			continue
		}
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			continue
		}
		if _, err := url.ParseRequestURI(link); err != nil {
			continue
		}
		if _, ok := seen[link]; ok {
			continue
		}
		seen[link] = struct{}{}
		links = append(links, link)
	}
	return links
}

func detectPanTypeByLink(link string) string {
	u := strings.ToLower(strings.TrimSpace(link))
	switch {
	case strings.Contains(u, "pan.baidu.com"):
		return "百度"
	case strings.Contains(u, "pan.quark.cn"):
		return "夸克"
	case strings.Contains(u, "pan.xunlei.com"):
		return "迅雷"
	case strings.Contains(u, "aliyundrive.com"), strings.Contains(u, "alipan.com"):
		return "阿里"
	case strings.Contains(u, "cloud.189.cn"):
		return "天翼"
	case strings.Contains(u, "115.com"):
		return "115"
	case strings.Contains(u, "123684.com"), strings.Contains(u, "123pan.com"):
		return "123"
	case strings.Contains(u, "drive.uc.cn"):
		return "UC"
	default:
		return "其他"
	}
}

func mergePanType(manual string, links []string) string {
	manual = strings.TrimSpace(manual)
	if manual != "" {
		return manual
	}
	if len(links) == 0 {
		return ""
	}
	set := map[string]struct{}{}
	ordered := make([]string, 0, 4)
	for _, link := range links {
		t := detectPanTypeByLink(link)
		if _, ok := set[t]; ok {
			continue
		}
		set[t] = struct{}{}
		ordered = append(ordered, t)
	}
	if len(ordered) == 0 {
		return ""
	}
	return strings.Join(ordered, "/")
}

func normalizeGameResourceType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "mod":
		return "mod"
	case "trainer":
		return "trainer"
	case "submission":
		return "submission"
	default:
		return "game"
	}
}

type gameDTO struct {
	ID               uint64      `json:"id"`
	CategoryID       *uint64     `json:"category_id,omitempty"`
	SteamAppID       uint64      `json:"steam_appid"`
	Title            string      `json:"title"`
	Cover            string      `json:"cover"`
	Banner           string      `json:"banner"`
	VideoURL         string      `json:"video_url"`
	ShortDescription string      `json:"short_description"`
	HeaderImage      string      `json:"header_image"`
	Website          string      `json:"website"`
	Publishers       string      `json:"publishers"`
	Genres           string      `json:"genres"`
	Tags             string      `json:"tags"`
	PriceText        string      `json:"price_text"`
	PriceCurrency    string      `json:"price_currency"`
	PriceInitial     int         `json:"price_initial"`
	PriceFinal       int         `json:"price_final"`
	PriceDiscount    int         `json:"price_discount"`
	MetacriticScore  int         `json:"metacritic_score"`
	Description      string      `json:"description"`
	ReleaseDate      *time.Time  `json:"release_date,omitempty"`
	Size             string      `json:"size"`
	Type             string      `json:"type"`
	Developer        string      `json:"developer"`
	Rating           float64     `json:"rating"`
	SteamScore       int         `json:"steam_score"`
	Downloads        uint64      `json:"downloads"`
	Likes            uint64      `json:"likes"`
	Dislikes         uint64      `json:"dislikes"`
	Gallery          []string    `json:"gallery"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	Resources        interface{} `json:"resources,omitempty"`
}

type gameResourceDTO struct {
	ID           uint64     `json:"id"`
	GameID       uint64     `json:"game_id"`
	Title        string     `json:"title"`
	ResourceType string     `json:"resource_type"`
	Version      string     `json:"version"`
	Size         string     `json:"size"`
	DownloadType string     `json:"download_type"`
	PanType      string     `json:"pan_type"`
	DownloadURL  string     `json:"download_url"`
	DownloadURLs []string   `json:"download_urls"`
	Tested       bool       `json:"tested"`
	Author       string     `json:"author"`
	PublishDate  *time.Time `json:"publish_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func toGameDTO(g model.Game) gameDTO {
	return gameDTO{
		ID:               g.ID,
		CategoryID:       g.CategoryID,
		SteamAppID:       g.SteamAppID,
		Title:            g.Title,
		Cover:            g.Cover,
		Banner:           g.Banner,
		VideoURL:         g.VideoURL,
		ShortDescription: g.ShortDescription,
		HeaderImage:      g.HeaderImage,
		Website:          g.Website,
		Publishers:       g.Publishers,
		Genres:           g.Genres,
		Tags:             g.Tags,
		PriceText:        g.PriceText,
		PriceCurrency:    g.PriceCurrency,
		PriceInitial:     g.PriceInitial,
		PriceFinal:       g.PriceFinal,
		PriceDiscount:    g.PriceDiscount,
		MetacriticScore:  g.MetacriticScore,
		Description:      g.Description,
		ReleaseDate:      g.ReleaseDate,
		Size:             g.Size,
		Type:             g.Type,
		Developer:        g.Developer,
		Rating:           g.Rating,
		SteamScore:       g.SteamScore,
		Downloads:        g.Downloads,
		Likes:            g.Likes,
		Dislikes:         g.Dislikes,
		Gallery:          decodeGalleryString(g.Gallery),
		CreatedAt:        g.CreatedAt,
		UpdatedAt:        g.UpdatedAt,
	}
}

func toGameResourceDTO(item model.GameResource) gameResourceDTO {
	links := normalizeDownloadLinks(item.DownloadURL, nil)
	if len(links) == 0 && strings.TrimSpace(item.DownloadURL) != "" {
		links = []string{strings.TrimSpace(item.DownloadURL)}
	}
	return gameResourceDTO{
		ID:           item.ID,
		GameID:       item.GameID,
		Title:        item.Title,
		ResourceType: item.ResourceType,
		Version:      item.Version,
		Size:         item.Size,
		DownloadType: item.DownloadType,
		PanType:      item.PanType,
		DownloadURL:  item.DownloadURL,
		DownloadURLs: links,
		Tested:       item.Tested,
		Author:       item.Author,
		PublishDate:  item.PublishDate,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

type steamAppDetailsResp struct {
	Success bool `json:"success"`
	Data    struct {
		SteamAppID          uint64   `json:"steam_appid"`
		Name                string   `json:"name"`
		Type                string   `json:"type"`
		IsFree              bool     `json:"is_free"`
		ShortDescription    string   `json:"short_description"`
		DetailedDescription string   `json:"detailed_description"`
		AboutTheGame        string   `json:"about_the_game"`
		HeaderImage         string   `json:"header_image"`
		CapsuleImage        string   `json:"capsule_image"`
		Background          string   `json:"background"`
		BackgroundRaw       string   `json:"background_raw"`
		Website             string   `json:"website"`
		Developers          []string `json:"developers"`
		Publishers          []string `json:"publishers"`
		Categories          []struct {
			ID          int    `json:"id"`
			Description string `json:"description"`
		} `json:"categories"`
		Genres []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"genres"`
		PriceOverview struct {
			Currency         string `json:"currency"`
			Initial          int    `json:"initial"`
			Final            int    `json:"final"`
			DiscountPercent  int    `json:"discount_percent"`
			InitialFormatted string `json:"initial_formatted"`
			FinalFormatted   string `json:"final_formatted"`
		} `json:"price_overview"`
		Screenshots []struct {
			PathFull string `json:"path_full"`
		} `json:"screenshots"`
		Movies []struct {
			MP4      map[string]string `json:"mp4"`
			DashAV1  string            `json:"dash_av1"`
			DashH264 string            `json:"dash_h264"`
			HlsH264  string            `json:"hls_h264"`
		} `json:"movies"`
		ReleaseDate struct {
			ComingSoon bool   `json:"coming_soon"`
			Date       string `json:"date"`
		} `json:"release_date"`
		Metacritic struct {
			Score int    `json:"score"`
			URL   string `json:"url"`
		} `json:"metacritic"`
	} `json:"data"`
}

// ------------------------- Steam 搜索接口 -------------------------

var steamCNWordReg = regexp.MustCompile(`[\p{Han}]+`)
var steamENWordReg = regexp.MustCompile(`[a-zA-Z0-9\s]{3,}`)

type steamStoreSearchResp struct {
	Items []struct {
		ID        uint64 `json:"id"`
		Name      string `json:"name"`
		TinyImage string `json:"tiny_image"`
	} `json:"items"`
}

// GameSteamSearch 基于关键词从 Steam storesearch 检索候选游戏列表（用于后台下拉选择 appid）。
func GameSteamSearch(c *gin.Context) {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		gameErr(c, "name is required")
		return
	}
	cc := strings.TrimSpace(c.DefaultQuery("cc", "CN"))
	lang := strings.TrimSpace(c.DefaultQuery("l", "schinese"))
	if cc == "" {
		cc = "CN"
	}
	if lang == "" {
		lang = "schinese"
	}
	cc = strings.ToUpper(cc)
	hasHan := steamCNWordReg.MatchString(name)

	// 1) 智能拆分中英文：原始、中文段、英文/数字段
	cnParts := steamCNWordReg.FindAllString(name, -1)
	enParts := steamENWordReg.FindAllString(name, -1)
	cnQuery := strings.TrimSpace(strings.Join(cnParts, " "))
	enQuery := strings.TrimSpace(strings.Join(enParts, " "))

	tasks := make([]string, 0, 3)
	addTask := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		for _, t := range tasks {
			if t == s {
				return
			}
		}
		tasks = append(tasks, s)
	}
	addTask(name)
	addTask(cnQuery)
	addTask(enQuery)

	type itemOut struct {
		AppID      uint64 `json:"appid"`
		Name       string `json:"name"`
		Icon       string `json:"icon"`
		MatchScore int    `json:"match_score"`
	}
	combined := map[uint64]*itemOut{}

	client := &http.Client{Timeout: 10 * time.Second}
	for _, term := range tasks {
		endpoint := fmt.Sprintf(
			"https://store.steampowered.com/api/storesearch/?term=%s&l=%s&cc=%s",
			url.QueryEscape(term),
			url.QueryEscape(lang),
			url.QueryEscape(cc),
		)
		resp, err := client.Get(endpoint)
		if err != nil {
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			continue
		}
		var payload steamStoreSearchResp
		if err := json.Unmarshal(raw, &payload); err != nil {
			continue
		}
		for _, it := range payload.Items {
			if it.ID == 0 || strings.TrimSpace(it.Name) == "" {
				continue
			}
			if existing, ok := combined[it.ID]; ok {
				existing.MatchScore++
				continue
			}
			combined[it.ID] = &itemOut{
				AppID:      it.ID,
				Name:       strings.TrimSpace(it.Name),
				Icon:       strings.TrimSpace(it.TinyImage),
				MatchScore: 1,
			}
		}
	}

	list := make([]itemOut, 0, len(combined))
	for _, v := range combined {
		list = append(list, *v)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].MatchScore != list[j].MatchScore {
			return list[i].MatchScore > list[j].MatchScore
		}
		return list[i].AppID < list[j].AppID
	})
	if len(list) > 50 {
		list = list[:50]
	}

	gameOK(c, gin.H{
		"search_term": name,
		"count":       len(list),
		"data":        list,
		"hint": func() string {
			if len(list) > 0 {
				return ""
			}
			// Steam storesearch 对纯中文关键词经常返回空；提示用户改用英文名/手动填写 AppID。
			if hasHan {
				return "Steam storesearch 对中文关键词可能返回空结果。建议输入英文名（如 Star Valley）或直接填写 Steam AppID。"
			}
			return ""
		}(),
	})
}

func GameSteamAppDetail(c *gin.Context) {
	appID := strings.TrimSpace(c.Param("appid"))
	if appID == "" {
		gameErr(c, "invalid appid")
		return
	}
	cc := strings.TrimSpace(c.DefaultQuery("cc", "cn"))
	lang := strings.TrimSpace(c.DefaultQuery("l", "schinese"))
	if cc == "" {
		cc = "cn"
	}
	if lang == "" {
		lang = "schinese"
	}

	reqURL := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%s&cc=%s&l=%s", appID, cc, lang)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		gameErr(c, "fetch steam appdetails failed")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		gameErr(c, "read steam appdetails failed")
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		gameErr(c, "steam api response not ok")
		return
	}

	payload := map[string]steamAppDetailsResp{}
	if err := json.Unmarshal(body, &payload); err != nil {
		gameErr(c, "parse steam appdetails failed")
		return
	}
	raw, ok := payload[appID]
	if !ok || !raw.Success {
		gameErr(c, "steam app not found")
		return
	}

	genres := make([]string, 0, len(raw.Data.Genres))
	for _, g := range raw.Data.Genres {
		if s := strings.TrimSpace(g.Description); s != "" {
			genres = append(genres, s)
		}
	}
	categories := make([]string, 0, len(raw.Data.Categories))
	tags := make([]string, 0, len(raw.Data.Categories)+len(raw.Data.Genres))
	for _, c := range raw.Data.Categories {
		if s := strings.TrimSpace(c.Description); s != "" {
			categories = append(categories, s)
			tags = append(tags, s)
		}
	}
	for _, g := range genres {
		tags = append(tags, g)
	}
	screenshots := make([]string, 0, len(raw.Data.Screenshots))
	for _, s := range raw.Data.Screenshots {
		if strings.TrimSpace(s.PathFull) != "" {
			screenshots = append(screenshots, s.PathFull)
		}
	}

	videoURL := ""
	if len(raw.Data.Movies) > 0 {
		movie := raw.Data.Movies[0]
		candidates := []string{
			movie.HlsH264,
			movie.DashH264,
			movie.DashAV1,
			movie.MP4["max"],
			movie.MP4["480"],
		}
		for _, candidate := range candidates {
			if v := strings.TrimSpace(candidate); v != "" {
				videoURL = v
				break
			}
		}
	}
	priceText := ""
	if raw.Data.IsFree {
		priceText = "免费"
	} else if s := strings.TrimSpace(raw.Data.PriceOverview.FinalFormatted); s != "" {
		priceText = s
	}

	gameOK(c, gin.H{
		"appid":                raw.Data.SteamAppID,
		"name":                 raw.Data.Name,
		"type":                 raw.Data.Type,
		"is_free":              raw.Data.IsFree,
		"short_description":    raw.Data.ShortDescription,
		"detailed_description": raw.Data.DetailedDescription,
		"about_the_game":       raw.Data.AboutTheGame,
		"header_image":         raw.Data.HeaderImage,
		"capsule_image":        raw.Data.CapsuleImage,
		"background":           raw.Data.Background,
		"background_raw":       raw.Data.BackgroundRaw,
		"website":              raw.Data.Website,
		"developers":           raw.Data.Developers,
		"publishers":           raw.Data.Publishers,
		"categories":           categories,
		"genres":               genres,
		"tags":                 tags,
		"screenshots":          screenshots,
		"video_url":            videoURL,
		"price_text":           priceText,
		"price_currency":       raw.Data.PriceOverview.Currency,
		"price_initial":        raw.Data.PriceOverview.Initial,
		"price_final":          raw.Data.PriceOverview.Final,
		"price_discount":       raw.Data.PriceOverview.DiscountPercent,
		"metacritic_score":     raw.Data.Metacritic.Score,
		"metacritic_url":       raw.Data.Metacritic.URL,
		"release_date":         raw.Data.ReleaseDate.Date,
		"coming_soon":          raw.Data.ReleaseDate.ComingSoon,
		"cc":                   cc,
		"l":                    lang,
	})
}

// ------------------------- 分类接口 -------------------------

func GameCategoryCreate(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Slug        string `json:"slug" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gameErr(c, "invalid params")
		return
	}

	item := model.GameCategory{
		Name:        strings.TrimSpace(req.Name),
		Slug:        strings.TrimSpace(req.Slug),
		Description: strings.TrimSpace(req.Description),
	}
	if item.Name == "" || item.Slug == "" {
		gameErr(c, "name and slug are required")
		return
	}

	if err := database.DB().Create(&item).Error; err != nil {
		gameErr(c, "create category failed")
		return
	}
	gameOK(c, item)
}

func GameCategoryDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		gameErr(c, "invalid id")
		return
	}

	if err := database.DB().Delete(&model.GameCategory{}, id).Error; err != nil {
		gameErr(c, "delete category failed")
		return
	}
	gameOK(c, struct{}{})
}

func GameCategoryUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		gameErr(c, "invalid id")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gameErr(c, "invalid params")
		return
	}

	updates := map[string]interface{}{}
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Slug) != "" {
		updates["slug"] = strings.TrimSpace(req.Slug)
	}
	updates["description"] = strings.TrimSpace(req.Description)

	if len(updates) == 0 {
		gameErr(c, "no fields to update")
		return
	}

	if err := database.DB().Model(&model.GameCategory{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		gameErr(c, "update category failed")
		return
	}
	gameOK(c, struct{}{})
}

func GameCategoryList(c *gin.Context) {
	var list []model.GameCategory
	if err := database.DB().Order("id DESC").Find(&list).Error; err != nil {
		gameErr(c, "query category list failed")
		return
	}
	gameOK(c, list)
}

// ------------------------- 游戏接口 -------------------------

func GameCreate(c *gin.Context) {
	var req struct {
		CategoryID       *uint64  `json:"category_id"`
		SteamAppID       uint64   `json:"steam_appid"`
		Title            string   `json:"title" binding:"required"`
		Cover            string   `json:"cover"`
		Banner           string   `json:"banner"`
		VideoURL         string   `json:"video_url"`
		ShortDescription string   `json:"short_description"`
		HeaderImage      string   `json:"header_image"`
		Website          string   `json:"website"`
		Publishers       string   `json:"publishers"`
		Genres           string   `json:"genres"`
		Tags             string   `json:"tags"`
		PriceText        string   `json:"price_text"`
		PriceCurrency    string   `json:"price_currency"`
		PriceInitial     int      `json:"price_initial"`
		PriceFinal       int      `json:"price_final"`
		PriceDiscount    int      `json:"price_discount"`
		MetacriticScore  int      `json:"metacritic_score"`
		Description      string   `json:"description"`
		ReleaseDate      string   `json:"release_date"`
		Size             string   `json:"size"`
		Type             string   `json:"type"`
		Developer        string   `json:"developer"`
		Rating           float64  `json:"rating"`
		SteamScore       int      `json:"steam_score"`
		Downloads        uint64   `json:"downloads"`
		Likes            uint64   `json:"likes"`
		Dislikes         uint64   `json:"dislikes"`
		Gallery          []string `json:"gallery"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gameErr(c, "invalid params")
		return
	}

	releaseDate, err := parseFlexibleDate(req.ReleaseDate)
	if err != nil {
		gameErr(c, "invalid release_date")
		return
	}

	item := model.Game{
		CategoryID:       req.CategoryID,
		SteamAppID:       req.SteamAppID,
		Title:            strings.TrimSpace(req.Title),
		Cover:            strings.TrimSpace(req.Cover),
		Banner:           strings.TrimSpace(req.Banner),
		VideoURL:         strings.TrimSpace(req.VideoURL),
		ShortDescription: strings.TrimSpace(req.ShortDescription),
		HeaderImage:      strings.TrimSpace(req.HeaderImage),
		Website:          strings.TrimSpace(req.Website),
		Publishers:       strings.TrimSpace(req.Publishers),
		Genres:           strings.TrimSpace(req.Genres),
		Tags:             strings.TrimSpace(req.Tags),
		PriceText:        strings.TrimSpace(req.PriceText),
		PriceCurrency:    strings.TrimSpace(req.PriceCurrency),
		PriceInitial:     req.PriceInitial,
		PriceFinal:       req.PriceFinal,
		PriceDiscount:    req.PriceDiscount,
		MetacriticScore:  req.MetacriticScore,
		Description:      req.Description,
		ReleaseDate:      releaseDate,
		Size:             strings.TrimSpace(req.Size),
		Type:             strings.TrimSpace(req.Type),
		Developer:        strings.TrimSpace(req.Developer),
		Rating:           req.Rating,
		SteamScore:       req.SteamScore,
		Downloads:        req.Downloads,
		Likes:            req.Likes,
		Dislikes:         req.Dislikes,
		Gallery:          parseGalleryString(req.Gallery),
	}
	if item.Title == "" {
		gameErr(c, "title is required")
		return
	}

	if err := database.DB().Create(&item).Error; err != nil {
		gameErr(c, "create game failed")
		return
	}
	gameOK(c, toGameDTO(item))
}

func GameUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		gameErr(c, "invalid id")
		return
	}

	var req struct {
		CategoryID       *uint64  `json:"category_id"`
		SteamAppID       *uint64  `json:"steam_appid"`
		Title            string   `json:"title"`
		Cover            string   `json:"cover"`
		Banner           string   `json:"banner"`
		VideoURL         string   `json:"video_url"`
		ShortDescription string   `json:"short_description"`
		HeaderImage      string   `json:"header_image"`
		Website          string   `json:"website"`
		Publishers       string   `json:"publishers"`
		Genres           string   `json:"genres"`
		Tags             string   `json:"tags"`
		PriceText        string   `json:"price_text"`
		PriceCurrency    string   `json:"price_currency"`
		PriceInitial     *int     `json:"price_initial"`
		PriceFinal       *int     `json:"price_final"`
		PriceDiscount    *int     `json:"price_discount"`
		MetacriticScore  *int     `json:"metacritic_score"`
		Description      string   `json:"description"`
		ReleaseDate      string   `json:"release_date"`
		Size             string   `json:"size"`
		Type             string   `json:"type"`
		Developer        string   `json:"developer"`
		Rating           *float64 `json:"rating"`
		SteamScore       *int     `json:"steam_score"`
		Downloads        *uint64  `json:"downloads"`
		Likes            *uint64  `json:"likes"`
		Dislikes         *uint64  `json:"dislikes"`
		Gallery          []string `json:"gallery"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gameErr(c, "invalid params")
		return
	}

	updates := map[string]interface{}{}
	if req.CategoryID != nil {
		updates["category_id"] = req.CategoryID
	}
	if req.SteamAppID != nil {
		updates["steam_app_id"] = *req.SteamAppID
	}
	if strings.TrimSpace(req.Title) != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.Cover != "" {
		updates["cover"] = strings.TrimSpace(req.Cover)
	}
	if req.Banner != "" {
		updates["banner"] = strings.TrimSpace(req.Banner)
	}
	if req.VideoURL != "" {
		updates["video_url"] = strings.TrimSpace(req.VideoURL)
	}
	if req.ShortDescription != "" {
		updates["short_description"] = strings.TrimSpace(req.ShortDescription)
	}
	if req.HeaderImage != "" {
		updates["header_image"] = strings.TrimSpace(req.HeaderImage)
	}
	if req.Website != "" {
		updates["website"] = strings.TrimSpace(req.Website)
	}
	if req.Publishers != "" {
		updates["publishers"] = strings.TrimSpace(req.Publishers)
	}
	if req.Genres != "" {
		updates["genres"] = strings.TrimSpace(req.Genres)
	}
	if req.Tags != "" {
		updates["tags"] = strings.TrimSpace(req.Tags)
	}
	if req.PriceText != "" {
		updates["price_text"] = strings.TrimSpace(req.PriceText)
	}
	if req.PriceCurrency != "" {
		updates["price_currency"] = strings.TrimSpace(req.PriceCurrency)
	}
	if req.PriceInitial != nil {
		updates["price_initial"] = *req.PriceInitial
	}
	if req.PriceFinal != nil {
		updates["price_final"] = *req.PriceFinal
	}
	if req.PriceDiscount != nil {
		updates["price_discount"] = *req.PriceDiscount
	}
	if req.MetacriticScore != nil {
		updates["metacritic_score"] = *req.MetacriticScore
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ReleaseDate != "" {
		releaseDate, err := parseFlexibleDate(req.ReleaseDate)
		if err != nil {
			gameErr(c, "invalid release_date")
			return
		}
		updates["release_date"] = releaseDate
	}
	if req.Size != "" {
		updates["size"] = strings.TrimSpace(req.Size)
	}
	if req.Type != "" {
		updates["type"] = strings.TrimSpace(req.Type)
	}
	if req.Developer != "" {
		updates["developer"] = strings.TrimSpace(req.Developer)
	}
	if req.Rating != nil {
		updates["rating"] = *req.Rating
	}
	if req.SteamScore != nil {
		updates["steam_score"] = *req.SteamScore
	}
	if req.Downloads != nil {
		updates["downloads"] = *req.Downloads
	}
	if req.Likes != nil {
		updates["likes"] = *req.Likes
	}
	if req.Dislikes != nil {
		updates["dislikes"] = *req.Dislikes
	}
	if req.Gallery != nil {
		updates["gallery"] = parseGalleryString(req.Gallery)
	}

	if len(updates) == 0 {
		gameErr(c, "no fields to update")
		return
	}

	if err := database.DB().Model(&model.Game{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		gameErr(c, "update game failed")
		return
	}
	gameOK(c, struct{}{})
}

func GameDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		gameErr(c, "invalid id")
		return
	}

	tx := database.DB().Begin()
	if tx.Error != nil {
		gameErr(c, "delete game failed")
		return
	}

	if err := tx.Where("game_id = ?", id).Delete(&model.GameResource{}).Error; err != nil {
		tx.Rollback()
		gameErr(c, "delete game resources failed")
		return
	}
	if err := tx.Delete(&model.Game{}, id).Error; err != nil {
		tx.Rollback()
		gameErr(c, "delete game failed")
		return
	}
	if err := tx.Commit().Error; err != nil {
		gameErr(c, "delete game failed")
		return
	}
	gameOK(c, struct{}{})
}

func GameList(c *gin.Context) {
	var (
		db   = database.DB().Model(&model.Game{})
		list []model.Game
	)

	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		db = db.Where("title LIKE ?", "%"+keyword+"%")
	}
	if cid := strings.TrimSpace(c.Query("category_id")); cid != "" {
		db = db.Where("category_id = ?", cid)
	}
	if typ := strings.TrimSpace(c.Query("type")); typ != "" {
		db = db.Where("type = ?", typ)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		gameErr(c, "query game list failed")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	if err := db.Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		gameErr(c, "query game list failed")
		return
	}

	data := make([]gameDTO, 0, len(list))
	for _, item := range list {
		data = append(data, toGameDTO(item))
	}
	gameOK(c, gin.H{
		"list":  data,
		"total": total,
	})
}

func GameDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		gameErr(c, "invalid id")
		return
	}

	var item model.Game
	if err := database.DB().First(&item, id).Error; err != nil {
		gameErr(c, "game not found")
		return
	}

	var resources []model.GameResource
	_ = database.DB().Where("game_id = ?", id).Order("id DESC").Find(&resources).Error

	dto := toGameDTO(item)
	resourceDTOs := make([]gameResourceDTO, 0, len(resources))
	for _, res := range resources {
		resourceDTOs = append(resourceDTOs, toGameResourceDTO(res))
	}
	dto.Resources = resourceDTOs
	gameOK(c, dto)
}

// ------------------------- 游戏资源接口 -------------------------

func GameResourceCreate(c *gin.Context) {
	var req struct {
		GameID       uint64   `json:"game_id" binding:"required"`
		Title        string   `json:"title" binding:"required"`
		ResourceType string   `json:"resource_type"`
		Version      string   `json:"version"`
		Size         string   `json:"size"`
		DownloadType string   `json:"download_type"`
		PanType      string   `json:"pan_type"`
		DownloadURL  string   `json:"download_url" binding:"required"`
		DownloadURLs []string `json:"download_urls"`
		Tested       bool     `json:"tested"`
		Author       string   `json:"author"`
		PublishDate  string   `json:"publish_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gameErr(c, "invalid params")
		return
	}

	var gameCount int64
	if err := database.DB().Model(&model.Game{}).Where("id = ?", req.GameID).Count(&gameCount).Error; err != nil || gameCount == 0 {
		gameErr(c, "game not found")
		return
	}

	publishDate, err := parseFlexibleDate(req.PublishDate)
	if err != nil {
		gameErr(c, "invalid publish_date")
		return
	}

	links := normalizeDownloadLinks(req.DownloadURL, req.DownloadURLs)
	if len(links) == 0 {
		gameErr(c, "at least one valid download url is required")
		return
	}
	mergedURL := strings.Join(links, "\n")

	item := model.GameResource{
		GameID:       req.GameID,
		Title:        strings.TrimSpace(req.Title),
		ResourceType: normalizeGameResourceType(req.ResourceType),
		Version:      strings.TrimSpace(req.Version),
		Size:         strings.TrimSpace(req.Size),
		DownloadType: strings.TrimSpace(req.DownloadType),
		PanType:      mergePanType(req.PanType, links),
		DownloadURL:  mergedURL,
		Tested:       req.Tested,
		Author:       strings.TrimSpace(req.Author),
		PublishDate:  publishDate,
	}
	if err := database.DB().Create(&item).Error; err != nil {
		gameErr(c, "create resource failed")
		return
	}
	gameOK(c, item)
}

func GameResourceUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		gameErr(c, "invalid id")
		return
	}

	var req struct {
		ResourceType string   `json:"resource_type"`
		Title        string   `json:"title"`
		Version      string   `json:"version"`
		Size         string   `json:"size"`
		DownloadType string   `json:"download_type"`
		PanType      string   `json:"pan_type"`
		DownloadURL  string   `json:"download_url"`
		DownloadURLs []string `json:"download_urls"`
		Tested       *bool    `json:"tested"`
		Author       string   `json:"author"`
		PublishDate  string   `json:"publish_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		gameErr(c, "invalid params")
		return
	}

	updates := map[string]interface{}{}
	if req.ResourceType != "" {
		updates["resource_type"] = normalizeGameResourceType(req.ResourceType)
	}
	if req.Title != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.Version != "" {
		updates["version"] = strings.TrimSpace(req.Version)
	}
	if req.Size != "" {
		updates["size"] = strings.TrimSpace(req.Size)
	}
	if req.DownloadType != "" {
		updates["download_type"] = strings.TrimSpace(req.DownloadType)
	}
	if req.PanType != "" {
		updates["pan_type"] = strings.TrimSpace(req.PanType)
	}
	if req.DownloadURL != "" || len(req.DownloadURLs) > 0 {
		links := normalizeDownloadLinks(req.DownloadURL, req.DownloadURLs)
		if len(links) == 0 {
			gameErr(c, "at least one valid download url is required")
			return
		}
		updates["download_url"] = strings.Join(links, "\n")
		if strings.TrimSpace(req.PanType) == "" {
			updates["pan_type"] = mergePanType("", links)
		}
	}
	if req.Tested != nil {
		updates["tested"] = *req.Tested
	}
	if req.Author != "" {
		updates["author"] = strings.TrimSpace(req.Author)
	}
	if req.PublishDate != "" {
		publishDate, err := parseFlexibleDate(req.PublishDate)
		if err != nil {
			gameErr(c, "invalid publish_date")
			return
		}
		updates["publish_date"] = publishDate
	}

	if len(updates) == 0 {
		gameErr(c, "no fields to update")
		return
	}

	if err := database.DB().Model(&model.GameResource{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		gameErr(c, "update resource failed")
		return
	}
	gameOK(c, struct{}{})
}

func GameResourceDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		gameErr(c, "invalid id")
		return
	}

	if err := database.DB().Delete(&model.GameResource{}, id).Error; err != nil {
		gameErr(c, "delete resource failed")
		return
	}
	gameOK(c, struct{}{})
}

func GameResourceList(c *gin.Context) {
	gameID, _ := strconv.ParseUint(c.Query("game_id"), 10, 64)
	if gameID == 0 {
		gameErr(c, "game_id is required")
		return
	}

	var list []model.GameResource
	if err := database.DB().Where("game_id = ?", gameID).Order("id DESC").Find(&list).Error; err != nil {
		gameErr(c, "query resource list failed")
		return
	}
	result := make([]gameResourceDTO, 0, len(list))
	for _, item := range list {
		result = append(result, toGameResourceDTO(item))
	}
	gameOK(c, result)
}

// ------------------------- 上传接口 -------------------------

// GameUpload 上传封面/截图
// form-data:
// - file: 文件
// - dir: game-covers / game-gallery (可选)
func GameUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		gameErr(c, "file is required")
		return
	}

	dirName := strings.TrimSpace(c.DefaultPostForm("dir", "game-gallery"))
	if dirName == "" {
		dirName = "game-gallery"
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
	}
	if !allowed[ext] {
		gameErr(c, "only jpg/jpeg/png/webp/gif are supported")
		return
	}

	saveDir := filepath.Join("storage", "covers", dirName)
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		gameErr(c, "create upload dir failed")
		return
	}

	fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), os.Getpid(), ext)
	savePath := filepath.Join(saveDir, fileName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		gameErr(c, "save file failed")
		return
	}

	url := "/public/covers/" + dirName + "/" + fileName
	gameOK(c, gin.H{
		"url": url,
	})
}
