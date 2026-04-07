package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// keep-alive references to avoid gopls unusedfunc warnings (these helpers are used by optional flows / future extensions)
var (
	_ = aliyunGetResourceDriveID
	_ = aliyunCopyOne
	_ = aliyunPickNewFileIDFromCopyResp
	_ = aliyunListRecentFileIDsInParent
	_ = aliyunWalkShareFiles
)

type AliyunTransferResult struct {
	ShareID     string `json:"share_id"`
	Message     string `json:"message"`
	Raw         any    `json:"raw,omitempty"`
	OwnShareURL string `json:"own_share_url,omitempty"`
}

func aliyunTransferDebugEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("NETDISK_TRANSFER_DEBUG")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func aliyunDebugf(format string, args ...any) {
	if !aliyunTransferDebugEnabled() {
		return
	}
	fmt.Printf("[ALIYUN-TRANSFER] "+format+"\n", args...)
}

func parseAliyunShareID(link string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(link))
	if err != nil {
		return "", fmt.Errorf("解析链接失败")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("链接中未找到分享 ID")
	}
	id := parts[len(parts)-1]
	if id == "" {
		return "", fmt.Errorf("分享 ID 为空")
	}
	return id, nil
}

func aliyunRefreshAccessToken(refresh string) (access string, newRefresh string, refreshDriveID string, err error) {
	aliyunDebugf("refresh 开始 refresh_len=%d", len(strings.TrimSpace(refresh)))
	client := &http.Client{Timeout: 20 * time.Second}
	body, _ := json.Marshal(map[string]string{
		"refresh_token": strings.TrimSpace(refresh),
	})
	req, err := http.NewRequest(http.MethodPost, "https://api.aliyundrive.com/token/refresh", bytes.NewReader(body))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", "", "", fmt.Errorf("刷新阿里云盘令牌响应解析失败")
	}
	aliyunDebugf("refresh 响应 http=%d body_len=%d", resp.StatusCode, len(raw))
	if resp.StatusCode >= 400 {
		if msg, ok := m["message"].(string); ok && strings.TrimSpace(msg) != "" {
			return "", "", "", fmt.Errorf("刷新阿里云盘令牌失败: %s", msg)
		}
		return "", "", "", fmt.Errorf("刷新阿里云盘令牌失败: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if err := aliyunBizErr(m); err != nil {
		return "", "", "", err
	}
	access, _ = m["access_token"].(string)
	newRefresh, _ = m["refresh_token"].(string)
	refreshDriveID = firstNonEmptyString(
		getAnyString(m, "resource_drive_id"),
		getAnyString(m, "default_drive_id"),
	)
	if dm, ok := m["data"].(map[string]any); ok && strings.TrimSpace(refreshDriveID) == "" {
		refreshDriveID = firstNonEmptyString(
			getAnyString(dm, "resource_drive_id"),
			getAnyString(dm, "default_drive_id"),
		)
	}
	if access == "" {
		return "", "", "", fmt.Errorf("未获取到 access_token")
	}
	if newRefresh == "" {
		newRefresh = refresh
	}
	aliyunDebugf("refresh 成功 access_len=%d new_refresh_changed=%v", len(access), strings.TrimSpace(newRefresh) != strings.TrimSpace(refresh))
	return access, newRefresh, strings.TrimSpace(refreshDriveID), nil
}

// aliyunRefreshAccessTokenByRenewAPI 通过 OpenList 等第三方 renewapi 续期 access_token（官方 OpenAPI 常用）。
// 参考：https://doc.oplist.org/guide/drivers/aliyundrive_open#_2-%E5%87%86%E5%A4%87%E6%8E%A5%E5%85%A5
//
// 该接口的返回字段在不同部署/版本可能略有差异，这里做最大兼容解析：
// - access_token / accessToken
// - refresh_token / refreshToken
// - default_drive_id / resource_drive_id / drive_id / driveId
func aliyunRefreshAccessTokenByRenewAPI(renewAPIURL string, refresh string) (access string, newRefresh string, refreshDriveID string, err error) {
	renewAPIURL = strings.TrimSpace(renewAPIURL)
	if renewAPIURL == "" {
		return "", "", "", fmt.Errorf("renewapi 为空")
	}
	aliyunDebugf("renewapi 开始 url=%s refresh_len=%d", renewAPIURL, len(strings.TrimSpace(refresh)))
	client := &http.Client{Timeout: 25 * time.Second}
	body, _ := json.Marshal(map[string]string{
		"refresh_token": strings.TrimSpace(refresh),
	})
	req, err := http.NewRequest(http.MethodPost, renewAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	aliyunDebugf("renewapi 响应 http=%d body_len=%d", resp.StatusCode, len(raw))
	if resp.StatusCode >= 400 {
		sn := strings.TrimSpace(string(raw))
		if len(sn) > 800 {
			sn = sn[:800] + "…(trunc)"
		}
		if sn == "" {
			return "", "", "", fmt.Errorf("renewapi 续期失败: HTTP %d（空响应体）", resp.StatusCode)
		}
		return "", "", "", fmt.Errorf("renewapi 续期失败: HTTP %d: %s", resp.StatusCode, sn)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", "", "", fmt.Errorf("renewapi 响应解析失败")
	}
	// 兼容：部分服务会把真实数据包在 data 内
	root := m
	if dm, ok := m["data"].(map[string]any); ok && dm != nil {
		// 只有当 data 内看起来像 token 数据时才切换
		if getAnyString(dm, "access_token") != "" || getAnyString(dm, "accessToken") != "" {
			root = dm
		}
	}
	access = firstNonEmptyString(getAnyString(root, "access_token"), getAnyString(root, "accessToken"))
	newRefresh = firstNonEmptyString(getAnyString(root, "refresh_token"), getAnyString(root, "refreshToken"))
	refreshDriveID = firstNonEmptyString(
		getAnyString(root, "resource_drive_id"),
		getAnyString(root, "default_drive_id"),
		getAnyString(root, "drive_id"),
		getAnyString(root, "driveId"),
	)
	// 有些返回把 drive 信息放在 data.driver / data.user 里
	if strings.TrimSpace(refreshDriveID) == "" {
		if dm, ok := m["data"].(map[string]any); ok && dm != nil {
			refreshDriveID = firstNonEmptyString(
				getAnyString(dm, "resource_drive_id"),
				getAnyString(dm, "default_drive_id"),
				getAnyString(dm, "drive_id"),
				getAnyString(dm, "driveId"),
			)
		}
	}
	if strings.TrimSpace(access) == "" {
		// 尝试从 message/code 判断是否是业务错误
		if msg := firstNonEmptyString(getAnyString(m, "message"), getAnyString(m, "msg")); msg != "" {
			return "", "", "", fmt.Errorf("renewapi 续期失败: %s", msg)
		}
		return "", "", "", fmt.Errorf("renewapi 未返回 access_token")
	}
	if strings.TrimSpace(newRefresh) == "" {
		newRefresh = refresh
	}
	aliyunDebugf("renewapi 成功 access_len=%d new_refresh_changed=%v", len(access), strings.TrimSpace(newRefresh) != strings.TrimSpace(refresh))
	return strings.TrimSpace(access), strings.TrimSpace(newRefresh), strings.TrimSpace(refreshDriveID), nil
}

func aliyunGetResourceDriveID(client *http.Client, access string) (string, error) {
	resp, err := httpDoJSONBearerAliyun(client, http.MethodPost, "https://api.aliyundrive.com/v2/user/getDriveInfo", access, "", []byte("{}"), "阿里云盘")
	if err != nil {
		return "", err
	}
	dm, _ := resp["data"].(map[string]any)
	if dm != nil {
		if id, ok := dm["resource_drive_id"].(string); ok && id != "" {
			return id, nil
		}
		if id, ok := dm["default_drive_id"].(string); ok && id != "" {
			return id, nil
		}
	}
	if id, ok := resp["resource_drive_id"].(string); ok && id != "" {
		return id, nil
	}
	return "", fmt.Errorf("未解析到 drive_id")
}

func aliyunGetShareToken(client *http.Client, access, shareID, sharePwd string) (string, error) {
	reqBody := map[string]string{
		"share_id": shareID,
	}
	if strings.TrimSpace(sharePwd) != "" {
		reqBody["share_pwd"] = sharePwd
	}
	body, _ := json.Marshal(reqBody)
	resp, err := httpDoJSONBearerAliyun(client, http.MethodPost, "https://api.aliyundrive.com/v2/share_link/share_token", access, "", body, "阿里云盘")
	if err != nil {
		return "", err
	}
	if st, ok := resp["share_token"].(string); ok && st != "" {
		return st, nil
	}
	dm, _ := resp["data"].(map[string]any)
	if dm != nil {
		if st, ok := dm["share_token"].(string); ok && st != "" {
			return st, nil
		}
	}
	return "", fmt.Errorf("未获取到 share_token")
}

func aliyunGetShareAnonFileInfos(client *http.Client, access, shareID string) ([]string, error) {
	// xinyue-search: /adrive/v3/share_link/get_share_by_anonymous
	body, _ := json.Marshal(map[string]string{
		"share_id": shareID,
	})
	resp, err := httpDoJSONBearerAliyun(client, http.MethodPost, "https://api.aliyundrive.com/adrive/v3/share_link/get_share_by_anonymous", access, "", body, "阿里云盘")
	if err != nil {
		return nil, err
	}
	// 期望 resp['file_infos'] 为数组，元素中包含 file_id
	fileInfos, _ := resp["file_infos"].([]any)
	if len(fileInfos) == 0 {
		return nil, fmt.Errorf("未获取到分享 file_infos")
	}
	var out []string
	for _, it := range fileInfos {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		id, _ := m["file_id"].(string)
		if strings.TrimSpace(id) != "" {
			out = append(out, id)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("file_infos 中未解析到 file_id")
	}
	return out, nil
}

func aliyunBatchCopyShareFiles(client *http.Client, access, driveID, shareID, shareToken, toParent string, fileIDs []string) ([]string, error) {
	if len(fileIDs) == 0 {
		return nil, fmt.Errorf("fileIDs 为空")
	}
	requests := make([]any, 0, len(fileIDs))
	for i, fid := range fileIDs {
		req := map[string]any{
			"headers": map[string]string{
				"Content-Type": "application/json",
			},
			"id":     fmt.Sprintf("%d", i),
			"method": "POST",
			"url":    "/file/copy",
			"body": map[string]any{
				"auto_rename":       true,
				"file_id":           fid,
				"share_id":          shareID,
				"to_drive_id":       driveID,
				"to_parent_file_id": toParent,
			},
		}
		requests = append(requests, req)
	}
	payload, _ := json.Marshal(map[string]any{
		"resource": "file",
		"requests": requests,
	})
	// xinyue-search：batch 级别带 X-Share-Token
	resp, err := httpDoJSONBearerAliyun(client, http.MethodPost, "https://api.aliyundrive.com/adrive/v4/batch", access, shareToken, payload, "阿里云盘")
	if err != nil {
		return nil, err
	}
	responsesAny, _ := resp["responses"].([]any)
	if len(responsesAny) == 0 {
		// 有些返回结构可能在 data 下再包一层
		if dm, ok := resp["data"].(map[string]any); ok {
			responsesAny, _ = dm["responses"].([]any)
		}
	}
	if len(responsesAny) == 0 {
		return nil, fmt.Errorf("batch 响应无 responses")
	}
	var out []string
	for _, ri := range responsesAny {
		rm, ok := ri.(map[string]any)
		if !ok {
			continue
		}
		body, _ := rm["body"].(map[string]any)
		if body != nil {
			if codeV, ok := body["code"]; ok {
				code := int(getFloat(map[string]any{"code": codeV}, "code"))
				if code != 0 {
					msg, _ := body["message"].(string)
					if strings.TrimSpace(msg) == "" {
						if e, ok := body["error"].(string); ok {
							msg = e
						}
					}
					if strings.TrimSpace(msg) == "" {
						if e, ok := body["msg"].(string); ok {
							msg = e
						}
					}
					if strings.TrimSpace(msg) == "" {
						msg = "copy 失败"
					}
					return nil, fmt.Errorf("批量转存失败: %s", msg)
				}
			}
		}
		// file_id 从 body 中取
		if body != nil {
			if fid, ok := body["file_id"].(string); ok && strings.TrimSpace(fid) != "" {
				out = append(out, fid)
				continue
			}
		}
		// 兜底：递归查找 file_id
		if fid := aliyunFindFirstStringByKeys(rm, []string{"file_id", "fileId", "id"}); fid != "" {
			out = append(out, fid)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("batch 转存后未解析到新 file_id")
	}
	return out, nil
}

func aliyunFindFirstStringByKeys(m any, keys []string) string {
	switch x := m.(type) {
	case map[string]any:
		for _, k := range keys {
			if v, ok := x[k]; ok && v != nil {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					return s
				}
				if n := getFloat(x, k); n != 0 {
					return fmt.Sprintf("%.0f", n)
				}
			}
		}
		for _, v := range x {
			if s := aliyunFindFirstStringByKeys(v, keys); s != "" {
				return s
			}
		}
	case []any:
		for _, it := range x {
			if s := aliyunFindFirstStringByKeys(it, keys); s != "" {
				return s
			}
		}
	}
	return ""
}

func aliyunListShareFilesPage(client *http.Client, access, shareToken, shareID, parentID, marker string) (items []any, next string, err error) {
	m := map[string]any{
		"share_id":        shareID,
		"parent_file_id":  parentID,
		"limit":           100,
		"order_by":        "name",
		"order_direction": "ASC",
	}
	if marker != "" {
		m["marker"] = marker
	}
	body, _ := json.Marshal(m)
	resp, err := httpDoJSONBearerAliyun(client, http.MethodPost, "https://api.aliyundrive.com/adrive/v2/file/list_share_files", access, shareToken, body, "阿里云盘")
	if err != nil {
		return nil, "", err
	}
	dm, _ := resp["data"].(map[string]any)
	if dm == nil {
		dm = resp
	}
	rawItems, _ := dm["items"].([]any)
	nm, _ := dm["next_marker"].(string)
	return rawItems, nm, nil
}

func aliyunCopyOne(client *http.Client, access, driveID, shareID, shareToken, fileID, toParent string) (map[string]any, error) {
	body, _ := json.Marshal(map[string]any{
		"drive_id":          driveID,
		"file_id":           fileID,
		"share_id":          shareID,
		"share_token":       shareToken,
		"to_parent_file_id": toParent,
		"new_name":          "",
		"auto_rename":       true,
	})
	return httpDoJSONBearerAliyun(client, http.MethodPost, "https://api.aliyundrive.com/v2/file/copy", access, "", body, "阿里云盘")
}

func aliyunPickNewFileIDFromCopyResp(resp map[string]any) string {
	if resp == nil {
		return ""
	}
	if s, ok := getString(resp, "file_id"); ok && s != "" {
		return s
	}
	if dm, ok := resp["data"].(map[string]any); ok {
		if s, ok := dm["file_id"].(string); ok && s != "" {
			return s
		}
		if inner, ok := dm["file"].(map[string]any); ok {
			if s, ok := inner["file_id"].(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func aliyunListRecentFileIDsInParent(client *http.Client, access, driveID, parentID string, want int) []string {
	if want < 1 {
		want = 20
	}
	body, _ := json.Marshal(map[string]any{
		"drive_id":        driveID,
		"parent_file_id":  parentID,
		"limit":           100,
		"order_by":        "updated_at",
		"order_direction": "DESC",
	})
	resp, err := httpDoJSONBearerAliyun(client, http.MethodPost, "https://api.aliyundrive.com/v2/file/list", access, "", body, "阿里云盘")
	if err != nil {
		return nil
	}
	dm, _ := resp["data"].(map[string]any)
	if dm == nil {
		dm = resp
	}
	items, _ := dm["items"].([]any)
	var out []string
	for _, it := range items {
		if len(out) >= want {
			break
		}
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		fid, _ := row["file_id"].(string)
		if fid != "" {
			out = append(out, fid)
		}
	}
	return out
}

func aliyunPickShareURLFromCreate(resp map[string]any) string {
	if resp == nil {
		return ""
	}
	// OpenAPI v1.0 常见：shareUrl / share_id / share_url
	if s, ok := getString(resp, "shareUrl"); ok && s != "" {
		return s
	}
	if s, ok := getString(resp, "share_url"); ok && s != "" {
		return s
	}
	if sid, ok := getString(resp, "shareId"); ok && sid != "" {
		return "https://www.alipan.com/s/" + sid
	}
	if dm, ok := resp["data"].(map[string]any); ok {
		if s, ok := dm["shareUrl"].(string); ok && s != "" {
			return s
		}
		if s, ok := dm["share_url"].(string); ok && s != "" {
			return s
		}
		if s, ok := dm["url"].(string); ok && s != "" {
			return s
		}
		if sid, ok := dm["shareId"].(string); ok && sid != "" {
			return "https://www.alipan.com/s/" + sid
		}
		if sid, ok := dm["share_id"].(string); ok && sid != "" {
			return "https://www.alipan.com/s/" + sid
		}
	}
	return ""
}

func aliyunCreateShareLink(client *http.Client, access, driveID string, fileIDs []string) (string, error) {
	if len(fileIDs) == 0 {
		return "", fmt.Errorf("没有可用于创建分享的文件")
	}
	// 优先使用官方 OpenAPI：/adrive/v1.0/openFile/createShare（字段为 camelCase）
	// 文档示例见你提供的“创建文件分享”说明。
	openBody, _ := json.Marshal(map[string]any{
		"driveId":    driveID,
		"fileIdList": fileIDs,
		"expiration": "",
		"sharePwd":   "",
	})
	endpoint := "https://open.aliyundrive.com/adrive/v1.0/openFile/createShare"
	resp, err := httpDoJSONBearerAliyun(client, http.MethodPost, endpoint, access, "", openBody, "阿里云盘")
	if err != nil {
		// 兜底：老接口（部分账号/令牌来源可能仍需要）
		body, _ := json.Marshal(map[string]any{
			"drive_id":     driveID,
			"file_id_list": fileIDs,
			"share_pwd":    "",
			"expiration":   "",
		})
		resp, err = httpDoJSONBearerAliyun(client, http.MethodPost, "https://api.aliyundrive.com/adrive/v2/share_link/create", access, "", body, "阿里云盘")
		if err != nil {
			return "", err
		}
	}
	u := aliyunPickShareURLFromCreate(resp)
	if u == "" {
		return "", fmt.Errorf("创建分享成功但未解析到链接")
	}
	return u, nil
}

func aliyunWalkShareFiles(client *http.Client, access, shareID, shareToken, parentID string, depth int) ([]string, error) {
	if depth > 40 {
		return nil, fmt.Errorf("分享目录层级过深")
	}
	var ids []string
	marker := ""
	for {
		items, next, err := aliyunListShareFilesPage(client, access, shareToken, shareID, parentID, marker)
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			row, ok := it.(map[string]any)
			if !ok {
				continue
			}
			fid, _ := row["file_id"].(string)
			if fid == "" {
				continue
			}
			typ, _ := row["type"].(string)
			if typ == "folder" {
				sub, err := aliyunWalkShareFiles(client, access, shareID, shareToken, fid, depth+1)
				if err != nil {
					return nil, err
				}
				ids = append(ids, sub...)
			} else {
				ids = append(ids, fid)
			}
		}
		if next == "" {
			break
		}
		marker = next
	}
	return ids, nil
}

// AliyunSaveByShareLink 使用开放接口 refresh_token 将分享批量保存到指定目录
func AliyunSaveByShareLink(link string, sharePwdOverride string) (AliyunTransferResult, error) {
	cfg, err := LoadNetdiskCredentials()
	if err != nil {
		return AliyunTransferResult{}, err
	}
	picked := PickAliyunRefreshToken(cfg)
	refresh := picked.Cookie
	if refresh == "" {
		return AliyunTransferResult{}, fmt.Errorf("请先在「网盘凭证」填写阿里云 refresh_token（或多账号轮询列表）")
	}
	shareID, err := parseAliyunShareID(link)
	if err != nil {
		return AliyunTransferResult{}, err
	}
	pwd := strings.TrimSpace(sharePwdOverride)

	aliyunDebugf("开始 link=%s shareID=%s pwd_len=%d", strings.TrimSpace(link), shareID, len(pwd))

	// 若配置了 OpenList renewapi，则优先走 renewapi 续期（更贴近官方 OpenAPI 使用方式）；失败再回退 token/refresh。
	var access, refreshDriveID string
	if strings.TrimSpace(cfg.AliyunRenewAPIURL) != "" {
		if a, _, did, e := aliyunRefreshAccessTokenByRenewAPI(cfg.AliyunRenewAPIURL, refresh); e != nil {
			aliyunDebugf("renewapi 续期失败，回退 token/refresh：%v", e)
		} else {
			access = a
			refreshDriveID = did
		}
	}
	if strings.TrimSpace(access) == "" {
		var e error
		access, _, refreshDriveID, e = aliyunRefreshAccessToken(refresh)
		if e != nil {
			return AliyunTransferResult{}, fmt.Errorf("刷新 access_token 失败: %w", e)
		}
	}
	client := &http.Client{Timeout: 45 * time.Second}
	// 兼容 AlipanPan.php：优先使用 token/refresh 响应中的 drive_id，不再调用已返回 404 的 /v2/user/getDriveInfo。
	driveID := strings.TrimSpace(refreshDriveID)
	if strings.TrimSpace(driveID) == "" {
		driveID = "2008425230" // xinyue-search: 固定 drive_id 兜底
	}
	aliyunDebugf("driveID=%s", driveID)
	shareToken, err := aliyunGetShareToken(client, access, shareID, pwd)
	if err != nil {
		return AliyunTransferResult{}, fmt.Errorf("获取 share_token 失败: %w", err)
	}
	aliyunDebugf("shareToken_len=%d", len(strings.TrimSpace(shareToken)))
	toParent := effectiveAliyunParent(picked, cfg.AliyunTargetParentFileID)
	aliyunDebugf("toParent=%s", toParent)

	fileIDs, err := aliyunGetShareAnonFileInfos(client, access, shareID)
	if err != nil {
		return AliyunTransferResult{}, fmt.Errorf("获取分享文件列表失败: %w", err)
	}
	if len(fileIDs) == 0 {
		return AliyunTransferResult{}, fmt.Errorf("分享内没有可转存的文件")
	}
	aliyunDebugf("anon fileIDs=%d", len(fileIDs))

	newFileIDs, err := aliyunBatchCopyShareFiles(client, access, driveID, shareID, shareToken, toParent, fileIDs)
	if err != nil {
		return AliyunTransferResult{}, fmt.Errorf("批量转存失败: %w", err)
	}
	aliyunDebugf("batch copy newFileIDs=%d", len(newFileIDs))

	msg := "转存完成"
	out := AliyunTransferResult{ShareID: shareID, Message: msg}
	if cfg.ReplaceLinkAfterTransfer {
		if len(newFileIDs) > 0 {
			u, err := aliyunCreateShareLink(client, access, driveID, newFileIDs)
			if err != nil {
				out.Message = msg + "（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
			} else {
				out.OwnShareURL = u
			}
		}
	}
	return out, nil
}
