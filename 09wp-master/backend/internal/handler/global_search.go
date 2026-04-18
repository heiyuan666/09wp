package handler

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"sync"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/config"
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type globalSearchItem struct {
	URL       string   `json:"url"`
	Password  string   `json:"password"`
	Note      string   `json:"note"`
	Datetime  string   `json:"datetime"`
	Source    string   `json:"source"`
	Images    []string `json:"images"`
	CloudType string   `json:"cloud_type"`
	LinkStatus string  `json:"link_status,omitempty"` // valid / invalid / pending / unknown
}

type globalLinkStatusCache struct {
	Status    string `json:"status"`
	Msg       string `json:"msg,omitempty"`
	CheckedAt string `json:"checked_at,omitempty"`
}

var globalSearchCheckSem = make(chan struct{}, 2)

type globalCheckTask struct {
	Links             []string
	SelectedPlatforms []string
	BaseURL           string
}

var (
	globalCheckQueue     = make(chan globalCheckTask, 256)
	globalCheckWorkerOnce sync.Once
)

func normalizeGlobalURL(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil {
		return strings.ToLower(strings.TrimRight(s, "/"))
	}
	u.Fragment = ""
	u.Host = strings.ToLower(strings.TrimSpace(u.Host))
	u.Path = strings.TrimRight(u.Path, "/")
	q := u.Query()
	// 去掉常见无关追踪参数，减少同资源多 URL 抖动
	for _, k := range []string{"utm_source", "utm_medium", "utm_campaign", "from", "src", "spm"} {
		q.Del(k)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func globalDedupeKey(it globalSearchItem) string {
	return normalizeGlobalURL(it.URL) + "|" + strings.ToLower(strings.TrimSpace(it.CloudType))
}

func parseSearchDatetime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "0001-01-01") {
		return time.Time{}
	}
	// 常见格式优先
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func cloudTypeOrder(cloud string) int {
	switch strings.ToLower(strings.TrimSpace(cloud)) {
	case "quark":
		return 1
	case "baidu":
		return 2
	case "uc":
		return 3
	case "xunlei":
		return 4
	case "aliyun":
		return 5
	case "tianyi":
		return 6
	case "magnet", "bt", "torrent":
		return 7
	default:
		return 99
	}
}

// effectiveGlobalItemCloudType 优先用上游 cloud_type，空或 unknown 时按链接推断（与转存识别一致）。
func effectiveGlobalItemCloudType(it globalSearchItem) string {
	ct := strings.ToLower(strings.TrimSpace(it.CloudType))
	if ct != "" && ct != "unknown" && ct != "<nil>" {
		return ct
	}
	return strings.ToLower(service.DetectTransferPlatform(it.URL).String())
}

// filterGlobalSearchByCloudTypes 当前端/配置指定了 cloud_types 时，对聚合结果再收敛（兼容上游键名不一致，如 magnet/bt/torrent）。
func filterGlobalSearchByCloudTypes(items []globalSearchItem, cloudTypes string) []globalSearchItem {
	parts := parseSelectedPlatforms(cloudTypes)
	if len(parts) == 0 || len(items) == 0 {
		return items
	}
	want := make(map[string]struct{}, len(parts))
	wantMagnet := false
	for _, p := range parts {
		want[p] = struct{}{}
		if p == "magnet" {
			wantMagnet = true
		}
	}
	out := make([]globalSearchItem, 0, len(items))
	for _, it := range items {
		ct := effectiveGlobalItemCloudType(it)
		if _, ok := want[ct]; ok {
			out = append(out, it)
			continue
		}
		if wantMagnet {
			if ct == "magnet" || ct == "bt" || ct == "torrent" {
				out = append(out, it)
				continue
			}
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(it.URL)), "magnet:") {
				out = append(out, it)
			}
		}
	}
	return out
}

// filterKnownInvalidGlobalItems 仅过滤“已知失效”的链接，避免实时外部探测影响搜索速度。
// 规则：命中本站资源库中 link_valid=false 或 status=0 的同链接，直接从全网搜结果隐藏。
func filterKnownInvalidGlobalItems(items []globalSearchItem) []globalSearchItem {
	if len(items) == 0 {
		return items
	}
	rawToNorm := make(map[string]string, len(items))
	rawURLs := make([]string, 0, len(items))
	for _, it := range items {
		u := strings.TrimSpace(it.URL)
		if u == "" {
			continue
		}
		if _, ok := rawToNorm[u]; ok {
			continue
		}
		rawToNorm[u] = normalizeGlobalURL(u)
		rawURLs = append(rawURLs, u)
	}
	if len(rawURLs) == 0 {
		return items
	}
	var rows []model.Resource
	_ = database.DB().
		Select("link", "link_valid", "status").
		Where("link IN ?", rawURLs).
		Where("link_valid = ? OR status = ?", false, 0).
		Find(&rows).Error

	invalidNormSet := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		u := normalizeGlobalURL(strings.TrimSpace(r.Link))
		if u != "" {
			invalidNormSet[u] = struct{}{}
		}
	}
	if len(invalidNormSet) == 0 {
		return items
	}
	out := make([]globalSearchItem, 0, len(items))
	for _, it := range items {
		if _, bad := invalidNormSet[normalizeGlobalURL(it.URL)]; bad {
			continue
		}
		out = append(out, it)
	}
	return out
}

func betterGlobalItem(a, b globalSearchItem) globalSearchItem {
	// 规则：时间新的优先；备注更长优先；图片更多优先；有提取码优先
	at := parseSearchDatetime(a.Datetime)
	bt := parseSearchDatetime(b.Datetime)
	if !at.Equal(bt) {
		if bt.After(at) {
			return b
		}
		return a
	}
	if len(strings.TrimSpace(b.Note)) > len(strings.TrimSpace(a.Note)) {
		return b
	}
	if len(b.Images) > len(a.Images) {
		return b
	}
	if strings.TrimSpace(a.Password) == "" && strings.TrimSpace(b.Password) != "" {
		return b
	}
	return a
}

func readGlobalSearchConfig() model.SystemConfig {
	var cfg model.SystemConfig
	_ = database.DB().Order("id ASC").First(&cfg).Error
	return cfg
}

// reBuiltinExtractShareURL 从上游混有文案的字段中提取首段 http(s)/magnet 链接。
var reBuiltinExtractShareURL = regexp.MustCompile(`(?i)(https?://[^\s，。！；、]+|magnet:\?[^\s，。！；、]+)`)

func sanitizeGlobalSearchURLString(raw string, customRegex string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if customRegex != "" {
		if re, err := regexp.Compile(customRegex); err == nil {
			if s := strings.TrimSpace(re.FindString(raw)); s != "" {
				return s
			}
		}
	}
	if s := strings.TrimSpace(reBuiltinExtractShareURL.FindString(raw)); s != "" {
		return s
	}
	return raw
}

func globalSearchAPIsFingerprint(apis []model.GlobalSearchAPI, sanitizeRegex string) string {
	var b strings.Builder
	for _, a := range apis {
		b.WriteString(strings.TrimSpace(a.APIURL))
		b.WriteByte('|')
		b.WriteString(strings.TrimSpace(a.CloudTypes))
		b.WriteByte('|')
		b.WriteString(fmt.Sprintf("%d", a.ID))
		b.WriteByte(';')
	}
	b.WriteString(strings.TrimSpace(sanitizeRegex))
	sum := md5.Sum([]byte(b.String()))
	return fmt.Sprintf("%x", sum)
}

func globalSearchResultCacheKey(q, cloudTypes string, apis []model.GlobalSearchAPI, sanitizeRegex string) string {
	fp := globalSearchAPIsFingerprint(apis, sanitizeRegex)
	raw := strings.TrimSpace(strings.ToLower(q)) + "|" + strings.TrimSpace(strings.ToLower(cloudTypes)) + "|" + fp
	sum := md5.Sum([]byte(raw))
	return fmt.Sprintf("global:search:result:v3:%x", sum)
}

// mergedByTypeFromRoot 兼容 data.merged_by_type、根级 merged_by_type 等多种上游结构。
func mergedByTypeFromRoot(root map[string]any) map[string]any {
	if root == nil {
		return nil
	}
	if m, ok := root["merged_by_type"].(map[string]any); ok && len(m) > 0 {
		return m
	}
	if data, ok := root["data"].(map[string]any); ok && data != nil {
		if m, ok := data["merged_by_type"].(map[string]any); ok && len(m) > 0 {
			return m
		}
	}
	return nil
}

func ingestGlobalSearchRow(m map[string]any, cloudType string, byKey map[string]globalSearchItem, sanitizePat string) {
	if m == nil {
		return
	}
	ct := strings.ToLower(strings.TrimSpace(cloudType))
	urlRaw := strings.TrimSpace(fmt.Sprintf("%v", m["url"]))
	if urlRaw == "" {
		urlRaw = strings.TrimSpace(fmt.Sprintf("%v", m["link"]))
	}
	urlRaw = sanitizeGlobalSearchURLString(urlRaw, sanitizePat)
	if urlRaw == "" || strings.HasPrefix(strings.ToLower(urlRaw), "<nil>") {
		return
	}
	it := globalSearchItem{
		URL:       urlRaw,
		Password:  strings.TrimSpace(fmt.Sprintf("%v", m["password"])),
		Note:      strings.TrimSpace(fmt.Sprintf("%v", m["note"])),
		Datetime:  strings.TrimSpace(fmt.Sprintf("%v", m["datetime"])),
		Source:    strings.TrimSpace(fmt.Sprintf("%v", m["source"])),
		CloudType: ct,
	}
	if it.CloudType == "" || it.CloudType == "<nil>" {
		it.CloudType = "unknown"
	}
	it.URL = normalizeGlobalURL(it.URL)
	key := globalDedupeKey(it)
	if rawImages, ok := m["images"].([]any); ok {
		for _, x := range rawImages {
			v := strings.TrimSpace(fmt.Sprintf("%v", x))
			if v != "" && !strings.HasPrefix(strings.ToLower(v), "<nil>") {
				it.Images = append(it.Images, v)
			}
		}
	}
	if prev, ok := byKey[key]; ok {
		byKey[key] = betterGlobalItem(prev, it)
	} else {
		byKey[key] = it
	}
}

func appendGlobalSearchFromFlatList(arr []any, defaultCloud string, byKey map[string]globalSearchItem, sanitizePat string) {
	for _, row := range arr {
		m, _ := row.(map[string]any)
		if m == nil {
			continue
		}
		ct := strings.TrimSpace(defaultCloud)
		for _, key := range []string{"cloud_type", "cloudType", "type", "platform", "disk"} {
			v := strings.TrimSpace(fmt.Sprintf("%v", m[key]))
			if v != "" && !strings.HasPrefix(strings.ToLower(v), "<nil>") {
				ct = v
				break
			}
		}
		ingestGlobalSearchRow(m, ct, byKey, sanitizePat)
	}
}

func appendGlobalSearchFromDataUnknownShape(data any, byKey map[string]globalSearchItem, sanitizePat string) {
	if data == nil {
		return
	}
	switch v := data.(type) {
	case []any:
		appendGlobalSearchFromFlatList(v, "", byKey, sanitizePat)
	case map[string]any:
		if merged, ok := v["merged_by_type"].(map[string]any); ok && len(merged) > 0 {
			for cloudType, rawList := range merged {
				arr, _ := rawList.([]any)
				for _, row := range arr {
					m, _ := row.(map[string]any)
					ingestGlobalSearchRow(m, cloudType, byKey, sanitizePat)
				}
			}
		}
		if list, ok := v["list"].([]any); ok && len(list) > 0 {
			appendGlobalSearchFromFlatList(list, "", byKey, sanitizePat)
		}
	}
}

func globalLinkStatusKey(link string) string {
	raw := strings.TrimSpace(strings.ToLower(normalizeGlobalURL(link)))
	sum := md5.Sum([]byte(raw))
	return fmt.Sprintf("global:link:status:v1:%x", sum)
}

func getGlobalLinkStatus(link string) (globalLinkStatusCache, bool) {
	key := globalLinkStatusKey(link)
	if b, ok := service.GetSearchCache(context.Background(), key); ok && len(b) > 0 {
		var out globalLinkStatusCache
		if err := json.Unmarshal(b, &out); err == nil && strings.TrimSpace(out.Status) != "" {
			return out, true
		}
	}
	return globalLinkStatusCache{}, false
}

func setGlobalLinkStatus(link string, s globalLinkStatusCache, ttl time.Duration) {
	key := globalLinkStatusKey(link)
	raw, err := json.Marshal(s)
	if err != nil {
		return
	}
	service.SetSearchCacheWithTTL(context.Background(), key, raw, ttl)
}

func ensureGlobalCheckWorker() {
	globalCheckWorkerOnce.Do(func() {
		go runGlobalCheckWorker()
	})
}

func runGlobalCheckWorker() {
	ticker := time.NewTicker(1200 * time.Millisecond)
	defer ticker.Stop()

	pendingLinks := make(map[string]struct{}, 128)
	selectedPlatforms := []string{}
	baseURL := ""
	flush := func() {
		if len(pendingLinks) == 0 || strings.TrimSpace(baseURL) == "" {
			return
		}
		links := make([]string, 0, len(pendingLinks))
		for link := range pendingLinks {
			links = append(links, link)
		}
		// 重置缓冲，避免阻塞后续入队
		pendingLinks = make(map[string]struct{}, 128)

		resp, err := service.PanCheckLinksWithPolling(service.PanCheckRequest{
			Links:             links,
			SelectedPlatforms: selectedPlatforms,
		}, baseURL, 1, 2*time.Second)
		now := time.Now().Format(time.RFC3339)
		if err != nil {
			for _, v := range links {
				setGlobalLinkStatus(v, globalLinkStatusCache{Status: "unknown", Msg: "检测服务暂不可用", CheckedAt: now}, 2*time.Minute)
			}
			return
		}
		for _, v := range resp.ValidLinks {
			setGlobalLinkStatus(v, globalLinkStatusCache{Status: "valid", Msg: "有效", CheckedAt: now}, 30*time.Minute)
		}
		for _, v := range resp.InvalidLinks {
			setGlobalLinkStatus(v, globalLinkStatusCache{Status: "invalid", Msg: "失效", CheckedAt: now}, 2*time.Hour)
		}
		for _, v := range resp.PendingLinks {
			setGlobalLinkStatus(v, globalLinkStatusCache{Status: "pending", Msg: "待检测", CheckedAt: now}, 30*time.Second)
		}
	}

	for {
		select {
		case task := <-globalCheckQueue:
			if strings.TrimSpace(task.BaseURL) != "" {
				baseURL = strings.TrimSpace(task.BaseURL)
			}
			if len(task.SelectedPlatforms) > 0 {
				selectedPlatforms = task.SelectedPlatforms
			}
			for _, link := range task.Links {
				link = strings.TrimSpace(link)
				if link == "" {
					continue
				}
				pendingLinks[link] = struct{}{}
				if len(pendingLinks) >= 60 {
					flush()
				}
			}
		case <-ticker.C:
			flush()
		}
	}
}

func enqueueGlobalLinkChecks(links []string, cfg model.SystemConfig, selectedPlatforms []string) {
	baseURL := strings.TrimSpace(cfg.PanCheckBaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(config.DefaultPanCheckBaseURL)
	}
	if baseURL == "" || len(links) == 0 {
		return
	}
	ensureGlobalCheckWorker()
	select {
	case globalCheckQueue <- globalCheckTask{
		Links:             links,
		SelectedPlatforms: selectedPlatforms,
		BaseURL:           baseURL,
	}:
	default:
		// 队列繁忙时不阻塞搜索请求
	}
}

func globalSearchAPIURL(cfg model.SystemConfig) string {
	u := strings.TrimSpace(cfg.GlobalSearchAPIURL)
	if u == "" {
		u = strings.TrimSpace(cfg.IYunsAPIBaseURL)
		if u == "" {
			u = "https://api.iyuns.com"
		}
		u = strings.TrimRight(u, "/") + "/api/wpysso"
	}
	return u
}

func fetchGlobalSearchWithAPIs(kw, passedSessionCT string, apis []model.GlobalSearchAPI, cfg model.SystemConfig) ([]globalSearchItem, error) {
	sanitizePat := strings.TrimSpace(cfg.GlobalSearchURLSanitizeRegex)
	filterCT := resolveGlobalSearchFilterCloudTypes(passedSessionCT, apis)
	cacheKey := globalSearchResultCacheKey(kw, filterCT, apis, sanitizePat)
	if b, ok := service.GetSearchCache(context.Background(), cacheKey); ok && len(b) > 0 {
		var cached []globalSearchItem
		if err := json.Unmarshal(b, &cached); err == nil && len(cached) >= 0 {
			return cached, nil
		}
	}

	client := &http.Client{Timeout: 12 * time.Second}
	items := make([]globalSearchItem, 0, 64)
	byKey := map[string]globalSearchItem{}
	for _, api := range apis {
		apiURL := strings.TrimSpace(api.APIURL)
		if apiURL == "" {
			continue
		}
		ct, ok := perLineGlobalSearchRequestCloudTypes(strings.TrimSpace(passedSessionCT), api.CloudTypes)
		if !ok {
			continue
		}
		q := url.Values{}
		q.Set("kw", kw)
		if ct != "" {
			q.Set("cloud_types", ct)
		}
		reqURL := apiURL + "?" + q.Encode()
		resp, err := client.Get(reqURL)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			continue
		}
		var root map[string]any
		if err := json.Unmarshal(body, &root); err != nil {
			continue
		}
		merged := mergedByTypeFromRoot(root)
		if len(merged) > 0 {
			for cloudType, rawList := range merged {
				arr, _ := rawList.([]any)
				for _, row := range arr {
					m, _ := row.(map[string]any)
					ingestGlobalSearchRow(m, cloudType, byKey, sanitizePat)
				}
			}
			// 已按类型拆分时，data.list 常为全量混排；在已指定 cloud_types 时不再合并，避免混入其它网盘。
			if ct == "" {
				if data, has := root["data"]; has {
					if dm, ok := data.(map[string]any); ok {
						if list, ok := dm["list"].([]any); ok && len(list) > 0 {
							appendGlobalSearchFromFlatList(list, "", byKey, sanitizePat)
						}
					}
				}
			}
		} else {
			if data, has := root["data"]; has {
				appendGlobalSearchFromDataUnknownShape(data, byKey, sanitizePat)
			}
			if list, ok := root["list"].([]any); ok && len(list) > 0 {
				appendGlobalSearchFromFlatList(list, "", byKey, sanitizePat)
			}
			if itemsRoot, ok := root["items"].([]any); ok && len(itemsRoot) > 0 {
				appendGlobalSearchFromFlatList(itemsRoot, "", byKey, sanitizePat)
			}
		}
	}
	for _, v := range byKey {
		items = append(items, v)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("所有全网搜接口均失败或无结果")
	}
	// 稳定排序：网盘优先级 -> 时间倒序 -> 去重键字典序（保证一致性）
	sort.SliceStable(items, func(i, j int) bool {
		ai, aj := cloudTypeOrder(items[i].CloudType), cloudTypeOrder(items[j].CloudType)
		if ai != aj {
			return ai < aj
		}
		ti, tj := parseSearchDatetime(items[i].Datetime), parseSearchDatetime(items[j].Datetime)
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		return globalDedupeKey(items[i]) < globalDedupeKey(items[j])
	})
	if raw, err := json.Marshal(items); err == nil {
		// 聚合结果短缓存：减少外部 API 压力。失效过滤在 PublicGlobalSearch 再实时处理。
		service.SetSearchCacheWithTTL(context.Background(), cacheKey, raw, 2*time.Minute)
	}
	return items, nil
}

func parseSelectedPlatforms(cloudTypes string) []string {
	parts := strings.Split(strings.TrimSpace(cloudTypes), ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		v := strings.ToLower(strings.TrimSpace(p))
		if v == "" {
			continue
		}
		switch v {
		case "115":
			v = "pan115"
		case "123":
			v = "pan123"
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func normalizeGlobalSearchCloudToken(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "115":
		return "pan115"
	case "123":
		return "pan123"
	default:
		return v
	}
}

// splitGlobalSearchCloudTypesCSV 解析逗号分隔的云类型（去重、保序）。
func splitGlobalSearchCloudTypesCSV(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		v := normalizeGlobalSearchCloudToken(p)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// intersectGlobalSearchCloudTypesCSV 返回 a∩b 的逗号串；无交集时返回空串。
func intersectGlobalSearchCloudTypesCSV(a, b string) string {
	if strings.TrimSpace(a) == "" || strings.TrimSpace(b) == "" {
		return ""
	}
	setB := map[string]struct{}{}
	for _, t := range splitGlobalSearchCloudTypesCSV(b) {
		setB[t] = struct{}{}
	}
	out := make([]string, 0)
	seen := map[string]struct{}{}
	for _, t := range splitGlobalSearchCloudTypesCSV(a) {
		if _, ok := setB[t]; !ok {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return strings.Join(out, ",")
}

// unionCloudTypesFromGlobalSearchAPIs 合并各已启用线路的非空 cloud_types。
// 若任一线路 cloud_types 为空，视为该线路不限制类型，返回 ""（不能用线路表单独收紧全局）。
func unionCloudTypesFromGlobalSearchAPIs(apis []model.GlobalSearchAPI) string {
	anyUnrestricted := false
	seen := map[string]struct{}{}
	list := make([]string, 0)
	for _, api := range apis {
		line := strings.TrimSpace(api.CloudTypes)
		if line == "" {
			anyUnrestricted = true
			continue
		}
		for _, t := range splitGlobalSearchCloudTypesCSV(line) {
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			list = append(list, t)
		}
	}
	if anyUnrestricted || len(list) == 0 {
		return ""
	}
	sort.Strings(list)
	return strings.Join(list, ",")
}

// resolveGlobalSearchFilterCloudTypes 得到应对聚合结果做 cloud_type 收敛的类型串。
// 当系统/请求里写了多种网盘，但与「仅允许这些类型的线路」取交集后，可收敛为单盘（如仅百度）。
func resolveGlobalSearchFilterCloudTypes(passedCT string, apis []model.GlobalSearchAPI) string {
	passedCT = strings.TrimSpace(passedCT)
	union := unionCloudTypesFromGlobalSearchAPIs(apis)
	if passedCT == "" {
		return union
	}
	if union == "" {
		return passedCT
	}
	if x := intersectGlobalSearchCloudTypesCSV(passedCT, union); x != "" {
		return x
	}
	return passedCT
}

// perLineGlobalSearchRequestCloudTypes 单条线路实际请求上游时携带的 cloud_types；第二个返回值 false 表示该线路与当前筛选互斥应跳过。
func perLineGlobalSearchRequestCloudTypes(sessionCT, lineCT string) (string, bool) {
	s := strings.TrimSpace(sessionCT)
	l := strings.TrimSpace(lineCT)
	if l != "" && s != "" {
		v := intersectGlobalSearchCloudTypesCSV(s, l)
		if v == "" {
			return "", false
		}
		return v, true
	}
	if l != "" {
		return l, true
	}
	if s != "" {
		return s, true
	}
	return "", true
}

func loadGlobalSearchAPIRows(cfg model.SystemConfig) []model.GlobalSearchAPI {
	var apis []model.GlobalSearchAPI
	_ = database.DB().Where("enabled = ?", true).Order("sort_order DESC, id ASC").Find(&apis).Error
	if len(apis) == 0 {
		apis = []model.GlobalSearchAPI{{
			ID:         0,
			Name:       "default",
			APIURL:     globalSearchAPIURL(cfg),
			CloudTypes: strings.TrimSpace(cfg.GlobalSearchCloudTypes),
			Enabled:    true,
		}}
	}
	return apis
}

func asyncCheckGlobalLinks(items []globalSearchItem, cfg model.SystemConfig, selectedPlatforms []string) {
	select {
	case globalSearchCheckSem <- struct{}{}:
	default:
		return
	}
	go func() {
		defer func() {
			<-globalSearchCheckSem
			recover()
		}()
		seen := map[string]struct{}{}
		links := make([]string, 0, len(items))
		for _, it := range items {
			u := strings.TrimSpace(it.URL)
			if u == "" {
				continue
			}
			n := normalizeGlobalURL(u)
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			if cached, ok := getGlobalLinkStatus(u); ok && (cached.Status == "valid" || cached.Status == "invalid") {
				continue
			}
			links = append(links, u)
			if len(links) >= 40 {
				break
			}
		}
		enqueueGlobalLinkChecks(links, cfg, selectedPlatforms)
	}()
}

// PublicGlobalSearch 全网聚合搜索（公开）
func PublicGlobalSearch(c *gin.Context) {
	kw := strings.TrimSpace(c.Query("q"))
	if kw == "" {
		kw = strings.TrimSpace(c.Query("kw"))
	}
	if kw == "" {
		response.Error(c, 400, "q 不能为空")
		return
	}
	cfg := readGlobalSearchConfig()
	if !cfg.GlobalSearchEnabled {
		response.Error(c, 400, "全网搜未开启")
		return
	}
	passedCT := strings.TrimSpace(c.Query("cloud_types"))
	if passedCT == "" {
		passedCT = strings.TrimSpace(cfg.GlobalSearchCloudTypes)
	}
	apis := loadGlobalSearchAPIRows(cfg)
	filterCT := resolveGlobalSearchFilterCloudTypes(passedCT, apis)
	items, err := fetchGlobalSearchWithAPIs(kw, passedCT, apis, cfg)
	if err != nil {
		response.Error(c, 500, "全网搜失败: "+err.Error())
		return
	}
	if strings.TrimSpace(filterCT) != "" {
		items = filterGlobalSearchByCloudTypes(items, filterCT)
	}
	effectiveCloudTypesOut := filterCT
	if effectiveCloudTypesOut == "" {
		effectiveCloudTypesOut = passedCT
	}
	if cfg.HideInvalidLinksInSearch {
		items = filterKnownInvalidGlobalItems(items)
	}
	if !cfg.GlobalSearchLinkCheckEnabled {
		response.OK(c, gin.H{
			"list":        items,
			"total":       len(items),
			"q":           kw,
			"cloud_types": effectiveCloudTypesOut,
		})
		return
	}
	selectedPlatforms := parseSelectedPlatforms(effectiveCloudTypesOut)
	asyncCheckGlobalLinks(items, cfg, selectedPlatforms)

	withStatus := make([]globalSearchItem, 0, len(items))
	for _, it := range items {
		st := "unknown"
		if cached, ok := getGlobalLinkStatus(it.URL); ok {
			st = strings.TrimSpace(cached.Status)
			if st == "" {
				st = "unknown"
			}
		}
		it.LinkStatus = st
		// 仅在系统配置开启“搜索中隐藏无效链接”时隐藏
		if cfg.HideInvalidLinksInSearch && it.LinkStatus == "invalid" {
			continue
		}
		withStatus = append(withStatus, it)
	}
	response.OK(c, gin.H{
		"list":        withStatus,
		"total":       len(withStatus),
		"q":           kw,
		"cloud_types": effectiveCloudTypesOut,
	})
}

// AdminGlobalSearchTest 测试全网搜接口配置是否可用
func AdminGlobalSearchTest(c *gin.Context) {
	cfg := readGlobalSearchConfig()
	kw := strings.TrimSpace(c.DefaultQuery("q", "测试"))
	passedCT := strings.TrimSpace(c.Query("cloud_types"))
	if passedCT == "" {
		passedCT = strings.TrimSpace(cfg.GlobalSearchCloudTypes)
	}
	apis := loadGlobalSearchAPIRows(cfg)
	filterCT := resolveGlobalSearchFilterCloudTypes(passedCT, apis)
	items, err := fetchGlobalSearchWithAPIs(kw, passedCT, apis, cfg)
	if err == nil && strings.TrimSpace(filterCT) != "" {
		items = filterGlobalSearchByCloudTypes(items, filterCT)
	}
	if err != nil {
		response.OK(c, gin.H{
			"ok":      false,
			"message": err.Error(),
			"api_url": globalSearchAPIURL(cfg),
		})
		return
	}
	response.OK(c, gin.H{
		"ok":      true,
		"message": "连接正常",
		"api_url": globalSearchAPIURL(cfg),
		"count":   len(items),
	})
}

// PublicGlobalSearchClaim 一键获取链接：入库并可自动触发转存
func PublicGlobalSearchClaim(c *gin.Context) {
	var req struct {
		URL       string `json:"url" binding:"required"`
		Password  string `json:"password"`
		Note      string `json:"note"`
		Source    string `json:"source"`
		CloudType string `json:"cloud_type"`
		Image     string `json:"image"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	cfg := readGlobalSearchConfig()
	if !cfg.GlobalSearchEnabled {
		response.Error(c, 400, "全网搜未开启")
		return
	}

	sanitizePat := strings.TrimSpace(cfg.GlobalSearchURLSanitizeRegex)
	link := sanitizeGlobalSearchURLString(strings.TrimSpace(req.URL), sanitizePat)
	if link == "" {
		response.Error(c, 400, "url 不能为空")
		return
	}
	title := strings.TrimSpace(req.Note)
	if title == "" {
		title = "全网搜资源"
	}

	catID := cfg.GlobalSearchDefaultCategoryID
	if catID == 0 {
		var cat model.Category
		if err := database.DB().Where("status = 1").Order("id ASC").First(&cat).Error; err == nil {
			catID = cat.ID
		}
	}
	if catID == 0 {
		response.Error(c, 400, "请先创建资源分类或配置默认分类")
		return
	}

	var exist model.Resource
	if err := database.DB().Where("link = ?", link).Order("id DESC").First(&exist).Error; err == nil {
		response.OK(c, gin.H{"resource_id": exist.ID, "exists": true})
		return
	}

	res := model.Resource{
		Title:       title,
		Link:        link,
		CategoryID:  catID,
		Source:      "global_search",
		Description: strings.TrimSpace(req.Note),
		ExtractCode: strings.TrimSpace(req.Password),
		Cover:       strings.TrimSpace(req.Image),
		Tags:        strings.TrimSpace(req.CloudType),
		LinkValid:   true,
		Status:      1,
	}
	if err := database.DB().Create(&res).Error; err != nil {
		response.Error(c, 500, "入库失败")
		return
	}
	service.MeiliUpsertResourceAsync(res.ID)

	autoTransfer := cfg.GlobalSearchAutoTransfer
	if autoTransfer {
		if cred, err := service.LoadNetdiskCredentials(); err == nil && service.ShouldAutoTransferOnCreate(cred, res.Link) {
			service.MarkResourceTransferPending(res.ID, "全网搜获取链接后等待转存")
			rid := res.ID
			go func() {
				defer func() { recover() }()
				_ = service.TransferResourceWithRetry(rid, 3)
			}()
		}
	}
	response.OK(c, gin.H{
		"resource_id":   res.ID,
		"exists":        false,
		"auto_transfer": autoTransfer,
	})
}

// PublicGlobalSearchGetLink 获取转存后的链接（不入库）
func PublicGlobalSearchGetLink(c *gin.Context) {
	var req struct {
		URL      string `json:"url" binding:"required"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	cfg := readGlobalSearchConfig()
	if !cfg.GlobalSearchEnabled {
		response.Error(c, 400, "全网搜未开启")
		return
	}
	sanitizePat := strings.TrimSpace(cfg.GlobalSearchURLSanitizeRegex)
	link := sanitizeGlobalSearchURLString(strings.TrimSpace(req.URL), sanitizePat)
	pass := strings.TrimSpace(req.Password)
	if link == "" {
		response.Error(c, 400, "url 不能为空")
		return
	}
	platform := service.DetectTransferPlatform(link)
	if platform == service.PlatformUnknown {
		fallback := strings.TrimSpace(link)
		if fallback != "" && strings.TrimSpace(pass) != "" {
			if u, err := url.Parse(fallback); err == nil {
				q := u.Query()
				if strings.TrimSpace(q.Get("pwd")) == "" && strings.TrimSpace(q.Get("passcode")) == "" {
					q.Set("pwd", strings.TrimSpace(pass))
					u.RawQuery = q.Encode()
					fallback = u.String()
				}
			}
		}
		response.OK(c, gin.H{
			"platform":         "unknown",
			"link":             fallback,
			"message":          "",
			"status":           "fallback",
			"own_share_source": "fallback",
			"fallback_reason":  "platform_unsupported",
		})
		return
	}

	var outLink string
	var message string
	status := "success"
	ownShareSource := ""
	fallbackReason := ""
	switch platform {
	case service.PlatformQuark:
		r, err := service.QuarkSaveByShareLink(link, pass)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		outLink = strings.TrimSpace(r.OwnShareURL)
		message = strings.TrimSpace(r.Message)
		if strings.TrimSpace(r.Status) != "" {
			status = strings.TrimSpace(r.Status)
		}
		ownShareSource = strings.TrimSpace(r.OwnSource)
		fallbackReason = strings.TrimSpace(r.Reason)
	case service.PlatformBaidu:
		r, err := service.BaiduSaveByShareLink(link, pass)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		outLink = strings.TrimSpace(r.OwnShareURL)
		message = strings.TrimSpace(r.Message)
	case service.PlatformUC:
		r, err := service.UcSaveByShareLink(link, pass)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		outLink = strings.TrimSpace(r.OwnShareURL)
		message = strings.TrimSpace(r.Message)
	case service.PlatformXunlei:
		r, err := service.XunleiSaveByShareLink(link, pass)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		outLink = strings.TrimSpace(r.OwnShareURL)
		message = strings.TrimSpace(r.Message)
	default:
		status = "fallback"
		outLink = strings.TrimSpace(link)
		if outLink != "" && strings.TrimSpace(pass) != "" {
			if u, err := url.Parse(outLink); err == nil {
				q := u.Query()
				if strings.TrimSpace(q.Get("pwd")) == "" && strings.TrimSpace(q.Get("passcode")) == "" {
					q.Set("pwd", strings.TrimSpace(pass))
					u.RawQuery = q.Encode()
					outLink = u.String()
				}
			}
		}
		message = ""
		fallbackReason = "platform_unsupported"
	}
	if outLink == "" {
		// 容错：部分场景转存已提交，但短时间内无法定位到本人网盘目标文件用于创建分享；
		// 回退返回原始分享链接，避免前台直接失败。
		if status == "success" {
			status = "fallback"
		}
		fallback := strings.TrimSpace(link)
		if fallback != "" && strings.TrimSpace(pass) != "" {
			if u, err := url.Parse(fallback); err == nil {
				q := u.Query()
				if strings.TrimSpace(q.Get("pwd")) == "" && strings.TrimSpace(q.Get("passcode")) == "" {
					q.Set("pwd", strings.TrimSpace(pass))
					u.RawQuery = q.Encode()
					fallback = u.String()
				}
			}
		}
		outLink = fallback
		if outLink == "" {
			response.Error(c, 500, "转存已提交，但未生成可用链接，请稍后重试")
			return
		}
		// 不向前台拼接「已返回原始链接」类话术；若有网盘返回的 message（如失败原因）则保留
		if fallbackReason == "" {
			fallbackReason = "own_share_not_ready"
		}
	}
	response.OK(c, gin.H{
		"platform":          platform.String(),
		"link":              outLink,
		"message":           message,
		"status":            status,
		"own_share_source":  ownShareSource,
		"fallback_reason":   fallbackReason,
	})
}
