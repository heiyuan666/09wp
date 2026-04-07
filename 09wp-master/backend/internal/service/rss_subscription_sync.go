package service

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
	Entries []atomEntry `xml:"entry"`
}

// rssDescription 捕获 <description> 内完整 HTML（含子节点如 <p>、<img>）。
// 仅用 string + xml:"description" 时，带子元素的 description 会被解析成空串，导致无法从 img src 取封面。
type rssDescription struct {
	inner string
}

func (d rssDescription) html() string {
	return strings.TrimSpace(d.inner)
}

func rssDescriptionFromString(v string) rssDescription {
	return rssDescription{inner: strings.TrimSpace(v)}
}

func (d *rssDescription) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	type inner struct {
		Raw string `xml:",innerxml"`
	}
	var in inner
	if err := dec.DecodeElement(&in, &start); err != nil {
		return err
	}
	d.inner = in.Raw
	return nil
}

type rssItem struct {
	Title       string         `xml:"title"`
	Description rssDescription `xml:"description"`
	Link        string         `xml:"link"`
	GUID        string         `xml:"guid"`
	PubDate     string         `xml:"pubDate"`
	Categories  []string       `xml:"category"`
}

type atomEntry struct {
	Title      string `xml:"title"`
	Summary    string `xml:"summary"`
	Content    string `xml:"content"`
	ID         string `xml:"id"`
	Updated    string `xml:"updated"`
	Published  string `xml:"published"`
	Categories []struct {
		Term string `xml:"term,attr"`
	} `xml:"category"`
	Links []struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
	} `xml:"link"`
}

type mappedRSSResource struct {
	UID         string
	Title       string
	Description string
	Link        string
	ExtraLinks  []string
	ExtractCode string
	Cover       string
	Tags        string
}

var (
	rssURLReg     = regexp.MustCompile(`https?://[^\s"'<>()]+`)
	rssImageReg   = regexp.MustCompile(`(?is)<img[^>]+src=["']([^"']+)["']`)
	rssHashTagReg = regexp.MustCompile(`#([^\s#<>\[\]{}()"'` + "`" + `,.;:!?]+)`)
	// Telegram 频道 RSS 中话题常只在链接参数里：...?q=%23%E5%8A%A8%E6%BC%AB（#动漫）
	rssTelegramTagQueryReg = regexp.MustCompile(`[?&]q=%23([^&"'<>\s]+)`)
	rssTagStripRe = regexp.MustCompile(`(?is)<[^>]+>`)
	rssTitleTrim  = regexp.MustCompile(`^[#\s\-\|:：\[\]【】()（）]+|[#\s\-\|:：\[\]【】()（）]+$`)
	// 删除描述中“关键词 + 链接”的整行（如：夸克：https://...）
	rssWholeKeywordLinkLineReg = regexp.MustCompile(`(?m)^\s*(夸克|阿里云盘|alipan|迅雷|xunlei|百度网盘|pan\.baidu\.com|百度)\s*[：:]\s*https?://[^\s]+.*$`)
	// 删除正文中仍残留的裸链接（兜底）
	rssAnyURLReg = regexp.MustCompile(`https?://[^\s"'<>()]+`)
	// 去掉清理裸链接后只剩“关键词：”这种空行残影
	rssKeywordOnlyLineReg = regexp.MustCompile(`(?m)^\s*(夸克|阿里云盘|alipan|迅雷|xunlei|百度网盘|pan\.baidu\.com|百度|网盘|链接)\s*[：:]\s*$`)
)

// SyncRSSSubscriptionByID 同步指定 RSS 订阅
func SyncRSSSubscriptionByID(subscriptionID uint64) (int, int, error) {
	var sub model.RSSSubscription
	if err := database.DB().First(&sub, subscriptionID).Error; err != nil {
		return 0, 0, err
	}
	return syncRSSSubscription(&sub)
}

// SyncAllEnabledRSSSubscriptions 同步所有启用的 RSS 订阅
func SyncAllEnabledRSSSubscriptions() (int, int, int, error) {
	var subs []model.RSSSubscription
	if err := database.DB().Where("enabled = 1").Order("id ASC").Find(&subs).Error; err != nil {
		return 0, 0, 0, err
	}
	totalAdded := 0
	totalSkipped := 0
	synced := 0
	for i := range subs {
		added, skipped, err := syncRSSSubscription(&subs[i])
		if err != nil {
			continue
		}
		totalAdded += added
		totalSkipped += skipped
		synced++
	}
	return synced, totalAdded, totalSkipped, nil
}

// TestRSSFeed 测试 RSS 地址可用性
func TestRSSFeed(feedURL string) error {
	items, err := loadRSSItems(strings.TrimSpace(feedURL))
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("rss 为空或无可解析条目")
	}
	return nil
}

func syncRSSSubscription(sub *model.RSSSubscription) (int, int, error) {
	items, err := loadRSSItems(sub.FeedURL)
	if err != nil {
		markRSSSync(sub.ID, "failed", err.Error())
		return 0, 0, err
	}

	maxItems := sub.MaxItems
	if maxItems <= 0 {
		maxItems = 50
	}
	if maxItems > 200 {
		maxItems = 200
	}
	if len(items) > maxItems {
		items = items[:maxItems]
	}

	categoryID := sub.DefaultCatID
	if categoryID == 0 {
		categoryID = resolveRSSCategoryID()
	}
	if categoryID == 0 {
		err = fmt.Errorf("未找到可用分类，请先创建分类或设置默认分类")
		markRSSSync(sub.ID, "failed", err.Error())
		return 0, len(items), err
	}

	cred, autoTransferOnRSS := SyncTimeAutoTransferAllowed()

	added := 0
	skipped := 0
	db := database.DB()
	for _, raw := range items {
		mapped := mapRSSItemToResource(raw)
		if mapped.Link == "" {
			skipped++
			continue
		}

		externalID := buildRSSExternalID(sub.ID, mapped.UID)
		var exists int64
		dupQ := db.Model(&model.Resource{}).Where("external_id = ?", externalID)
		dupQ = dupQ.Or("title = ? AND link = ?", mapped.Title, mapped.Link)
		for _, u := range mapped.ExtraLinks {
			u = strings.TrimSpace(u)
			if u == "" {
				continue
			}
			dupQ = dupQ.Or("title = ? AND link = ?", mapped.Title, u)
		}
		_ = dupQ.Count(&exists).Error
		if exists > 0 {
			skipped++
			continue
		}

		item := model.Resource{
			Title:       trimRSSText(mapped.Title, 200),
			Link:        trimRSSText(mapped.Link, 500),
			ExtraLinks:  model.NormalizeExtraShareLinks(mapped.ExtraLinks),
			CategoryID:  categoryID,
			Source:      "rss",
			ExternalID:  externalID,
			Description: trimRSSText(mapped.Description, 2000),
			ExtractCode: trimRSSText(mapped.ExtractCode, 50),
			Cover:       trimRSSText(mapped.Cover, 2048),
			Tags:        trimRSSText(mapped.Tags, 255),
			LinkValid:   true,
			Status:      1,
			SortOrder:   0,
		}
		if err := db.Create(&item).Error; err != nil {
			skipped++
			continue
		}
		added++

		if autoTransferOnRSS && ShouldAutoTransferOnCreateMulti(cred, item.Link, item.ExtraLinks) {
			MarkResourceTransferPending(item.ID, "RSS 抓取等待自动转存")
			rid := item.ID
			go func() {
				defer func() { recover() }()
				_ = TransferResourceWithRetry(rid, 3)
			}()
		}
	}

	markRSSSync(sub.ID, "success", fmt.Sprintf("同步完成: 新增%d 跳过%d", added, skipped))
	return added, skipped, nil
}

func markRSSSync(id uint64, status, msg string) {
	now := time.Now()
	_ = database.DB().Model(&model.RSSSubscription{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_sync_status": status,
		"last_sync_msg":    trimRSSText(msg, 255),
		"last_sync_at":     &now,
	}).Error
}

func resolveRSSCategoryID() uint64 {
	var cat model.Category
	if err := database.DB().Where("status = 1").Order("sort_order DESC, id ASC").First(&cat).Error; err != nil {
		return 0
	}
	return cat.ID
}

func loadRSSItems(feedURL string) ([]rssItem, error) {
	feedURL = strings.TrimSpace(feedURL)
	if feedURL == "" {
		return nil, fmt.Errorf("feed_url 不能为空")
	}
	req, _ := http.NewRequest(http.MethodGet, feedURL, nil)
	req.Header.Set("User-Agent", "DFAN-RSS-Sync/1.0")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("rss 请求失败: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	items := make([]rssItem, 0, len(feed.Channel.Items)+len(feed.Entries))
	items = append(items, feed.Channel.Items...)
	for _, e := range feed.Entries {
		link := ""
		for _, l := range e.Links {
			if strings.TrimSpace(l.Href) == "" {
				continue
			}
			if strings.TrimSpace(l.Rel) == "" || strings.EqualFold(strings.TrimSpace(l.Rel), "alternate") {
				link = l.Href
				break
			}
		}
		cats := make([]string, 0, len(e.Categories))
		for _, c := range e.Categories {
			if strings.TrimSpace(c.Term) != "" {
				cats = append(cats, c.Term)
			}
		}
		items = append(items, rssItem{
			Title:       e.Title,
			Description: rssDescriptionFromString(firstNonEmpty(e.Content, e.Summary)),
			Link:        link,
			GUID:        e.ID,
			PubDate:     firstNonEmpty(e.Published, e.Updated),
			Categories:  cats,
		})
	}
	return items, nil
}

func mapRSSItemToResource(item rssItem) mappedRSSResource {
	title := normalizeRSSTitle(item.Title)
	// 部分 RSS（如 Telegram）在 XML 里双重转义：&lt;div、href=&quot;...&quot;
	// 先解码再剥标签/抽链接，否则描述会残留整段 HTML，URL 会带上 &quot。
	descHTML := decodeHTMLEntitiesLoop(item.Description.html())

	descText := extractRSSIntro(descHTML, title)
	descText = cleanRSSDescriptionLinks(descText)

	cover := ""
	if m := rssImageReg.FindStringSubmatch(descHTML); len(m) > 1 {
		cover = strings.TrimSpace(html.UnescapeString(m[1]))
	}

	urls := extractURLs(descHTML + "\n" + decodeHTMLEntitiesLoop(item.Link))
	resourceLink, extraNetdisk := pickRSSResourceLinks(urls)
	uid := firstNonEmpty(
		strings.TrimSpace(item.GUID),
		resourceLink,
		strings.TrimSpace(item.Link),
		title+"|"+strings.TrimSpace(item.PubDate),
	)

	if title == "" {
		title = "RSS 抓取资源"
	}
	// 标签需从整段描述提取：简介 extractRSSIntro 可能只取前几行，#话题 常在段末或仅在 t.me ?q=%23 里。
	fullPlain := htmlToText(descHTML)
	tags := normalizeRSSTags(item.Categories, title+"\n"+fullPlain, extractTelegramStyleTagQueriesList(descHTML))

	return mappedRSSResource{
		UID:         uid,
		Title:       title,
		Description: descText,
		Link:        resourceLink,
		ExtraLinks:  extraNetdisk,
		ExtractCode: "",
		Cover:       cover,
		Tags:        tags,
	}
}

// cleanRSSDescriptionLinks 删除 RSS 描述中的下载/分享链接，避免前台详情页出现“夸克：https://...”等裸链接。
func cleanRSSDescriptionLinks(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return s
	}

	// 1) 删除整行关键词链接（最干净）
	s = rssWholeKeywordLinkLineReg.ReplaceAllString(s, "")
	// 2) 删除正文中所有裸 URL（兜底）
	s = rssAnyURLReg.ReplaceAllString(s, "")
	// 3) 删除清理后只剩“关键词：”的空残影行
	s = rssKeywordOnlyLineReg.ReplaceAllString(s, "")

	// 4) 压缩空行
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\n{3,}`).ReplaceAllString(s, "\n\n")
	return s
}

// pickRSSResourceLinks 返回主链接（首条网盘 URL）及后续其它网盘 URL
func pickRSSResourceLinks(urls []string) (primary string, extras []string) {
	var netdisk []string
	seen := make(map[string]struct{}, len(urls))
	for _, u := range urls {
		u = sanitizeRSSExtractedURL(u)
		u = strings.TrimSpace(strings.TrimRight(u, `,.;:!?)"'`))
		if u == "" {
			continue
		}
		if !isNetdiskURL(u) {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		netdisk = append(netdisk, u)
	}
	if len(netdisk) == 0 {
		if len(urls) == 0 {
			return "", nil
		}
		u := sanitizeRSSExtractedURL(urls[0])
		u = strings.TrimSpace(strings.TrimRight(u, `,.;:!?)"'`))
		return u, nil
	}
	return netdisk[0], netdisk[1:]
}

func isNetdiskURL(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	return strings.Contains(v, "pan.quark.cn/") ||
		strings.Contains(v, "pan.baidu.com/") ||
		strings.Contains(v, "aliyundrive.com/") ||
		strings.Contains(v, "alipan.com/") ||
		strings.Contains(v, "pan.xunlei.com/") ||
		strings.Contains(v, "drive-h.uc.cn/") ||
		strings.Contains(v, "drive.uc.cn/") ||
		strings.Contains(v, "cloud.189.cn/") ||
		strings.Contains(v, "115.com/") ||
		strings.Contains(v, "123pan.") ||
		strings.Contains(v, "123.net/")
}

func extractURLs(text string) []string {
	text = decodeHTMLEntitiesLoop(text)
	matches := rssURLReg.FindAllString(text, -1)
	// href 里的网盘链（含 &quot; 包裹时正则可能切不准，补抓一轮）
	matches = append(matches, extractURLsFromHrefs(text)...)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		m = sanitizeRSSExtractedURL(m)
		m = strings.TrimSpace(strings.TrimRight(m, `,.;:!?)"'`))
		if m == "" {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	return out
}

func extractRSSIntro(raw, title string) string {
	full := htmlToText(raw)
	if full == "" {
		return ""
	}

	lines := strings.Split(full, "\n")
	out := make([]string, 0, len(lines))
	normalizedTitle := normalizeRSSLineForCompare(title)
	for _, line := range lines {
		line = normalizeSpace(line)
		if line == "" {
			if len(out) > 0 {
				break
			}
			continue
		}
		if normalizedTitle != "" && normalizeRSSLineForCompare(line) == normalizedTitle {
			continue
		}
		if isRSSDescriptionStopLine(line) {
			break
		}
		out = append(out, line)
		if len(out) >= 3 {
			break
		}
	}
	if len(out) == 0 {
		for _, line := range lines {
			line = normalizeSpace(line)
			if line == "" || isRSSDescriptionStopLine(line) {
				continue
			}
			if normalizedTitle != "" && normalizeRSSLineForCompare(line) == normalizedTitle {
				continue
			}
			out = append(out, line)
			if len(out) >= 2 {
				break
			}
		}
	}
	return strings.Join(out, "\n")
}

// decodeHTMLEntitiesLoop 反复 html.UnescapeString，处理 &amp;quot; → &quot; → " 等多层转义。
func decodeHTMLEntitiesLoop(s string) string {
	s = strings.TrimSpace(s)
	for i := 0; i < 8; i++ {
		t := html.UnescapeString(s)
		t = strings.TrimSpace(t)
		if t == s {
			break
		}
		s = t
	}
	return s
}

var rssHrefURLReg = regexp.MustCompile(`(?is)\bhref\s*=\s*["'](https?://[^"']+)["']`)

func extractURLsFromHrefs(html string) []string {
	var out []string
	for _, m := range rssHrefURLReg.FindAllStringSubmatch(html, -1) {
		if len(m) < 2 {
			continue
		}
		u := sanitizeRSSExtractedURL(m[1])
		if u != "" {
			out = append(out, u)
		}
	}
	return out
}

func sanitizeRSSExtractedURL(u string) string {
	u = strings.TrimSpace(u)
	u = decodeHTMLEntitiesLoop(u)
	for {
		orig := u
		u = strings.TrimSuffix(u, "&quot;")
		u = strings.TrimSuffix(u, "&quot")
		u = strings.TrimSuffix(u, "&#34;")
		u = strings.TrimSuffix(u, "&#x22;")
		u = strings.TrimRight(u, `"'`+"`")
		u = strings.TrimSpace(u)
		if u == orig {
			break
		}
	}
	return u
}

func htmlToText(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	s := decodeHTMLEntitiesLoop(raw)
	replacer := strings.NewReplacer(
		"<br/>", "\n",
		"<br />", "\n",
		"<br>", "\n",
		"</p>", "\n",
		"</div>", "\n",
	)
	s = replacer.Replace(s)
	s = rssTagStripRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)

	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = normalizeSpace(line)
		if line == "" {
			continue
		}
		if isRSSMetaLine(line) {
			continue
		}
		if isOnlyURLLine(line) {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func normalizeRSSLineForCompare(s string) string {
	s = normalizeSpace(html.UnescapeString(s))
	s = rssTitleTrim.ReplaceAllString(s, "")
	return strings.ToLower(s)
}

func normalizeRSSTitle(raw string) string {
	title := normalizeSpace(html.UnescapeString(raw))
	title = stripTitleKVPrefix(title)
	title = strings.TrimSpace(title)
	if title == "" {
		return "RSS 抓取资源"
	}
	return title
}

func stripTitleKVPrefix(title string) string {
	for _, sep := range []string{"：", ":"} {
		idx := strings.Index(title, sep)
		if idx <= 0 {
			continue
		}
		key := normalizeSpace(title[:idx])
		val := normalizeSpace(title[idx+len(sep):])
		if val == "" {
			return title
		}
		keyLower := strings.ToLower(key)
		if strings.Contains(keyLower, "name") || strings.Contains(key, "\u540d") || strings.Contains(key, "\u79f0") {
			return val
		}
	}
	return title
}

func normalizeSpace(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return strings.TrimSpace(s)
}

func isOnlyURLLine(line string) bool {
	line = normalizeSpace(line)
	urls := extractURLs(line)
	if len(urls) != 1 {
		return false
	}
	return urls[0] == line
}

func isRSSMetaLine(line string) bool {
	line = normalizeSpace(line)
	lineLower := strings.ToLower(line)
	prefixes := []string{
		"\u94fe\u63a5", // 链接
		"\u5927\u5c0f", // 大小
		"\u6807\u7b7e", // 标签
		"link",
		"size",
		"tag",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lineLower, p) && (strings.Contains(line, "：") || strings.Contains(line, ":")) {
			return true
		}
	}
	return false
}

func isRSSDescriptionStopLine(line string) bool {
	line = normalizeSpace(line)
	if line == "" {
		return false
	}
	lineLower := strings.ToLower(line)
	stopExact := []string{
		"版本介绍",
		"版本信息",
		"资源信息",
		"下载地址",
		"下载链接",
		"网盘下载",
		"提取码",
		"解压码",
	}
	for _, s := range stopExact {
		if lineLower == strings.ToLower(s) {
			return true
		}
	}
	if strings.HasPrefix(line, "#") {
		return true
	}
	if strings.Contains(line, "网盘下载") || strings.Contains(line, "下载地址") || strings.Contains(line, "下载链接") {
		return true
	}
	if isRSSMetaLine(line) || isOnlyURLLine(line) {
		return true
	}
	return false
}

// extractTelegramStyleTagQueriesList 从 href 的 q=%23xxx（URL 编码话题）解析标签，补充正文里未出现的 #。
func extractTelegramStyleTagQueriesList(raw string) []string {
	raw = decodeHTMLEntitiesLoop(raw)
	var out []string
	seen := map[string]struct{}{}
	for _, m := range rssTelegramTagQueryReg.FindAllStringSubmatch(raw, -1) {
		if len(m) < 2 {
			continue
		}
		tag, err := url.QueryUnescape(m[1])
		if err != nil {
			tag = m[1]
		}
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func normalizeRSSTags(cats []string, text string, telegramTags []string) string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 8)

	for _, c := range cats {
		c = strings.TrimSpace(strings.TrimPrefix(c, "#"))
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}

	for _, m := range rssHashTagReg.FindAllStringSubmatch(text, -1) {
		if len(m) < 2 {
			continue
		}
		tag := strings.TrimSpace(strings.TrimPrefix(m[1], "#"))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}

	for _, t := range telegramTags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}

	if len(out) > 10 {
		out = out[:10]
	}
	return strings.Join(out, ",")
}

func buildRSSExternalID(subID uint64, uid string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(uid)))
	return fmt.Sprintf("rss:%d:%s", subID, hex.EncodeToString(sum[:]))
}

func trimRSSText(s string, n int) string {
	s = strings.TrimSpace(s)
	s = sanitizeRSSUTF8MB3(s)
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func sanitizeRSSUTF8MB3(s string) string {
	if s == "" {
		return s
	}
	return strings.Map(func(r rune) rune {
		if r > 0xFFFF {
			return -1
		}
		return r
	}, s)
}
