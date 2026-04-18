package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var quarkSharePattern = regexp.MustCompile(`https?://pan\.quark\.cn/s/([a-zA-Z0-9]+)`)

func quarkTransferDebugEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("NETDISK_TRANSFER_DEBUG")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func quarkDebugJSON(prefix string, v any) {
	if !quarkTransferDebugEnabled() {
		return
	}
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("[QUARK-TRANSFER] %s marshal_err=%v", prefix, err)
		return
	}
	s := string(b)
	if len(s) > 1200 {
		s = s[:1200] + "...(trunc)"
	}
	log.Printf("[QUARK-TRANSFER] %s %s", prefix, s)
}

type QuarkTransferResult struct {
	ShareCode   string          `json:"share_code"`
	Title       string          `json:"title,omitempty"`
	Message     string          `json:"message"`
	Raw         any             `json:"raw,omitempty"`
	OwnShareURL string          `json:"own_share_url,omitempty"`
	SavedFids   []string        `json:"saved_fids,omitempty"`
	FilterLog   *QuarkFilterLog `json:"filter_log,omitempty"`
	Status      string          `json:"status,omitempty"`            // success|pending|fallback
	OwnSource   string          `json:"own_share_source,omitempty"`  // task|folder_new_fids|fallback
	Reason      string          `json:"fallback_reason,omitempty"`   // when no own link
}

type quarkAdItem struct {
	Fid  string
	Name string
	Path string
}

type QuarkFilterLog struct {
	ScanStartTime string           `json:"scan_start_time"`
	ScanEndTime   string           `json:"scan_end_time"`
	Structure     []string         `json:"structure"`
	TotalFiles    int              `json:"total_files"`
	TotalFolders  int              `json:"total_folders"`
	AdFiles       []map[string]any `json:"ad_files"`
	DeletedFids   []string         `json:"deleted_fids"`
	Result        string           `json:"result"`
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
	if quarkTransferDebugEnabled() {
		log.Printf(
			"[QUARK-TRANSFER] account=%q ad_filter_enabled=%v banned_keywords=%q target_folder=%q replace_link_after_transfer=%v",
			strings.TrimSpace(picked.Name),
			cfg.QuarkAdFilterEnabled,
			strings.TrimSpace(cfg.QuarkBannedKeywords),
			strings.TrimSpace(folderID),
			cfg.ReplaceLinkAfterTransfer,
		)
	}

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
	quarkDebugJSON("token_resp=", tokenResp)
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
	quarkDebugJSON("detail_resp=", detailResp)
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
	// 先记录目标目录已有 fid，后续兜底时优先取“本次新出现”的文件，避免误复用旧链接。
	beforeSet := map[string]struct{}{}
	if prev, err := ucproListRecentFidsInFolder(client, ucproShareBaseURL(baseURL), cookie, folderID, 200, "夸克"); err == nil {
		for _, fid := range prev {
			fid = strings.TrimSpace(fid)
			if fid != "" {
				beforeSet[fid] = struct{}{}
			}
		}
		if quarkTransferDebugEnabled() {
			log.Printf("[QUARK-TRANSFER] before_save_folder_snapshot folder=%s count=%d", strings.TrimSpace(folderID), len(beforeSet))
		}
	}
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
	quarkDebugJSON("save_resp=", saveResp)
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
	wantTopFids := len(fids)
	if wantTopFids < 20 {
		wantTopFids = 20
	}
	if wantTopFids > 200 {
		wantTopFids = 200
	}
	pickStart := time.Now()
	topFids, topFidSource := quarkPickSavedTopFids(client, baseURL, cookie, folderID, wantTopFids, saveResp, beforeSet)
	pollElapsedMs := time.Since(pickStart).Milliseconds()
	if quarkTransferDebugEnabled() {
		log.Printf("[QUARK-TRANSFER] pick_saved_top_fids got=%d source=%s poll_elapsed_ms=%d values=%v", len(topFids), topFidSource, pollElapsedMs, topFids)
	}
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
		if quarkTransferDebugEnabled() {
			log.Printf(
				"[QUARK-TRANSFER] filtered all files: top_fids=%d keywords=%q",
				len(topFids),
				strings.TrimSpace(cfg.QuarkBannedKeywords),
			)
		}
		reason := "top_fids_empty"
		if topFidSource != "" {
			reason = topFidSource + "_empty"
		}
		if cfg.QuarkAdFilterEnabled && len(words) > 0 {
			return QuarkTransferResult{}, fmt.Errorf("资源内容为空（全部被广告词过滤）")
		}
		if cfg.ReplaceLinkAfterTransfer {
			return QuarkTransferResult{
				ShareCode: shareCode,
				Title:     strings.TrimSpace(title),
				Message:   "转存已提交，但未在目标目录定位到文件（无法生成本人分享链接），请稍后重试",
				Raw:       saveResp,
				Status:    "pending",
				OwnSource: "fallback",
				Reason:    reason,
			}, nil
		}
		return QuarkTransferResult{
			ShareCode: shareCode,
			Title:     strings.TrimSpace(title),
			Message:   "转存已提交，但未在目标目录定位到文件",
			Raw:       saveResp,
			Status:    "pending",
			OwnSource: "fallback",
			Reason:    reason,
		}, nil
	}

	out := QuarkTransferResult{
		ShareCode: shareCode, Title: strings.TrimSpace(title), Message: msg, Raw: saveResp, FilterLog: filterLog,
		SavedFids: append([]string{}, filteredTopFids...),
		Status:    "success",
		OwnSource: topFidSource,
	}
	if cfg.ReplaceLinkAfterTransfer {
		u, err := quarkCreateOwnShareLinkByFids(client, ucproShareBaseURL(baseURL), cookie, filteredTopFids, "pan.quark.cn")
		if err != nil {
			if quarkTransferDebugEnabled() {
				log.Printf("[QUARK-TRANSFER] own_share_failed err=%v", err)
			}
			out.Message = msg + "（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
			out.Status = "fallback"
			out.OwnSource = "fallback"
			out.Reason = "create_own_share_failed"
		} else {
			if quarkTransferDebugEnabled() {
				log.Printf("[QUARK-TRANSFER] own_share_success url=%s", strings.TrimSpace(u))
			}
			out.OwnShareURL = u
		}
	}
	return out, nil
}

// DeleteQuarkFilesByFids 删除夸克网盘中已转存的文件/目录（fid 列表）。
func DeleteQuarkFilesByFids(fidList []string) error {
	cfg, err := LoadNetdiskCredentials()
	if err != nil {
		return err
	}
	picked := PickQuarkCookie(cfg)
	cookie := strings.TrimSpace(picked.Cookie)
	if cookie == "" {
		return fmt.Errorf("夸克 Cookie 未配置")
	}
	client := &http.Client{Timeout: 25 * time.Second}
	baseURL := "https://drive-h.quark.cn"
	return quarkDeleteByFids(client, ucproShareBaseURL(baseURL), cookie, fidList)
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

func quarkPickSavedTopFids(client *http.Client, baseURL, cookie, folderID string, want int, saveResp map[string]any, beforeSet map[string]struct{}) ([]string, string) {
	if taskID := ucproPickTaskIDFromSaveResp(saveResp); taskID != "" {
		if quarkTransferDebugEnabled() {
			log.Printf("[QUARK-TRANSFER] pick_saved_top_fids task_id=%s want=%d", taskID, want)
		}
		if fids, err := ucproQueryTaskTopFids(client, ucproShareBaseURL(baseURL), cookie, taskID, want, "夸克"); err == nil && len(fids) > 0 {
			if quarkTransferDebugEnabled() {
				log.Printf("[QUARK-TRANSFER] task_top_fids_ok count=%d values=%v", len(fids), fids)
			}
			return fids, "task"
		} else if quarkTransferDebugEnabled() {
			log.Printf("[QUARK-TRANSFER] task_top_fids_fail err=%v", err)
		}
	}
	listLimit := want * 2
	if listLimit < 50 {
		listLimit = 50
	}
	if listLimit > 200 {
		listLimit = 200
	}
	if fids, err := ucproListRecentFidsInFolder(client, ucproShareBaseURL(baseURL), cookie, folderID, listLimit, "夸克"); err == nil && len(fids) > 0 {
		// 优先取本次转存后新增的 fid，减少“总是同一个旧链接”的概率。
		if len(beforeSet) > 0 {
			newFids := make([]string, 0, len(fids))
			for _, fid := range fids {
				fid = strings.TrimSpace(fid)
				if fid == "" {
					continue
				}
				if _, existed := beforeSet[fid]; !existed {
					newFids = append(newFids, fid)
				}
			}
			if len(newFids) > 0 {
				if quarkTransferDebugEnabled() {
					log.Printf(
						"[QUARK-TRANSFER] list_recent_new_fids_ok folder=%s count=%d values=%v",
						strings.TrimSpace(folderID),
						len(newFids),
						newFids,
					)
				}
				if len(newFids) > want {
					return newFids[:want], "folder_new_fids"
				}
				return newFids, "folder_new_fids"
			}
			// 当前目录只有旧 fid（未出现新增），说明本次转存结果尚未可见。
			// 这里不要回退到旧 fid，否则会反复生成同一个本人分享链接。
			if quarkTransferDebugEnabled() {
				log.Printf("[QUARK-TRANSFER] list_recent_only_old_fids folder=%s old_count=%d after_count=%d new_count=0", strings.TrimSpace(folderID), len(beforeSet), len(fids))
			}
			// 二次确认窗口：夸克转存偶发延迟写入，延迟后再查一次差集。
			time.Sleep(6 * time.Second)
			if refids, rerr := ucproListRecentFidsInFolder(client, ucproShareBaseURL(baseURL), cookie, folderID, listLimit, "夸克"); rerr == nil {
				reNewFids := make([]string, 0, len(refids))
				for _, fid := range refids {
					fid = strings.TrimSpace(fid)
					if fid == "" {
						continue
					}
					if _, existed := beforeSet[fid]; !existed {
						reNewFids = append(reNewFids, fid)
					}
				}
				if quarkTransferDebugEnabled() {
					log.Printf("[QUARK-TRANSFER] second_window_check folder=%s after_count=%d new_count=%d", strings.TrimSpace(folderID), len(refids), len(reNewFids))
				}
				if len(reNewFids) > 0 {
					if len(reNewFids) > want {
						return reNewFids[:want], "folder_new_fids"
					}
					return reNewFids, "folder_new_fids"
				}
			} else if quarkTransferDebugEnabled() {
				log.Printf("[QUARK-TRANSFER] second_window_check_fail folder=%s err=%v", strings.TrimSpace(folderID), rerr)
			}
			return nil, "folder_old_only"
		}
		if quarkTransferDebugEnabled() {
			log.Printf("[QUARK-TRANSFER] list_recent_fids_ok folder=%s count=%d values=%v", strings.TrimSpace(folderID), len(fids), fids)
		}
		if len(fids) > want {
			return fids[:want], "folder_recent"
		}
		return fids, "folder_recent"
	} else if quarkTransferDebugEnabled() {
		log.Printf("[QUARK-TRANSFER] list_recent_fids_fail folder=%s err=%v", strings.TrimSpace(folderID), err)
	}
	return nil, "none"
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
