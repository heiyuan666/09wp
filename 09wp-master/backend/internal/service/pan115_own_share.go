package service

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func pan115ReplaceWithCookieShare(client *http.Client, cookie, targetCid string, wantFileCount int) (string, error) {
	if wantFileCount < 1 {
		wantFileCount = 1
	}
	uid, err := pan115GetUserID(client, cookie)
	if err != nil {
		return "", err
	}
	time.Sleep(400 * time.Millisecond)
	fileIDs, err := pan115ListRecentFileIDs(client, cookie, targetCid, wantFileCount+8)
	if err != nil {
		return "", err
	}
	if len(fileIDs) == 0 {
		return "", fmt.Errorf("目标目录下列不到文件")
	}
	if len(fileIDs) > wantFileCount {
		fileIDs = fileIDs[:wantFileCount]
	}
	return pan115ShareAdd(client, cookie, uid, fileIDs)
}

func pan115GetUserID(client *http.Client, cookie string) (string, error) {
	if u := pan115UIDFromCookie(cookie); u != "" {
		return u, nil
	}
	endpoints := []string{
		"https://webapi.115.com/files?aid=1&cid=0&offset=0&limit=1&show_dir=0",
		"https://webapi.115.com/user/info",
	}
	for _, ep := range endpoints {
		raw, err := http115GET(client, ep, cookie)
		if err != nil {
			continue
		}
		if u := pan115PickUserID(raw); u != "" {
			return u, nil
		}
	}
	return "", fmt.Errorf("未获取 115 user_id")
}

func pan115UIDFromCookie(cookie string) string {
	for _, p := range strings.Split(cookie, ";") {
		p = strings.TrimSpace(p)
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		if strings.EqualFold(kv[0], "uid") {
			return strings.TrimSpace(kv[1])
		}
	}
	return ""
}

func pan115PickUserID(m map[string]any) string {
	if id := stringFromAny(m["user_id"]); id != "" {
		return id
	}
	if dm, ok := m["data"].(map[string]any); ok {
		if id := stringFromAny(dm["uid"]); id != "" {
			return id
		}
		if id := stringFromAny(dm["user_id"]); id != "" {
			return id
		}
	}
	return ""
}

func pan115ListRecentFileIDs(client *http.Client, cookie, cid string, limit int) ([]string, error) {
	if limit < 1 {
		limit = 20
	}
	if cid == "" {
		cid = "0"
	}
	u := fmt.Sprintf("https://webapi.115.com/files?aid=1&cid=%s&offset=0&show_dir=0&limit=%d&o=user_ptime&asc=0",
		url.QueryEscape(cid), limit)
	raw, err := http115GET(client, u, cookie)
	if err != nil {
		return nil, err
	}
	state, _ := raw["state"].(bool)
	if !state {
		e, _ := raw["error"].(string)
		if e == "" {
			e = "列目录失败"
		}
		return nil, fmt.Errorf("%s", e)
	}
	data := pan115ExtractDataArray(raw)
	var out []string
	for _, it := range data {
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		if v, ok := row["isfolder"].(bool); ok && v {
			continue
		}
		fid := stringFromAny(row["fid"])
		if fid == "" {
			fid = stringFromAny(row["file_id"])
		}
		if fid != "" {
			out = append(out, fid)
		}
	}
	return out, nil
}

func pan115ExtractDataArray(raw map[string]any) []any {
	if arr, ok := raw["data"].([]any); ok {
		return arr
	}
	if dm, ok := raw["data"].(map[string]any); ok {
		if arr, ok := dm["list"].([]any); ok {
			return arr
		}
	}
	return nil
}

func pan115ShareAdd(client *http.Client, cookie, userID string, fileIDs []string) (string, error) {
	if len(fileIDs) == 0 {
		return "", fmt.Errorf("file_id 为空")
	}
	form := url.Values{}
	form.Set("user_id", userID)
	form.Set("file_id", strings.Join(fileIDs, ","))
	form.Set("share_channel", "share_web")
	form.Set("share_init", "1")
	form.Set("share_mode", "1")
	form.Set("ignore_same_file", "1")
	api := "https://webapi.115.com/share/add"
	extra := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36",
		"Referer":    "https://115.com/",
	}
	resp, err := httpDoFormPost(client, api, cookie, form.Encode(), extra)
	if err != nil {
		return "", err
	}
	ok, _ := resp["state"].(bool)
	if !ok {
		e, _ := resp["error"].(string)
		if e == "" {
			e = "创建分享失败"
		}
		return "", fmt.Errorf("%s", e)
	}
	if data, ok := resp["data"].(map[string]any); ok {
		if link, ok := data["share_url"].(string); ok && link != "" {
			return link, nil
		}
		if code, ok := data["share_code"].(string); ok && code != "" {
			return "https://115.com/s/" + code, nil
		}
	}
	return "", fmt.Errorf("未解析到 115 分享链接")
}
