package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	xunleiClientID = "Xqp0kJBXWhwaTpB6"
	xunleiDeviceID = "925b7631473a13716b791d7f28289cad"
)

var reXunleiShare = regexp.MustCompile(`(?i)(?:https?://)?pan\.xunlei\.com/s/([a-zA-Z0-9_-]+)`)

type XunleiTransferResult struct {
	ShareID     string `json:"share_id"`
	Title       string `json:"title,omitempty"`
	Message     string `json:"message"`
	OwnShareURL string `json:"own_share_url,omitempty"`
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
	accessToken, _, err := xunleiGetAccessToken(client, refreshToken)
	if err != nil {
		return XunleiTransferResult{}, err
	}
	captchaToken, err := xunleiGetCaptchaToken(client)
	if err != nil {
		return XunleiTransferResult{}, err
	}

	headers := xunleiHeaders(accessToken, captchaToken)
	shareInfo, err := xunleiGetShare(client, shareID, passcode, headers)
	if err != nil {
		return XunleiTransferResult{}, err
	}
	passCodeToken, _ := shareInfo["pass_code_token"].(string)

	fileIDs := xunleiExtractShareFileIDs(shareInfo)
	if len(fileIDs) == 0 {
		return XunleiTransferResult{}, fmt.Errorf("分享内容为空或无法解析文件")
	}

	restoreTaskID, err := xunleiRestore(client, shareID, passCodeToken, targetParentID, fileIDs, headers)
	if err != nil {
		return XunleiTransferResult{}, err
	}
	taskData, err := xunleiWaitTask(client, restoreTaskID, headers)
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
		ShareID: shareID,
		Title:   title,
		Message: "转存成功",
	}
	if cfg.ReplaceLinkAfterTransfer {
		ownURL, err := xunleiCreateOwnShare(client, traceIDs, headers)
		if err != nil {
			out.Message = "转存成功（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
		} else {
			out.OwnShareURL = ownURL
		}
	}
	return out, nil
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

func xunleiHeaders(accessToken string, captchaToken string) map[string]string {
	return map[string]string{
		"accept":          "*/*",
		"content-type":    "application/json",
		"origin":          "https://pan.xunlei.com",
		"referer":         "https://pan.xunlei.com/",
		"user-agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
		"x-client-id":     xunleiClientID,
		"x-device-id":     xunleiDeviceID,
		"authorization":   "Bearer " + accessToken,
		"x-captcha-token": captchaToken,
	}
}

func xunleiGetAccessToken(client *http.Client, refreshToken string) (string, string, error) {
	reqBody := map[string]any{
		"client_id":     xunleiClientID,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	raw, _ := json.Marshal(reqBody)
	resp, err := httpDoJSONWithHeaders(client, http.MethodPost, "https://xluser-ssl.xunlei.com/v1/auth/token", map[string]string{
		"content-type": "application/json",
		"user-agent":   "Mozilla/5.0",
	}, raw, "迅雷")
	if err != nil {
		return "", "", err
	}
	access, _ := resp["access_token"].(string)
	if access == "" {
		if data, ok := resp["data"].(map[string]any); ok {
			access, _ = data["access_token"].(string)
		}
	}
	if access == "" {
		return "", "", fmt.Errorf("迅雷 access_token 获取失败")
	}
	newRefresh, _ := resp["refresh_token"].(string)
	if newRefresh == "" {
		if data, ok := resp["data"].(map[string]any); ok {
			newRefresh, _ = data["refresh_token"].(string)
		}
	}
	return access, newRefresh, nil
}

func xunleiGetCaptchaToken(client *http.Client) (string, error) {
	reqBody := map[string]any{
		"client_id": xunleiClientID,
		"action":    "get:/drive/v1/share",
		"device_id": xunleiDeviceID,
		"meta": map[string]any{
			"package_name":   "pan.xunlei.com",
			"client_version": "1.45.0",
			"user_id":        "0",
		},
	}
	raw, _ := json.Marshal(reqBody)
	resp, err := httpDoJSONWithHeaders(client, http.MethodPost, "https://xluser-ssl.xunlei.com/v1/shield/captcha/init", map[string]string{
		"content-type": "application/json",
		"user-agent":   "Mozilla/5.0",
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
	resp, err := httpDoJSONWithHeaders(client, http.MethodGet, endpoint, headers, nil, "迅雷")
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

func xunleiRestore(client *http.Client, shareID, passCodeToken, parentID string, fileIDs []string, headers map[string]string) (string, error) {
	reqBody := map[string]any{
		"parent_id":         parentID,
		"share_id":          shareID,
		"pass_code_token":   passCodeToken,
		"ancestor_ids":      []string{},
		"specify_parent_id": true,
		"file_ids":          fileIDs,
	}
	raw, _ := json.Marshal(reqBody)
	resp, err := httpDoJSONWithHeaders(client, http.MethodPost, "https://api-pan.xunlei.com/drive/v1/share/restore", headers, raw, "迅雷")
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
		resp, err := httpDoJSONWithHeaders(client, http.MethodGet, endpoint, headers, nil, "迅雷")
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
	resp, err := httpDoJSONWithHeaders(client, http.MethodPost, "https://api-pan.xunlei.com/drive/v1/share", headers, raw, "迅雷")
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
	if code, ok := m["error_code"].(string); ok && code != "" {
		msg, _ := m["error_description"].(string)
		if msg == "" {
			msg = code
		}
		return fmt.Errorf("迅雷: %s", msg)
	}
	if data, ok := m["data"].(map[string]any); ok {
		if code, ok := data["error_code"].(string); ok && code != "" {
			msg, _ := data["error_description"].(string)
			if msg == "" {
				msg = code
			}
			return fmt.Errorf("迅雷: %s", msg)
		}
	}
	return nil
}

