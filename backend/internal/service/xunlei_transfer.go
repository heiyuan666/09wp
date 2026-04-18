package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	xunleiClientID = "Xqp0kJBXWhwaTpB6"
	xunleiDeviceID = "925b7631473a13716b791d7f28289cad"
	// 以下与 xinyue-search XunleiPan.php getCaptchaToken 保持一致（网页端 shield 常用固定签名）
	xunleiLegacyCaptchaSign      = "1.fe2108ad808a74c9ac0243309242726c"
	xunleiLegacyCaptchaTimestamp = "1645241033384"
)

var reXunleiShare = regexp.MustCompile(`(?i)(?:https?://)?pan\.xunlei\.com/s/([a-zA-Z0-9_-]+)`)

type XunleiTransferResult struct {
	ShareID      string   `json:"share_id"`
	Title        string   `json:"title,omitempty"`
	Message      string   `json:"message"`
	OwnShareURL  string   `json:"own_share_url,omitempty"`
	SavedFileIDs []string `json:"saved_file_ids,omitempty"`
}

func XunleiSaveByShareLink(link string, passcodeOverride string) (XunleiTransferResult, error) {
	cfg, err := LoadNetdiskCredentials()
	if err != nil {
		return XunleiTransferResult{}, err
	}
	picked := PickXunleiRefreshToken(cfg)
	refreshToken := picked.Cookie
	if refreshToken == "" {
		return XunleiTransferResult{}, fmt.Errorf("请先在「网盘凭证」填写迅雷 refresh_token（或多账号轮询列表）")
	}

	shareID, passcode, err := parseXunleiShare(link)
	if err != nil {
		return XunleiTransferResult{}, err
	}
	if strings.TrimSpace(passcodeOverride) != "" {
		passcode = strings.TrimSpace(passcodeOverride)
	}
	targetParentID := effectiveXunleiFolder(picked, cfg.XunleiTargetFolderID)

	client := &http.Client{Timeout: 30 * time.Second}
	accessToken, newRefresh, err := xunleiGetAccessToken(client, refreshToken)
	if err != nil {
		return XunleiTransferResult{}, err
	}
	if err := PersistXunleiRefreshToken(refreshToken, newRefresh); err != nil {
		log.Printf("persist xunlei refresh_token: %v", err)
	}

	// 对齐 xinyue-search：只请求一次 get:/drive/v1/share 的 captcha，全流程共用（含 share / restore / tasks / 建分享）
	captchaTok, err := xunleiInitCaptchaToken(client)
	if err != nil {
		return XunleiTransferResult{}, err
	}
	driveHeaders := xunleiHeaders(accessToken, captchaTok)

	shareInfo, err := xunleiGetShare(client, shareID, passcode, driveHeaders)
	if err != nil {
		return XunleiTransferResult{}, err
	}
	passCodeToken, _ := shareInfo["pass_code_token"].(string)

	fileIDs := xunleiExtractShareFileIDs(shareInfo)
	if len(fileIDs) == 0 {
		return XunleiTransferResult{}, fmt.Errorf("分享内容为空或无法解析文件")
	}

	restoreTaskID, err := xunleiRestore(client, shareID, passCodeToken, targetParentID, fileIDs, driveHeaders)
	if err != nil {
		return XunleiTransferResult{}, err
	}
	taskData, err := xunleiWaitTask(client, restoreTaskID, driveHeaders)
	if err != nil {
		return XunleiTransferResult{}, err
	}
	traceIDs := xunleiExtractTraceFileIDs(taskData)
	if len(traceIDs) == 0 {
		return XunleiTransferResult{}, fmt.Errorf("转存成功但未解析到目标文件")
	}

	title := ""
	if t, ok := shareInfo["title"].(string); ok {
		title = strings.TrimSpace(t)
	} else if data, ok := shareInfo["data"].(map[string]any); ok {
		if t2, ok2 := data["title"].(string); ok2 {
			title = strings.TrimSpace(t2)
		}
	}
	out := XunleiTransferResult{
		ShareID:      shareID,
		Title:        title,
		Message:      "转存成功",
		SavedFileIDs: append([]string{}, traceIDs...),
	}
	if cfg.ReplaceLinkAfterTransfer {
		ownURL, err := xunleiCreateOwnShare(client, traceIDs, driveHeaders)
		if err != nil {
			out.Message = "转存成功（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
		} else {
			out.OwnShareURL = ownURL
		}
	}
	return out, nil
}

// DeleteXunleiFiles 删除迅雷网盘中的文件（file_id 列表）。
func DeleteXunleiFiles(fileIDs []string) error {
	cfg, err := LoadNetdiskCredentials()
	if err != nil {
		return err
	}
	picked := PickXunleiRefreshToken(cfg)
	refreshToken := strings.TrimSpace(picked.Cookie)
	if refreshToken == "" {
		return fmt.Errorf("迅雷 refresh_token 未配置")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	accessToken, _, err := xunleiGetAccessToken(client, refreshToken)
	if err != nil {
		return err
	}
	captchaTok, err := xunleiInitCaptchaToken(client)
	if err != nil {
		return err
	}
	headers := xunleiHeaders(accessToken, captchaTok)
	seen := map[string]struct{}{}
	for _, id := range fileIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		u := "https://api-pan.xunlei.com/drive/v1/files/" + url.PathEscape(id) + "?trash=true"
		resp, err := xunleiDoJSON(client, http.MethodDelete, u, headers, nil)
		if err != nil {
			return err
		}
		if err := xunleiRespErr(resp); err != nil {
			return err
		}
	}
	return nil
}

func parseXunleiShare(link string) (string, string, error) {
	m := reXunleiShare.FindStringSubmatch(strings.TrimSpace(link))
	if len(m) < 2 {
		return "", "", fmt.Errorf("不是有效的迅雷分享链接")
	}
	shareID := strings.TrimSpace(m[1])
	pass := ""
	u, err := url.Parse(strings.TrimSpace(link))
	if err == nil {
		pass = strings.TrimSpace(u.Query().Get("pwd"))
	}
	return shareID, pass, nil
}

// xunleiDoJSON 调用 api-pan.xunlei.com：HTTP 4xx 时尝试按业务 JSON 解析（含 captcha_invalid）
func xunleiDoJSON(client *http.Client, method, endpoint string, headers map[string]string, body []byte) (map[string]any, error) {
	var rd io.Reader
	if len(body) > 0 {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, endpoint, rd)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var j map[string]any
		if json.Unmarshal(raw, &j) == nil {
			if err := xunleiRespErr(j); err != nil {
				return nil, err
			}
		}
		return nil, fmt.Errorf("迅雷错误: %s", string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("迅雷返回解析失败")
	}
	return out, nil
}

const xunleiWebClientVersion = "1.45.0"

// xunleiHeaders 对齐 xinyue-search XunleiPan.php urlHeader（浏览器访问 api-pan）
func xunleiHeaders(accessToken string, captchaToken string) map[string]string {
	return map[string]string{
		"accept":             "*/*",
		"accept-language":    "zh-CN,zh;q=0.9",
		"content-type":       "application/json",
		"origin":             "https://pan.xunlei.com",
		"referer":            "https://pan.xunlei.com/",
		"sec-ch-ua":          `"Not;A=Brand";v="99", "Google Chrome";v="139", "Chromium";v="139"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "same-site",
		"user-agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
		"x-client-id":        xunleiClientID,
		"x-client-version":   xunleiWebClientVersion,
		"x-device-id":        xunleiDeviceID,
		"authorization":      "Bearer " + accessToken,
		"x-captcha-token":    captchaToken,
	}
}

func xunleiGetAccessToken(client *http.Client, refreshToken string) (access string, newRefresh string, err error) {
	reqBody := map[string]any{
		"client_id":     xunleiClientID,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequest(http.MethodPost, "https://xluser-ssl.xunlei.com/v1/auth/token", bytes.NewReader(raw))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var j map[string]any
		if json.Unmarshal(body, &j) == nil {
			if errStr, _ := j["error"].(string); errStr == "invalid_grant" {
				desc, _ := j["error_description"].(string)
				return "", "", fmt.Errorf(
					"迅雷 refresh_token 已失效（invalid_grant）：%s。若提示已在某时刻刷新，说明旧 token 已作废，请在浏览器打开 pan.xunlei.com 登录后重新抓取 refresh_token 填入「网盘凭证」",
					trimTo255(desc),
				)
			}
		}
		return "", "", fmt.Errorf("迅雷错误: %s", string(body))
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return "", "", fmt.Errorf("迅雷 access_token 响应解析失败")
	}
	access, _ = out["access_token"].(string)
	if access == "" {
		if data, ok := out["data"].(map[string]any); ok {
			access, _ = data["access_token"].(string)
		}
	}
	if access == "" {
		return "", "", fmt.Errorf("迅雷 access_token 获取失败")
	}
	newRefresh, _ = out["refresh_token"].(string)
	if newRefresh == "" {
		if data, ok := out["data"].(map[string]any); ok {
			newRefresh, _ = data["refresh_token"].(string)
		}
	}
	return access, newRefresh, nil
}

// xunleiInitCaptchaToken 对齐 xinyue-search XunleiPan.php getCaptchaToken（shield 仅 Chrome/91 UA + 固定 meta）
func xunleiInitCaptchaToken(client *http.Client) (string, error) {
	reqBody := map[string]any{
		"client_id": xunleiClientID,
		"action":    "get:/drive/v1/share",
		"device_id": xunleiDeviceID,
		"meta": map[string]string{
			"username":       "",
			"phone_number":   "",
			"email":          "",
			"package_name":   "pan.xunlei.com",
			"client_version": xunleiWebClientVersion,
			"captcha_sign":   xunleiLegacyCaptchaSign,
			"timestamp":      xunleiLegacyCaptchaTimestamp,
			"user_id":        "0",
		},
	}
	raw, _ := json.Marshal(reqBody)
	resp, err := httpDoJSONWithHeaders(client, http.MethodPost, "https://xluser-ssl.xunlei.com/v1/shield/captcha/init", map[string]string{
		"content-type": "application/json",
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}, raw, "迅雷")
	if err != nil {
		return "", err
	}
	token, _ := resp["captcha_token"].(string)
	if token == "" {
		if data, ok := resp["data"].(map[string]any); ok {
			token, _ = data["captcha_token"].(string)
		}
	}
	if token == "" {
		return "", fmt.Errorf("迅雷 captcha_token 获取失败")
	}
	return token, nil
}

func xunleiGetShare(client *http.Client, shareID, passcode string, headers map[string]string) (map[string]any, error) {
	q := url.Values{}
	q.Set("share_id", shareID)
	q.Set("pass_code", passcode)
	q.Set("limit", "100")
	q.Set("pass_code_token", "")
	q.Set("page_token", "")
	q.Set("thumbnail_size", "SIZE_SMALL")
	endpoint := "https://api-pan.xunlei.com/drive/v1/share?" + q.Encode()
	resp, err := xunleiDoJSON(client, http.MethodGet, endpoint, headers, nil)
	if err != nil {
		return nil, err
	}
	if err := xunleiRespErr(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func xunleiExtractShareFileIDs(shareInfo map[string]any) []string {
	var files []any
	if v, ok := shareInfo["files"].([]any); ok {
		files = v
	} else if data, ok := shareInfo["data"].(map[string]any); ok {
		files, _ = data["files"].([]any)
	}
	out := make([]string, 0, len(files))
	for _, it := range files {
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		id, _ := row["id"].(string)
		if id != "" {
			out = append(out, id)
		}
	}
	return out
}

// xunleiNormalizeParentID 根目录：迅雷部分环境不接受字面 "0"，需空串且 specify_parent_id=false（见 pan.xunlei.com 与第三方实现）
func xunleiNormalizeParentID(parentID string) (id string, specify bool) {
	s := strings.TrimSpace(parentID)
	if s == "" || s == "0" || strings.EqualFold(s, "root") {
		return "", false
	}
	return s, true
}

func xunleiRestore(client *http.Client, shareID, passCodeToken, parentID string, fileIDs []string, headers map[string]string) (string, error) {
	pid, specify := xunleiNormalizeParentID(parentID)
	reqBody := map[string]any{
		"parent_id":         pid,
		"share_id":          shareID,
		"pass_code_token":   passCodeToken,
		"ancestor_ids":      []string{},
		"specify_parent_id": specify,
		"file_ids":          fileIDs,
	}
	raw, _ := json.Marshal(reqBody)
	resp, err := xunleiDoJSON(client, http.MethodPost, "https://api-pan.xunlei.com/drive/v1/share/restore", headers, raw)
	if err != nil {
		return "", err
	}
	if err := xunleiRespErr(resp); err != nil {
		return "", err
	}
	taskID, _ := resp["restore_task_id"].(string)
	if taskID == "" {
		if data, ok := resp["data"].(map[string]any); ok {
			taskID, _ = data["restore_task_id"].(string)
		}
	}
	if taskID == "" {
		return "", fmt.Errorf("未获取到迅雷转存任务 ID")
	}
	return taskID, nil
}

func xunleiWaitTask(client *http.Client, taskID string, headers map[string]string) (map[string]any, error) {
	var last map[string]any
	for i := 0; i < 20; i++ {
		endpoint := "https://api-pan.xunlei.com/drive/v1/tasks/" + url.PathEscape(taskID)
		resp, err := xunleiDoJSON(client, http.MethodGet, endpoint, headers, nil)
		if err != nil {
			return nil, err
		}
		if err := xunleiRespErr(resp); err != nil {
			return nil, err
		}
		last = resp
		progress := int(getFloat(resp, "progress"))
		if progress == 0 {
			if data, ok := resp["data"].(map[string]any); ok {
				progress = int(getFloat(data, "progress"))
				last = data
			}
		}
		if progress >= 100 {
			return last, nil
		}
		time.Sleep(1200 * time.Millisecond)
	}
	return nil, fmt.Errorf("迅雷转存超时")
}

func xunleiExtractTraceFileIDs(taskData map[string]any) []string {
	if taskData == nil {
		return nil
	}
	var traceText string
	if params, ok := taskData["params"].(map[string]any); ok {
		traceText, _ = params["trace_file_ids"].(string)
	}
	if traceText == "" {
		if p, ok := taskData["trace_file_ids"].(string); ok {
			traceText = p
		}
	}
	if traceText == "" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(traceText), &m); err != nil {
		return nil
	}
	out := make([]string, 0, len(m))
	for _, v := range m {
		switch vv := v.(type) {
		case string:
			if vv != "" {
				out = append(out, vv)
			}
		}
	}
	return out
}

func xunleiCreateOwnShare(client *http.Client, fileIDs []string, headers map[string]string) (string, error) {
	reqBody := map[string]any{
		"file_ids":        fileIDs,
		"share_to":        "copy",
		"title":           "云盘资源分享",
		"restore_limit":   "-1",
		"expiration_days": "-1",
		"params": map[string]any{
			"subscribe_push":     "false",
			"WithPassCodeInLink": "true",
		},
	}
	raw, _ := json.Marshal(reqBody)
	resp, err := xunleiDoJSON(client, http.MethodPost, "https://api-pan.xunlei.com/drive/v1/share", headers, raw)
	if err != nil {
		return "", err
	}
	if err := xunleiRespErr(resp); err != nil {
		return "", err
	}
	shareURL, _ := resp["share_url"].(string)
	passCode, _ := resp["pass_code"].(string)
	if shareURL == "" {
		if data, ok := resp["data"].(map[string]any); ok {
			shareURL, _ = data["share_url"].(string)
			passCode, _ = data["pass_code"].(string)
		}
	}
	if shareURL == "" {
		return "", fmt.Errorf("迅雷未返回分享链接")
	}
	if passCode != "" {
		return shareURL + "?pwd=" + passCode, nil
	}
	return shareURL, nil
}

func xunleiRespErr(m map[string]any) error {
	if m == nil {
		return fmt.Errorf("迅雷返回为空")
	}
	if errStr, ok := m["error"].(string); ok && errStr != "" {
		if errStr == "captcha_invalid" {
			return fmt.Errorf("迅雷风控校验未通过（captcha_invalid）。请稍后重试；若多次失败，在浏览器打开 pan.xunlei.com 完成验证后重抓 refresh_token，并避免多台服务器共用同一 token")
		}
		msg, _ := m["error_description"].(string)
		if msg == "" {
			msg = errStr
		}
		return fmt.Errorf("迅雷: %s", msg)
	}
	codeStr, hasStr := m["error_code"].(string)
	codeNum, hasNum := m["error_code"].(float64)
	if hasStr && codeStr != "" {
		if codeStr == "9" {
			return fmt.Errorf("迅雷风控校验未通过（captcha_invalid）。请稍后重试；若多次失败，在浏览器打开 pan.xunlei.com 完成验证后重抓 refresh_token，并避免多台服务器共用同一 token")
		}
		msg, _ := m["error_description"].(string)
		if msg == "" {
			msg = codeStr
		}
		return fmt.Errorf("迅雷: %s", msg)
	}
	if hasNum && codeNum != 0 {
		msg, _ := m["error_description"].(string)
		if msg == "" {
			msg = fmt.Sprintf("error_code=%.0f", codeNum)
		}
		if codeNum == 9 {
			return fmt.Errorf("迅雷风控校验未通过（captcha_invalid）。请稍后重试；若多次失败，在浏览器打开 pan.xunlei.com 完成验证后重抓 refresh_token，并避免多台服务器共用同一 token")
		}
		return fmt.Errorf("迅雷: %s", msg)
	}
	if data, ok := m["data"].(map[string]any); ok {
		if errStr, ok := data["error"].(string); ok && errStr == "captcha_invalid" {
			return fmt.Errorf("迅雷风控校验未通过（captcha_invalid）。请稍后重试；若多次失败，在浏览器打开 pan.xunlei.com 完成验证后重抓 refresh_token，并避免多台服务器共用同一 token")
		}
		if code, ok := data["error_code"].(string); ok && code != "" {
			msg, _ := data["error_description"].(string)
			if msg == "" {
				msg = code
			}
			return fmt.Errorf("迅雷: %s", msg)
		}
		if n, ok := data["error_code"].(float64); ok && n == 9 {
			return fmt.Errorf("迅雷风控校验未通过（captcha_invalid）。请稍后重试；若多次失败，在浏览器打开 pan.xunlei.com 完成验证后重抓 refresh_token，并避免多台服务器共用同一 token")
		}
	}
	return nil
}
