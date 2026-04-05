package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

)

var quarkSharePattern = regexp.MustCompile(`https?://pan\.quark\.cn/s/([a-zA-Z0-9]+)`)

type QuarkTransferResult struct {
	ShareCode   string `json:"share_code"`
	Title       string `json:"title,omitempty"`
	Message     string `json:"message"`
	Raw         any    `json:"raw,omitempty"`
	OwnShareURL string `json:"own_share_url,omitempty"`
	FilterLog   *QuarkFilterLog `json:"filter_log,omitempty"`
}

type quarkAdItem struct {
	Fid  string
	Name string
	Path string
}

type QuarkFilterLog struct {
	ScanStartTime string `json:"scan_start_time"`
	ScanEndTime   string `json:"scan_end_time"`
	Structure     []string `json:"structure"`
	TotalFiles    int `json:"total_files"`
	TotalFolders  int `json:"total_folders"`
	AdFiles       []map[string]any `json:"ad_files"`
	DeletedFids   []string `json:"deleted_fids"`
	Result        string `json:"result"`
}

func ParseQuarkShare(link string) (shareCode string, passcode string, err error) {
	m := quarkSharePattern.FindStringSubmatch(link)
	if len(m) < 2 {
		return "", "", fmt.Errorf("不是有效的夸克分享链接")
	}
	shareCode = m[1]
	u, perr := url.Parse(link)
	if perr == nil {
		passcode = strings.TrimSpace(u.Query().Get("pwd"))
	}
	return shareCode, passcode, nil
}

func QuarkSaveByShareLink(link string, passcodeOverride string) (QuarkTransferResult, error) {
	cfg, err := LoadNetdiskCredentials()
	if err != nil {
		return QuarkTransferResult{}, err
	}
	picked := PickQuarkCookie(cfg)
	cookie := picked.Cookie
	if cookie == "" {
		return QuarkTransferResult{}, fmt.Errorf("请先在「网盘凭证」页面填写夸克 Cookie（或多账号轮询列表）")
	}

	shareCode, passcode, err := ParseQuarkShare(link)
	if err != nil {
		return QuarkTransferResult{}, err
	}
	if strings.TrimSpace(passcodeOverride) != "" {
		passcode = strings.TrimSpace(passcodeOverride)
	}
	folderID := effectiveQuarkUCFolderID(picked, cfg.QuarkTargetFolderID)

	client := &http.Client{Timeout: 25 * time.Second}
	baseURL := "https://drive-h.quark.cn"

	// 1) token
	tokenReq := map[string]string{
		"pwd_id":   shareCode,
		"passcode": passcode,
	}
	tokenBody, _ := json.Marshal(tokenReq)
	tokenURL := fmt.Sprintf("%s/1/clouddrive/share/sharepage/token?pr=ucpro&fr=pc&uc_param_str=&__dt=994&__t=%d", baseURL, time.Now().UnixMilli())
	tokenResp, err := httpDoJSON(client, http.MethodPost, tokenURL, cookie, tokenBody, "夸克")
	if err != nil {
		return QuarkTransferResult{}, err
	}
	stoken, _ := getString(tokenResp, "data", "stoken")
	if stoken == "" {
		return QuarkTransferResult{}, fmt.Errorf("获取 stoken 失败")
	}

	// 2) share detail
	detailURL := fmt.Sprintf("%s/1/clouddrive/share/sharepage/detail?pr=ucpro&fr=pc&uc_param_str=&pwd_id=%s&stoken=%s&pdir_fid=0&force=0&_page=1&_size=50&_fetch_banner=1&_fetch_share=1&_fetch_total=1&_sort=file_type:asc,updated_at:desc&__dt=1589&__t=%d",
		baseURL, url.QueryEscape(shareCode), url.QueryEscape(stoken), time.Now().UnixMilli())
	detailResp, err := httpDoJSON(client, http.MethodGet, detailURL, cookie, nil, "夸克")
	if err != nil {
		return QuarkTransferResult{}, err
	}
	list, ok := getAny(detailResp, "data", "list").([]any)
	if !ok || len(list) == 0 {
		return QuarkTransferResult{}, fmt.Errorf("分享内容为空")
	}
	fids := make([]string, 0, len(list))
	tokens := make([]string, 0, len(list))
	for _, item := range list {
		m, _ := item.(map[string]any)
		fid, _ := m["fid"].(string)
		tk, _ := m["share_fid_token"].(string)
		if fid != "" && tk != "" {
			fids = append(fids, fid)
			tokens = append(tokens, tk)
		}
	}
	if len(fids) == 0 {
		return QuarkTransferResult{}, fmt.Errorf("未解析到可转存文件")
	}

	// 3) save to configured folder
	saveReq := map[string]any{
		"fid_list":       fids,
		"fid_token_list": tokens,
		"to_pdir_fid":    folderID,
		"pwd_id":         shareCode,
		"stoken":         stoken,
		"pdir_fid":       "0",
		"scene":          "link",
	}
	saveBody, _ := json.Marshal(saveReq)
	saveURL := fmt.Sprintf("%s/1/clouddrive/share/sharepage/save?pr=ucpro&fr=pc&uc_param_str=&__dt=208097&__t=%d", baseURL, time.Now().UnixMilli())
	saveResp, err := httpDoJSON(client, http.MethodPost, saveURL, cookie, saveBody, "夸克")
	if err != nil {
		return QuarkTransferResult{}, err
	}
	msg, _ := saveResp["message"].(string)
	if msg == "" {
		msg = "转存请求已提交"
	}
	title := ""
	if first, ok := list[0].(map[string]any); ok {
		title, _ = first["file_name"].(string)
		if strings.TrimSpace(title) == "" {
			title, _ = first["name"].(string)
		}
	}

	// 4) 可选：转存后递归广告过滤（后台可配置）
	topFids := quarkPickSavedTopFids(client, baseURL, cookie, folderID, len(fids), saveResp)
	filteredTopFids := topFids
	var filterLog *QuarkFilterLog
	words := parseCommaKeywords(cfg.QuarkBannedKeywords)
	if cfg.QuarkAdFilterEnabled && len(words) > 0 && len(topFids) > 0 {
		ff, log, err := quarkFilterAdFids(client, baseURL, cookie, topFids, words)
		if err != nil {
			msg = msg + "（广告过滤失败：" + trimTo255(err.Error()) + "）"
		} else {
			filteredTopFids = ff
			filterLog = log
			if log != nil && len(log.DeletedFids) > 0 {
				msg = fmt.Sprintf("%s（已过滤 %d 个广告项）", msg, len(log.DeletedFids))
			}
		}
	}
	if len(filteredTopFids) == 0 {
		return QuarkTransferResult{}, fmt.Errorf("资源内容为空（全部被广告词过滤）")
	}

	out := QuarkTransferResult{ShareCode: shareCode, Title: strings.TrimSpace(title), Message: msg, Raw: saveResp, FilterLog: filterLog}
	if cfg.ReplaceLinkAfterTransfer {
		u, err := quarkCreateOwnShareLinkByFids(client, ucproShareBaseURL(baseURL), cookie, filteredTopFids, "pan.quark.cn")
		if err != nil {
			out.Message = msg + "（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
		} else {
			out.OwnShareURL = u
		}
	}
	return out, nil
}

// quarkCreateOwnShareLinkByFids 夸克：用 fid_list 创建分享任务并返回分享链接
// 使用已验证可用的接口：POST /1/clouddrive/share -> 轮询 task -> POST /1/clouddrive/share/password
func quarkCreateOwnShareLinkByFids(client *http.Client, shareBaseURL, cookie string, fidList []string, shareHost string) (string, error) {
	taskID, err := ucproShareBtnCreateTask(client, shareBaseURL, cookie, fidList, len(fidList), "夸克")
	if err != nil {
		return "", err
	}
	shareID, err := ucproQueryTaskShareID(client, shareBaseURL, cookie, taskID, "夸克")
	if err != nil {
		return "", err
	}
	return ucproSharePassword(client, shareBaseURL, cookie, shareID, shareHost, "夸克")
}

func parseCommaKeywords(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		w := strings.TrimSpace(strings.ToLower(p))
		if w == "" {
			continue
		}
		if _, ok := seen[w]; ok {
			continue
		}
		seen[w] = struct{}{}
		out = append(out, w)
	}
	return out
}

func quarkPickSavedTopFids(client *http.Client, baseURL, cookie, folderID string, want int, saveResp map[string]any) []string {
	if taskID := ucproPickTaskIDFromSaveResp(saveResp); taskID != "" {
		if fids, err := ucproQueryTaskTopFids(client, ucproShareBaseURL(baseURL), cookie, taskID, want, "夸克"); err == nil && len(fids) > 0 {
			return fids
		}
	}
	if fids, err := ucproListRecentFidsInFolder(client, ucproShareBaseURL(baseURL), cookie, folderID, want*2, "夸克"); err == nil && len(fids) > 0 {
		if len(fids) > want {
			return fids[:want]
		}
		return fids
	}
	return nil
}

func quarkFilterAdFids(client *http.Client, baseURL, cookie string, topFids []string, words []string) ([]string, *QuarkFilterLog, error) {
	start := time.Now()
	log := &QuarkFilterLog{
		ScanStartTime: start.Format(time.RFC3339),
		Structure:     []string{},
		AdFiles:       []map[string]any{},
		DeletedFids:   []string{},
		Result:        "success",
	}
	items := make([]quarkAdItem, 0, 64)
	for _, fid := range topFids {
		log.Structure = append(log.Structure, fmt.Sprintf("[SCAN] root fid=%s", fid))
		all, tree, files, folders, err := quarkListAllItemsRecursively(client, ucproShareBaseURL(baseURL), cookie, fid, 0, 8, "")
		if err != nil {
			return topFids, log, err
		}
		items = append(items, all...)
		log.Structure = append(log.Structure, tree...)
		log.TotalFiles += files
		log.TotalFolders += folders
	}
	if len(items) == 0 {
		log.Result = "empty_file_list"
		log.ScanEndTime = time.Now().Format(time.RFC3339)
		return topFids, log, nil
	}

	normWords := make([]string, 0, len(words))
	for _, w := range words {
		w = strings.TrimSpace(strings.ToLower(w))
		if w != "" {
			normWords = append(normWords, w)
		}
	}
	if len(normWords) == 0 {
		log.Result = "no_keywords"
		log.ScanEndTime = time.Now().Format(time.RFC3339)
		return topFids, log, nil
	}

	delSet := map[string]struct{}{}
	for _, it := range items {
		name := strings.ToLower(strings.TrimSpace(it.Name))
		if name == "" {
			continue
		}
		for _, w := range normWords {
			if strings.Contains(name, w) {
				delSet[it.Fid] = struct{}{}
				log.AdFiles = append(log.AdFiles, map[string]any{
					"name":    it.Name,
					"fid":     it.Fid,
					"keyword": w,
					"path":    it.Path,
				})
				break
			}
		}
	}
	if len(delSet) == 0 {
		log.Result = "no_ads_found"
		log.ScanEndTime = time.Now().Format(time.RFC3339)
		return topFids, log, nil
	}

	// 全部命中时按“资源为空”处理：删除顶层并返回空
	if len(delSet) >= len(items) {
		_ = quarkDeleteByFids(client, ucproShareBaseURL(baseURL), cookie, topFids)
		for _, fid := range topFids {
			log.DeletedFids = append(log.DeletedFids, fid)
		}
		log.Result = "all_files_are_ads"
		log.ScanEndTime = time.Now().Format(time.RFC3339)
		return nil, log, nil
	}

	delList := make([]string, 0, len(delSet))
	for fid := range delSet {
		delList = append(delList, fid)
		log.DeletedFids = append(log.DeletedFids, fid)
	}
	_ = quarkDeleteByFids(client, ucproShareBaseURL(baseURL), cookie, delList)

	remain := make([]string, 0, len(topFids))
	for _, fid := range topFids {
		if _, deleted := delSet[fid]; !deleted {
			remain = append(remain, fid)
		}
	}
	log.Result = "partial_deletion"
	log.ScanEndTime = time.Now().Format(time.RFC3339)
	return remain, log, nil
}

func quarkListAllItemsRecursively(client *http.Client, baseURL, cookie, pdirFid string, depth, maxDepth int, parentPath string) ([]quarkAdItem, []string, int, int, error) {
	if depth > maxDepth {
		return nil, nil, 0, 0, nil
	}
	q := url.Values{}
	q.Set("pr", "ucpro")
	q.Set("fr", "pc")
	q.Set("uc_param_str", "")
	q.Set("pdir_fid", pdirFid)
	q.Set("_page", "1")
	q.Set("_size", "200")
	q.Set("_fetch_total", "1")
	q.Set("_fetch_sub_dirs", "1")
	q.Set("_sort", "file_type:asc,updated_at:desc")
	q.Set("__t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	resp, err := httpDoJSON(client, http.MethodGet, fmt.Sprintf("%s/1/clouddrive/file/sort?%s", baseURL, q.Encode()), cookie, nil, "夸克")
	if err != nil {
		return nil, nil, 0, 0, err
	}
	list, _ := getAny(resp, "data", "list").([]any)
	out := make([]quarkAdItem, 0, len(list))
	tree := make([]string, 0, len(list))
	files := 0
	folders := 0
	for _, it := range list {
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		fid, _ := row["fid"].(string)
		if fid == "" {
			continue
		}
		name, _ := row["file_name"].(string)
		if strings.TrimSpace(name) == "" {
			name, _ = row["name"].(string)
		}
		curPath := strings.TrimSpace(name)
		if strings.TrimSpace(parentPath) != "" {
			curPath = strings.TrimRight(parentPath, "/") + "/" + curPath
		}
		out = append(out, quarkAdItem{Fid: fid, Name: name, Path: curPath})
		isDir := int(getFloat(row, "dir")) == 1
		indent := strings.Repeat("  ", depth)
		if isDir {
			folders++
			tree = append(tree, fmt.Sprintf("%s[DIR]  %s", indent, strings.TrimSpace(name)))
		} else {
			files++
			tree = append(tree, fmt.Sprintf("%s[FILE] %s", indent, strings.TrimSpace(name)))
		}
		if isDir {
			sub, subTree, sf, sfd, err := quarkListAllItemsRecursively(client, baseURL, cookie, fid, depth+1, maxDepth, curPath)
			if err == nil {
				out = append(out, sub...)
				tree = append(tree, subTree...)
				files += sf
				folders += sfd
			}
		}
	}
	return out, tree, files, folders, nil
}

func quarkDeleteByFids(client *http.Client, baseURL, cookie string, fidList []string) error {
	if len(fidList) == 0 {
		return nil
	}
	body, _ := json.Marshal(map[string]any{
		"action_type":  2,
		"exclude_fids": []string{},
		"filelist":     fidList,
	})
	u := fmt.Sprintf("%s/1/clouddrive/file/delete?pr=ucpro&fr=pc&uc_param_str=&__t=%d", baseURL, time.Now().UnixMilli())
	_, err := httpDoJSON(client, http.MethodPost, u, cookie, body, "夸克")
	return err
}
