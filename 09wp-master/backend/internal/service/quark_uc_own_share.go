package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// ucproProductParam chooses product param by API host.
func ucproProductParam(apiBase string) string {
	if strings.Contains(apiBase, "pc-api.uc.cn") {
		return "UCBrowser"
	}
	return "ucpro"
}

func ucproDebugEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("UC_TRANSFER_DEBUG")))
	if v == "1" || v == "true" || v == "yes" || v == "on" {
		return true
	}
	v2 := strings.ToLower(strings.TrimSpace(os.Getenv("NETDISK_TRANSFER_DEBUG")))
	return v2 == "1" || v2 == "true" || v2 == "yes" || v2 == "on"
}

func ucproDebugf(format string, args ...any) {
	if !ucproDebugEnabled() {
		return
	}
	log.Printf("[UC-OWN-SHARE] "+format, args...)
}

func ucproBuildCommonQuery(baseURL string) url.Values {
	q := url.Values{}
	q.Set("pr", ucproProductParam(baseURL))
	q.Set("fr", "pc")
	if !strings.Contains(baseURL, "pc-api.uc.cn") {
		q.Set("uc_param_str", "")
	}
	return q
}

func ucproResponseCodeStatus(resp map[string]any) (string, string, string) {
	if resp == nil {
		return "", "", ""
	}
	code := ucproToString(resp["code"])
	status := ucproToString(resp["status"])
	msg := strings.TrimSpace(ucproToString(resp["message"]))
	return code, status, msg
}

func ucproEnsureOK(resp map[string]any, apiName string) error {
	code, status, msg := ucproResponseCodeStatus(resp)
	okCode := code == "" || code == "0" || strings.EqualFold(code, "OK")
	okStatus := status == "" || status == "0" || status == "200"
	if okCode && okStatus {
		return nil
	}
	if msg == "" {
		msg = fmt.Sprintf("code=%s status=%s", code, status)
	}
	if fe := ucproUCPolicyErrorFromMessage(msg, code); fe != nil {
		return fe
	}
	return fmt.Errorf("%s失败：%s", apiName, msg)
}

// ucproSaveTaskDone UC 转存异步任务：完成态在部分账号/版本下为 3，夸克侧常见为 2（与 cloud_driver_sdk 等仅校验顶层 envelope 不同，子状态需单独判断）。
func ucproSaveTaskDone(st string) bool {
	switch st {
	case "2", "2.0", "3", "3.0":
		return true
	default:
		return false
	}
}

// ucproShareTaskDone 分享创建任务轮询完成态
func ucproShareTaskDone(st string) bool {
	return ucproSaveTaskDone(st)
}

// ucproUCPolicyErrorFromMessage 将 UC 业务码转为可读说明（如实名、容量等）
func ucproUCPolicyErrorFromMessage(message, code string) error {
	m := strings.ToLower(strings.TrimSpace(message))
	c := strings.TrimSpace(code)
	if strings.Contains(m, "not real name") || c == "32011" || strings.Contains(m, "32011") {
		return fmt.Errorf("UC 网盘：当前账号未完成实名认证，无法创建分享。请登录 https://drive.uc.cn 完成实名后再试（错误码 32011）")
	}
	return nil
}

// ucproErrorFromHTTPRaw HTTP 4xx/5xx 时解析 UC JSON 业务错误（如 32011）并返回中文说明
func ucproErrorFromHTTPRaw(label string, raw []byte) error {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err == nil && len(m) > 0 {
		if fe := ucproUCPolicyErrorFromMessage(ucproToString(m["message"]), ucproToString(m["code"])); fe != nil {
			return fe
		}
	}
	s := string(raw)
	if fe := ucproUCPolicyErrorFromMessage(s, ""); fe != nil {
		return fe
	}
	return fmt.Errorf("%s错误: %s", label, trimTo500(s))
}

// ucproReplaceWithOwnShareLink creates self-share link after transfer for Quark/UC.
// shareBaseURLOverride wins when not empty.
func ucproReplaceWithOwnShareLink(client *http.Client, baseURL, cookie, folderID, shareHost string, label string, wantFidCount int, saveResp map[string]any, shareBaseURLOverride string) (string, error) {
	if wantFidCount < 1 {
		wantFidCount = 1
	}

	shareBaseURL := ucproShareBaseURL(baseURL)
	if s := strings.TrimSpace(shareBaseURLOverride); s != "" {
		shareBaseURL = strings.TrimRight(s, "/")
	}
	ucproDebugf("start label=%s base=%s shareBase=%s folder=%s want=%d", label, baseURL, shareBaseURL, folderID, wantFidCount)

	var topFids []string

	if taskID := ucproPickTaskIDFromSaveResp(saveResp); taskID != "" {
		ucproDebugf("save response task_id=%s", taskID)
		fids, err := ucproQueryTaskTopFids(client, shareBaseURL, cookie, taskID, wantFidCount, label)
		if err == nil && len(fids) > 0 {
			topFids = fids
			ucproDebugf("task top fids success count=%d", len(topFids))
		} else if err != nil {
			ucproDebugf("task top fids failed err=%v", err)
		}
	}

	if len(topFids) == 0 {
		if fids := ucproPickFidsFromSaveResp(saveResp); len(fids) > 0 {
			if len(fids) > wantFidCount {
				fids = fids[:wantFidCount]
			}
			topFids = fids
			ucproDebugf("fallback from save resp fids count=%d", len(topFids))
		}
	}

	if len(topFids) == 0 {
		listLimit := wantFidCount * 2
		if listLimit < 50 {
			listLimit = 50
		}
		var lastListErr error
		maxAttempts := 25
		if strings.Contains(shareBaseURL, "pc-api.uc.cn") {
			maxAttempts = 55
		}
		dirCandidates := ucproFolderListCandidates(folderID)

		for attempt := 0; attempt < maxAttempts; attempt++ {
			if attempt > 0 {
				if attempt < 12 {
					time.Sleep(700 * time.Millisecond)
				} else {
					time.Sleep(1000 * time.Millisecond)
				}
			}
			for _, dir := range dirCandidates {
				fids, err := ucproListRecentFidsInFolder(client, shareBaseURL, cookie, dir, listLimit, label)
				lastListErr = err
				if err != nil {
					ucproDebugf("list attempt=%d/%d dir=%q err=%v", attempt+1, maxAttempts, dir, err)
					continue
				}
				ucproDebugf("list attempt=%d/%d dir=%q got=%d", attempt+1, maxAttempts, dir, len(fids))
				if len(fids) > 0 {
					if len(fids) > wantFidCount {
						fids = fids[:wantFidCount]
					}
					topFids = fids
					break
				}
			}
			if len(topFids) > 0 {
				break
			}
		}

		if len(topFids) == 0 {
			if lastListErr != nil {
				return "", fmt.Errorf("目标目录未列出到文件：%w", lastListErr)
			}
			return "", fmt.Errorf("目标目录未列出到文件，无法创建分享（转存可能仍在处理，请稍后重试转存或检查网盘目标文件夹）")
		}
	}

	ucproDebugf("resolved top fids count=%d", len(topFids))
	shareTaskID, err := ucproShareBtnCreateTask(client, shareBaseURL, cookie, topFids, wantFidCount, label)
	if err != nil {
		return "", err
	}
	ucproDebugf("share create task_id=%s", shareTaskID)

	shareID, err := ucproQueryTaskShareID(client, shareBaseURL, cookie, shareTaskID, label)
	if err != nil {
		return "", err
	}
	ucproDebugf("share task completed share_id=%s", shareID)

	shareURL, err := ucproSharePassword(client, shareBaseURL, cookie, shareID, shareHost, label)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(shareURL) == "" {
		return "", fmt.Errorf("未解析到分享链接")
	}
	ucproDebugf("share url generated")
	return shareURL, nil
}

// ucproFolderListCandidates 列目录时尝试的 pdir_fid：用户目录、根 0、空串（部分 UC 接口根目录表现不一致）
func ucproFolderListCandidates(folderID string) []string {
	folderID = strings.TrimSpace(folderID)
	seen := map[string]struct{}{}
	var out []string
	add := func(s string) {
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	if folderID != "" {
		add(folderID)
	}
	add("0")
	if folderID != "" && folderID != "0" {
		add("")
	}
	return out
}

func ucproJSONObjMap(v any) map[string]any {
	m, ok := v.(map[string]any)
	if ok {
		return m
	}
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil
	}
	return out
}

func ucproShareBaseURL(baseURL string) string {
	if strings.Contains(baseURL, "drive-h.quark.cn") {
		return "https://drive-pc.quark.cn"
	}
	if strings.Contains(baseURL, "drive-h.uc.cn") {
		return "https://pc-api.uc.cn"
	}
	if strings.Contains(baseURL, "pc-api.uc.cn") {
		return "https://pc-api.uc.cn"
	}
	return baseURL
}

func ucproShareBtnCreateTask(client *http.Client, shareBaseURL, cookie string, fidList []string, wantFidCount int, label string) (string, error) {
	if wantFidCount > 0 && len(fidList) > wantFidCount {
		fidList = fidList[:wantFidCount]
	}
	if len(fidList) == 0 {
		return "", fmt.Errorf("fid 列表为空")
	}

	body, _ := json.Marshal(map[string]any{
		"fid_list":     fidList,
		"expired_type": 1,
		"title":        "",
		"url_type":     1,
	})
	q := ucproBuildCommonQuery(shareBaseURL)
	q.Set("__t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	u := fmt.Sprintf("%s/1/clouddrive/share?%s", shareBaseURL, q.Encode())
	resp, err := httpDoJSONQuarkUCAuto(client, http.MethodPost, u, cookie, body, label)
	if err != nil {
		return "", err
	}
	if err := ucproEnsureOK(resp, "创建分享"); err != nil {
		return "", err
	}

	var dm map[string]any
	if d, ok := resp["data"].(map[string]any); ok {
		dm = d
	} else {
		dm = resp
	}
	if s := ucproToString(dm["task_id"]); s != "" {
		return s, nil
	}
	if s := ucproToString(dm["taskId"]); s != "" {
		return s, nil
	}
	return "", fmt.Errorf("创建分享失败：未找到 task_id")
}

func ucproQueryTaskShareID(client *http.Client, shareBaseURL, cookie, taskID string, label string) (string, error) {
	if strings.TrimSpace(taskID) == "" {
		return "", fmt.Errorf("task_id 为空")
	}
	for retry := 0; retry < 24; retry++ {
		q := ucproBuildCommonQuery(shareBaseURL)
		q.Set("task_id", taskID)
		q.Set("retry_index", strconv.Itoa(retry))
		q.Set("__t", strconv.FormatInt(time.Now().Unix(), 10))
		q.Set("__dt", fmt.Sprintf("%d", (1+rand.Intn(5))*60*1000))
		endpoint := fmt.Sprintf("%s/1/clouddrive/task?%s", shareBaseURL, q.Encode())

		resp, err := httpDoJSONQuarkUCAuto(client, http.MethodGet, endpoint, cookie, nil, label)
		if err != nil {
			return "", err
		}
		if err := ucproEnsureOK(resp, "查询分享任务"); err != nil {
			return "", err
		}

		if data, ok := resp["data"].(map[string]any); ok {
			status := ucproToString(data["status"])
			if ucproShareTaskDone(status) {
				if sid := ucproToString(data["share_id"]); sid != "" {
					return sid, nil
				}
				if sid := ucproToString(data["shareId"]); sid != "" {
					return sid, nil
				}
			}
			ucproDebugf("share task polling retry=%d status=%s", retry, status)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", fmt.Errorf("轮询分享 task 超时，无法获取 share_id")
}

func ucproSharePassword(client *http.Client, shareBaseURL, cookie, shareID, shareHost, label string) (string, error) {
	body, _ := json.Marshal(map[string]any{"share_id": shareID})
	q := ucproBuildCommonQuery(shareBaseURL)
	q.Set("__t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	u := fmt.Sprintf("%s/1/clouddrive/share/password?%s", shareBaseURL, q.Encode())
	resp, err := httpDoJSONQuarkUCAuto(client, http.MethodPost, u, cookie, body, label)
	if err != nil {
		return "", err
	}
	if err := ucproEnsureOK(resp, "获取分享链接信息"); err != nil {
		return "", err
	}

	var dm map[string]any
	if d, ok := resp["data"].(map[string]any); ok {
		dm = d
	} else {
		dm = resp
	}
	if u := ucproFindFirstString(dm, []string{"share_url", "shareUrl", "url", "shareLink", "short_url"}); u != "" {
		return u, nil
	}
	if sid := ucproFindFirstString(dm, []string{"pwd_id", "share_id", "shareId"}); sid != "" && strings.TrimSpace(shareHost) != "" {
		return "https://" + strings.TrimSpace(shareHost) + "/s/" + sid, nil
	}
	return "", fmt.Errorf("未解析到分享链接")
}

func ucproPickFidsFromSaveResp(m map[string]any) []string {
	if m == nil {
		return nil
	}
	dm, _ := m["data"].(map[string]any)
	if dm == nil {
		dm = m
	}

	var out []string
	appendArr := func(v any) {
		arr, ok := v.([]any)
		if !ok {
			return
		}
		for _, a := range arr {
			if s := ucproToString(a); s != "" {
				out = append(out, s)
			}
		}
	}
	appendArr(dm["fid_list"])
	if len(out) == 0 {
		appendArr(dm["fids"])
	}
	if len(out) == 0 {
		if saveAs, ok := dm["save_as"].(map[string]any); ok {
			appendArr(saveAs["save_as_top_fids"])
			if len(out) == 0 {
				appendArr(saveAs["top_fids"])
			}
		}
	}
	return out
}

func ucproRowFid(row map[string]any) string {
	if row == nil {
		return ""
	}
	if s := ucproToString(row["fid"]); s != "" {
		return s
	}
	if s := ucproToString(row["file_id"]); s != "" {
		return s
	}
	return ""
}

func ucproListRecentFidsInFolder(client *http.Client, baseURL, cookie, pdirFid string, limit int, label string) ([]string, error) {
	if limit < 1 {
		limit = 50
	}
	q := ucproBuildCommonQuery(baseURL)
	q.Set("pdir_fid", pdirFid)
	q.Set("_page", "1")
	q.Set("_size", strconv.Itoa(limit))
	q.Set("_fetch_total", "1")
	q.Set("_fetch_sub_dirs", "0")
	q.Set("_sort", "file_type:asc,updated_at:desc")
	if !strings.Contains(baseURL, "pc-api.uc.cn") {
		q.Set("fetch_all_file", "1")
		q.Set("fetch_risk_file_name", "1")
		q.Set("_fetch_full_path", "0")
	}
	q.Set("__t", strconv.FormatInt(time.Now().UnixMilli(), 10))

	endpoint := fmt.Sprintf("%s/1/clouddrive/file/sort?%s", baseURL, q.Encode())
	resp, err := httpDoJSONQuarkUCAuto(client, http.MethodGet, endpoint, cookie, nil, label)
	if err != nil {
		return nil, err
	}
	if err := ucproEnsureOK(resp, "列目录"); err != nil {
		return nil, err
	}

	fids := ucproExtractFidsFromFolderListResp(resp)
	ucproDebugf("list folder=%s limit=%d got=%d", pdirFid, limit, len(fids))
	return fids, nil
}

func ucproExtractFidsFromFolderListResp(resp map[string]any) []string {
	if resp == nil {
		return nil
	}
	dm, _ := resp["data"].(map[string]any)
	if dm == nil {
		dm = resp
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, 16)
	appendFromList := func(v any) {
		arr, ok := v.([]any)
		if !ok {
			return
		}
		for _, item := range arr {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			fid := ucproRowFid(row)
			if fid == "" {
				continue
			}
			if _, exists := seen[fid]; exists {
				continue
			}
			seen[fid] = struct{}{}
			out = append(out, fid)
		}
	}

	for _, key := range []string{"list", "file_list", "files", "items"} {
		appendFromList(dm[key])
	}
	if len(out) > 0 {
		return out
	}
	if res, ok := dm["result"].(map[string]any); ok {
		for _, key := range []string{"list", "file_list", "files", "items"} {
			appendFromList(res[key])
		}
	}
	return out
}

func ucproPickTaskIDFromSaveResp(saveResp map[string]any) string {
	if saveResp == nil {
		return ""
	}
	pick := func(m map[string]any) string {
		for _, k := range []string{"task_id", "taskId"} {
			if s := ucproToString(m[k]); s != "" {
				return s
			}
		}
		return ""
	}

	if dm, ok := saveResp["data"].(map[string]any); ok {
		if id := pick(dm); id != "" {
			return id
		}
		if ti, ok := dm["task_info"].(map[string]any); ok {
			if id := pick(ti); id != "" {
				return id
			}
		}
		if res, ok := dm["result"].(map[string]any); ok {
			if id := pick(res); id != "" {
				return id
			}
		}
		if t, ok := dm["task"].(map[string]any); ok {
			if id := pick(t); id != "" {
				return id
			}
		}
	}
	if s := ucproToString(saveResp["task_id"]); s != "" {
		return s
	}
	return ""
}

func ucproToString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		if x == 0 {
			return ""
		}
		return strconv.FormatInt(int64(x), 10)
	case int:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return ""
	}
}

func ucproQueryTaskTopFids(client *http.Client, baseURL, cookie, taskID string, want int, label string) ([]string, error) {
	if want < 1 {
		want = 1
	}
	maxRetry := 48
	if strings.Contains(baseURL, "drive-pc.quark.cn") {
		// 夸克全网搜场景：短等待，快速返回 pending/fallback，避免前端长时间卡住。
		maxRetry = 12 // ~6s
	}
	if strings.Contains(baseURL, "pc-api.uc.cn") {
		maxRetry = 100 // UC 转存任务偶发较慢，约 50s
	}

	for retry := 0; retry < maxRetry; retry++ {
		q := ucproBuildCommonQuery(baseURL)
		q.Set("task_id", taskID)
		q.Set("retry_index", strconv.Itoa(retry))
		q.Set("__t", strconv.FormatInt(time.Now().Unix(), 10))
		q.Set("__dt", fmt.Sprintf("%d", (1+rand.Intn(5))*60*1000))
		endpoint := fmt.Sprintf("%s/1/clouddrive/task?%s", baseURL, q.Encode())

		resp, err := httpDoJSONQuarkUCAuto(client, http.MethodGet, endpoint, cookie, nil, label)
		if err != nil {
			return nil, err
		}
		// UC/Quark 任务轮询：进行中时顶层 code/status 可能不是 200，用 ucproEnsureOK 会误杀轮询
		data, ok := resp["data"].(map[string]any)
		if !ok || len(data) == 0 {
			code, status, msg := ucproResponseCodeStatus(resp)
			if msg != "" && (strings.Contains(msg, "不存在") || strings.Contains(strings.ToLower(msg), "not found")) {
				return nil, fmt.Errorf("%s", msg)
			}
			ucproDebugf("save task poll retry=%d empty data code=%s status=%s msg=%s", retry, code, status, msg)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		st := ucproToString(data["status"])
		if ucproSaveTaskDone(st) {
			out := ucproExtractSaveAsTopFids(data)
			if len(out) > 0 {
				if len(out) > want {
					out = out[:want]
				}
				return out, nil
			}
			ucproDebugf("save task retry=%d status=%s but no top fids", retry, st)
		}
		if msg := strings.TrimSpace(ucproToString(data["message"])); msg != "" {
			if strings.Contains(strings.ToLower(msg), "capacity") {
				return nil, fmt.Errorf("%s", msg)
			}
			if st == "-1" || strings.Contains(strings.ToLower(msg), "fail") {
				return nil, fmt.Errorf("%s", msg)
			}
		}
		ucproDebugf("save task polling retry=%d status=%s", retry, st)

		time.Sleep(500 * time.Millisecond)
	}
	return nil, fmt.Errorf("轮询转存任务超时，无法获取保存后的 fid 列表")
}

func ucproExtractSaveAsTopFids(data map[string]any) []string {
	if data == nil {
		return nil
	}
	var out []string
	appendAny := func(v any) {
		arr, ok := v.([]any)
		if !ok {
			return
		}
		for _, f := range arr {
			if s := ucproToString(f); s != "" {
				out = append(out, s)
			}
		}
	}
	appendFidsFromRows := func(v any) {
		arr, ok := v.([]any)
		if !ok {
			return
		}
		for _, item := range arr {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if fid := ucproRowFid(row); fid != "" {
				out = append(out, fid)
			}
		}
	}

	saveAs := ucproJSONObjMap(data["save_as"])
	if saveAs == nil {
		saveAs = ucproJSONObjMap(getAny(data, "result"))
	}
	if saveAs != nil {
		appendAny(saveAs["save_as_top_fids"])
		if len(out) == 0 {
			appendAny(saveAs["top_fids"])
		}
		if len(out) == 0 {
			appendAny(saveAs["fids"])
		}
		if len(out) == 0 {
			appendFidsFromRows(saveAs["list"])
		}
	}
	if len(out) == 0 {
		appendAny(data["save_as_top_fids"])
	}
	if len(out) == 0 {
		appendFidsFromRows(data["list"])
	}
	return out
}

func ucproCreateShareByFids(client *http.Client, baseURL, cookie string, fidList []string, shareHost, label string) (string, error) {
	if len(fidList) == 0 {
		return "", fmt.Errorf("fid 列表为空")
	}
	body, _ := json.Marshal(map[string]any{
		"fid_list":     fidList,
		"title":        "",
		"expired_type": 1,
	})

	q := ucproBuildCommonQuery(baseURL)
	q.Set("__t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	u := fmt.Sprintf("%s/1/clouddrive/share/sharepage/create?%s", baseURL, q.Encode())
	resp, err := httpDoJSONQuarkUCAuto(client, http.MethodPost, u, cookie, body, label)
	if err != nil {
		body2, _ := json.Marshal(map[string]any{"fid_list": fidList, "expired_type": 1})
		q2 := ucproBuildCommonQuery(baseURL)
		q2.Set("__t", strconv.FormatInt(time.Now().UnixMilli(), 10))
		u2 := fmt.Sprintf("%s/1/clouddrive/share/create?%s", baseURL, q2.Encode())
		resp, err = httpDoJSONQuarkUCAuto(client, http.MethodPost, u2, cookie, body2, label)
		if err != nil {
			return "", err
		}
	}
	if link, e := ucproPickShareLinkFromCreateResp(resp, shareHost); e == nil {
		return link, nil
	}

	body2, _ := json.Marshal(map[string]any{"fid_list": fidList, "expired_type": 1})
	q3 := ucproBuildCommonQuery(baseURL)
	q3.Set("__t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	u2 := fmt.Sprintf("%s/1/clouddrive/share/create?%s", baseURL, q3.Encode())
	resp2, err2 := httpDoJSONQuarkUCAuto(client, http.MethodPost, u2, cookie, body2, label)
	if err2 != nil {
		return ucproPickShareLinkFromCreateResp(resp, shareHost)
	}
	return ucproPickShareLinkFromCreateResp(resp2, shareHost)
}

// keep-alive reference to avoid gopls unusedfunc warning (used by optional flows / future extensions)
var _ = ucproCreateShareByFids

func ucproPickShareLinkFromCreateResp(resp map[string]any, shareHost string) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("创建分享无响应")
	}
	shareHost = strings.TrimSpace(shareHost)
	if shareHost == "" {
		shareHost = "pan.quark.cn"
	}

	var dm map[string]any
	if d, ok := resp["data"].(map[string]any); ok {
		dm = d
	} else {
		dm = resp
	}

	if u := ucproFindFirstString(dm, []string{"share_url", "shareUrl", "url", "shareLink", "short_url"}); u != "" {
		return u, nil
	}
	if pwdID := ucproFindFirstString(dm, []string{"pwd_id", "pwdId", "share_id", "shareId"}); pwdID != "" {
		return "https://" + shareHost + "/s/" + pwdID, nil
	}
	if u, ok := dm["link"].(string); ok && u != "" {
		return u, nil
	}
	return "", fmt.Errorf("创建分享成功但未解析到链接")
}

func ucproFindFirstString(m map[string]any, keys []string) string {
	if m == nil {
		return ""
	}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s := strings.TrimSpace(ucproToString(v)); s != "" {
				return s
			}
		}
	}
	for _, v := range m {
		if dm, ok := v.(map[string]any); ok {
			if s := ucproFindFirstString(dm, keys); s != "" {
				return s
			}
		}
	}
	return ""
}
