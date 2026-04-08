package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/config"
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"

	"github.com/meilisearch/meilisearch-go"
)

type meiliState struct {
	enabled bool
	cfg     config.MeiliConfig
	client  meilisearch.ServiceManager
	index   meilisearch.IndexManager
	gameIdx meilisearch.IndexManager
}

var meili meiliState

type meiliResourceDoc struct {
	ID uint64 `json:"id"`

	Title       string `json:"title"`
	Description string `json:"description"`
	Tags        string `json:"tags"`

	// 网盘资源关键字段：也导入到索引，便于搜索/返回
	Cover       string   `json:"cover"`
	Link        string   `json:"link"`
	ExtraLinks  []string `json:"extra_links"`
	ExtractCode string   `json:"extract_code"`

	CategoryID  uint64 `json:"category_id"`
	Platform    string `json:"platform"`
	Source      string `json:"source"`
	LinkValid   bool   `json:"link_valid"`
	Status      int8   `json:"status"`
	ViewCount   uint64 `json:"view_count"`
	CreatedAtTS int64  `json:"created_at_ts"`
	UpdatedAtTS int64  `json:"updated_at_ts"`

	TransferStatus string `json:"transfer_status"`
	TransferMsg    string `json:"transfer_msg"`
}

func InitMeili(cfg config.MeiliConfig) error {
	meili = meiliState{enabled: false, cfg: cfg, client: nil, index: nil, gameIdx: nil}
	if !cfg.Enabled {
		return nil
	}
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		return errors.New("meilisearch url is empty")
	}

	hc := &http.Client{Timeout: time.Duration(cfg.TimeoutMS) * time.Millisecond}
	client := meilisearch.New(url,
		meilisearch.WithAPIKey(strings.TrimSpace(cfg.APIKey)),
		meilisearch.WithCustomClient(hc),
	)
	idxName := strings.TrimSpace(cfg.Index)
	if idxName == "" {
		idxName = "resources"
	}
	index := client.Index(idxName)
	// 游戏索引：默认使用 resources 的后缀，避免引入新配置字段
	gameIndexName := strings.TrimSpace(cfg.Index)
	if gameIndexName == "" {
		gameIndexName = "resources"
	}
	gameIndexName = gameIndexName + "_games"
	gameIndex := client.Index(gameIndexName)
	meili.client = client
	meili.index = index
	meili.gameIdx = gameIndex
	meili.enabled = true

	// 尽力初始化索引设置（失败不阻断启动；搜索时会自动回退 MySQL）
	go func() {
		defer func() { recover() }()
		_ = ensureMeiliIndexSettings()
		_ = ensureMeiliGameIndexSettings()
	}()
	return nil
}

func MeiliEnabled() bool {
	return meili.enabled && meili.index != nil
}

func MeiliGameEnabled() bool {
	return meili.enabled && meili.gameIdx != nil
}

func ensureMeiliIndexSettings() error {
	if !MeiliEnabled() {
		return nil
	}
	// 可排序字段（用于 latest/hot）
	_, _ = meili.index.UpdateSortableAttributes(&[]string{"created_at_ts", "view_count"})
	// 可过滤字段（用于 category/platform/link_valid/status）
	_, _ = meili.index.UpdateFilterableAttributes(&[]interface{}{"category_id", "platform", "link_valid", "status"})
	// 可搜索字段
	_, _ = meili.index.UpdateSearchableAttributes(&[]string{
		"title",
		"description",
		"tags",
		"link",
		"extra_links",
		"extract_code",
	})
	return nil
}

type meiliGameDoc struct {
	ID uint64 `json:"id"`

	Title            string `json:"title"`
	ShortDescription string `json:"short_description"`
	Developers       string `json:"developers"`
	Publishers       string `json:"publishers"`
	Platforms        string `json:"platforms"`
	Genres           string `json:"genres"`
	Tags             string `json:"tags"`

	Cover       string `json:"cover"`
	HeaderImage string `json:"header_image"`
	VideoURL    string `json:"video_url"`

	CategoryID  uint64 `json:"category_id"`
	Type        string `json:"type"`
	SteamAppID  uint64 `json:"steam_appid"`
	IsFree      bool   `json:"is_free"`

	Rating          float64 `json:"rating"`
	Downloads       uint64  `json:"downloads"`
	Recommendations uint64  `json:"recommendations_total"`

	CreatedAtTS int64 `json:"created_at_ts"`
	UpdatedAtTS int64 `json:"updated_at_ts"`
}

func toMeiliGameDoc(g model.Game) meiliGameDoc {
	var cid uint64
	if g.CategoryID != nil {
		cid = *g.CategoryID
	}
	return meiliGameDoc{
		ID:               g.ID,
		Title:            strings.TrimSpace(g.Title),
		ShortDescription: strings.TrimSpace(g.ShortDescription),
		Developers:       strings.TrimSpace(g.Developers),
		Publishers:       strings.TrimSpace(g.Publishers),
		Platforms:        strings.TrimSpace(g.Platforms),
		Genres:           strings.TrimSpace(g.Genres),
		Tags:             strings.TrimSpace(g.Tags),
		Cover:            strings.TrimSpace(g.Cover),
		HeaderImage:      strings.TrimSpace(g.HeaderImage),
		VideoURL:         strings.TrimSpace(g.VideoURL),
		CategoryID:       cid,
		Type:             strings.TrimSpace(g.Type),
		SteamAppID:       g.SteamAppID,
		IsFree:           g.IsFree,
		Rating:           g.Rating,
		Downloads:        g.Downloads,
		Recommendations:  g.Recommendations,
		CreatedAtTS:      g.CreatedAt.Unix(),
		UpdatedAtTS:      g.UpdatedAt.Unix(),
	}
}

func ensureMeiliGameIndexSettings() error {
	if !MeiliGameEnabled() {
		return nil
	}
	_, _ = meili.gameIdx.UpdateSortableAttributes(&[]string{"created_at_ts", "downloads", "rating"})
	_, _ = meili.gameIdx.UpdateFilterableAttributes(&[]interface{}{"category_id", "type", "is_free"})
	_, _ = meili.gameIdx.UpdateSearchableAttributes(&[]string{
		"title",
		"short_description",
		"developers",
		"publishers",
		"platforms",
		"genres",
		"tags",
	})
	return nil
}

// EscapeMeiliFilterValue escapes quotes/backslashes in filter string values.
func EscapeMeiliFilterValue(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

func MeiliUpsertGameAsync(id uint64) {
	if !MeiliGameEnabled() || id == 0 {
		return
	}
	go func() {
		defer func() { recover() }()
		var g model.Game
		if err := database.DB().Where("id = ?", id).First(&g).Error; err != nil {
			return
		}
		pk := "id"
		_, _ = meili.gameIdx.AddDocuments([]meiliGameDoc{toMeiliGameDoc(g)}, &meilisearch.DocumentOptions{
			PrimaryKey: &pk,
		})
	}()
}

func MeiliDeleteGameAsync(id uint64) {
	if !MeiliGameEnabled() || id == 0 {
		return
	}
	go func() {
		defer func() { recover() }()
		_, _ = meili.gameIdx.DeleteDocument(strconv.FormatUint(id, 10), nil)
	}()
}

type MeiliGameSearchParams struct {
	Query      string
	Page       int
	PageSize   int
	Sort       string // latest/hot/rating (optional)
	CategoryID string
	Type       string
}

type MeiliGameSearchResult struct {
	List  []model.Game
	Total int64
}

func SearchGamesByMeili(ctx context.Context, p MeiliGameSearchParams) (MeiliGameSearchResult, error) {
	if !MeiliGameEnabled() {
		return MeiliGameSearchResult{}, errors.New("meili disabled")
	}
	page := p.Page
	if page <= 0 {
		page = 1
	}
	pageSize := p.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	filters := make([]string, 0, 3)
	if strings.TrimSpace(p.CategoryID) != "" {
		filters = append(filters, "category_id = "+strings.TrimSpace(p.CategoryID))
	}
	if strings.TrimSpace(p.Type) != "" {
		// string filter needs quotes
		filters = append(filters, "type = \""+EscapeMeiliFilterValue(strings.TrimSpace(p.Type))+"\"")
	}
	filter := strings.Join(filters, " AND ")

	sort := []string(nil)
	switch strings.TrimSpace(p.Sort) {
	case "rating":
		sort = []string{"rating:desc", "created_at_ts:desc"}
	case "hot":
		sort = []string{"downloads:desc", "created_at_ts:desc"}
	default:
		sort = []string{"created_at_ts:desc"}
	}

	q := strings.TrimSpace(p.Query)
	resp, err := meili.gameIdx.Search(q, &meilisearch.SearchRequest{
		Limit:  int64(pageSize),
		Offset: int64(offset),
		Filter: filter,
		Sort:   sort,
	})
	if err != nil {
		return MeiliGameSearchResult{}, err
	}

	// Decode hits into docs, then load from DB by IDs to keep fields consistent
	type hit struct {
		ID uint64 `json:"id"`
	}
	hitsBytes, _ := json.Marshal(resp.Hits)
	var hits []hit
	_ = json.Unmarshal(hitsBytes, &hits)
	ids := make([]uint64, 0, len(hits))
	for _, h := range hits {
		if h.ID > 0 {
			ids = append(ids, h.ID)
		}
	}
	if len(ids) == 0 {
		return MeiliGameSearchResult{List: []model.Game{}, Total: int64(resp.EstimatedTotalHits)}, nil
	}

	// Keep order as hits order
	var rows []model.Game
	if err := database.DB().Model(&model.Game{}).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return MeiliGameSearchResult{}, err
	}
	byID := make(map[uint64]model.Game, len(rows))
	for _, r := range rows {
		byID[r.ID] = r
	}
	out := make([]model.Game, 0, len(ids))
	for _, id := range ids {
		if r, ok := byID[id]; ok {
			out = append(out, r)
		}
	}

	return MeiliGameSearchResult{List: out, Total: int64(resp.EstimatedTotalHits)}, nil
}

func toMeiliResourceDoc(res model.Resource) meiliResourceDoc {
	extra := []string(nil)
	if len(res.ExtraLinks) > 0 {
		extra = append(extra, []string(res.ExtraLinks)...)
	}
	return meiliResourceDoc{
		ID: res.ID,

		Title:       strings.TrimSpace(res.Title),
		Description: strings.TrimSpace(res.Description),
		Tags:        strings.TrimSpace(res.Tags),

		Cover:       strings.TrimSpace(res.Cover),
		Link:        strings.TrimSpace(res.Link),
		ExtraLinks:  extra,
		ExtractCode: strings.TrimSpace(res.ExtractCode),

		CategoryID:  res.CategoryID,
		Platform:    DetectPlatformFromLink(res.Link),
		Source:      strings.TrimSpace(res.Source),
		LinkValid:   res.LinkValid,
		Status:      res.Status,
		ViewCount:   res.ViewCount,
		CreatedAtTS: res.CreatedAt.Unix(),
		UpdatedAtTS: res.UpdatedAt.Unix(),

		TransferStatus: strings.TrimSpace(res.TransferStatus),
		TransferMsg:    strings.TrimSpace(res.TransferMsg),
	}
}

func meiliDocToResource(doc meiliResourceDoc) model.Resource {
	r := model.Resource{
		ID:             doc.ID,
		Title:          doc.Title,
		Link:           doc.Link,
		ExtraLinks:     model.NormalizeExtraShareLinks(doc.ExtraLinks),
		CategoryID:     doc.CategoryID,
		Source:         doc.Source,
		Description:    doc.Description,
		ExtractCode:    doc.ExtractCode,
		Cover:          doc.Cover,
		Tags:           doc.Tags,
		LinkValid:      doc.LinkValid,
		TransferStatus: doc.TransferStatus,
		TransferMsg:    doc.TransferMsg,
		ViewCount:      doc.ViewCount,
		Status:         doc.Status,
	}
	if doc.CreatedAtTS > 0 {
		r.CreatedAt = time.Unix(doc.CreatedAtTS, 0)
	}
	if doc.UpdatedAtTS > 0 {
		r.UpdatedAt = time.Unix(doc.UpdatedAtTS, 0)
	}
	return r
}

func MeiliUpsertResourceAsync(id uint64) {
	if !MeiliEnabled() || id == 0 {
		return
	}
	go func() {
		defer func() { recover() }()
		var res model.Resource
		if err := database.DB().Where("id = ?", id).First(&res).Error; err != nil {
			return
		}
		pk := "id"
		_, _ = meili.index.AddDocuments([]meiliResourceDoc{toMeiliResourceDoc(res)}, &meilisearch.DocumentOptions{
			PrimaryKey: &pk,
		})
	}()
}

func MeiliDeleteResourceAsync(id uint64) {
	if !MeiliEnabled() || id == 0 {
		return
	}
	go func() {
		defer func() { recover() }()
		_, _ = meili.index.DeleteDocument(strconv.FormatUint(id, 10), nil)
	}()
}

type MeiliSearchParams struct {
	Query       string
	Page        int
	PageSize    int
	Sort        string // relevance/latest/hot
	CategoryID  string
	Platform    string
	LinkValid   string
	HideInvalid bool
}

type MeiliSearchResult struct {
	List  []model.Resource
	Total int64
}

func SearchResourcesByMeili(ctx context.Context, p MeiliSearchParams) (MeiliSearchResult, error) {
	if !MeiliEnabled() {
		return MeiliSearchResult{}, errors.New("meili disabled")
	}
	q := strings.TrimSpace(p.Query)
	if q == "" {
		return MeiliSearchResult{List: []model.Resource{}, Total: 0}, nil
	}
	page := p.Page
	if page < 1 {
		page = 1
	}
	pageSize := p.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	filters := make([]string, 0, 6)
	// 前台搜索：只看 status=1
	filters = append(filters, "status = 1")

	if cid := strings.TrimSpace(p.CategoryID); cid != "" {
		if n, err := strconv.ParseUint(cid, 10, 64); err == nil && n > 0 {
			filters = append(filters, "category_id = "+strconv.FormatUint(n, 10))
		}
	}
	if plat := strings.TrimSpace(p.Platform); plat != "" {
		// 与 API 参数保持一致
		filters = append(filters, "platform = \""+strings.ReplaceAll(plat, "\"", "")+"\"")
	}
	// link_valid 的默认隐藏逻辑保持与 MySQL 一致
	if lv := strings.TrimSpace(p.LinkValid); lv != "" {
		switch strings.ToLower(lv) {
		case "1", "true":
			filters = append(filters, "link_valid = true")
		case "0", "false":
			filters = append(filters, "link_valid = false")
		}
	} else if p.HideInvalid {
		filters = append(filters, "link_valid = true")
	}

	filterExpr := strings.Join(filters, " AND ")
	offset := (page - 1) * pageSize

	var sort []string
	switch strings.TrimSpace(p.Sort) {
	case "latest":
		sort = []string{"created_at_ts:desc"}
	case "hot":
		sort = []string{"view_count:desc", "created_at_ts:desc"}
	default:
		// relevance: 不指定 sort，使用 Meili 默认相关度
	}

	res, err := meili.index.Search(q, &meilisearch.SearchRequest{
		Offset: int64(offset),
		Limit:  int64(pageSize),
		Filter: filterExpr,
		Sort:   sort,
		AttributesToRetrieve: []string{
			"id",
			"title",
			"description",
			"tags",
			"cover",
			"link",
			"extra_links",
			"extract_code",
			"category_id",
			"platform",
			"source",
			"link_valid",
			"status",
			"view_count",
			"created_at_ts",
			"updated_at_ts",
			"transfer_status",
			"transfer_msg",
		},
	})
	if err != nil {
		return MeiliSearchResult{}, err
	}

	out := make([]model.Resource, 0, len(res.Hits))
	for _, h := range res.Hits {
		// meilisearch.Hit 的 value 是 json.RawMessage
		b, err := json.Marshal(h)
		if err != nil {
			continue
		}
		var doc meiliResourceDoc
		if err := json.Unmarshal(b, &doc); err != nil {
			continue
		}
		if doc.ID == 0 {
			continue
		}
		out = append(out, meiliDocToResource(doc))
	}

	total := res.EstimatedTotalHits
	if total <= 0 && res.TotalHits > 0 {
		total = res.TotalHits
	}
	if total < 0 {
		total = 0
	}

	return MeiliSearchResult{List: out, Total: total}, nil
}

func MeiliTryLog(err error) {
	if err == nil {
		return
	}
	// 避免刷屏：仅打印关键一行
	log.Printf("[meili] %v", err)
}
