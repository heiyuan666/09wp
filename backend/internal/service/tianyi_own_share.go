package service

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func tianyiTryOwnShareLink(client *http.Client, cookie, ua, referer, folderID string, want int) (string, error) {
	if want < 1 {
		want = 1
	}
	time.Sleep(900 * time.Millisecond)
	q := url.Values{}
	q.Set("folderId", folderID)
	q.Set("pageNum", "1")
	q.Set("pageSize", "100")
	q.Set("iconOption", "5")
	q.Set("orderBy", "lastOpTime")
	q.Set("descending", "true")
	q.Set("mediaType", "0")
	listURL := "https://cloud.189.cn/api/open/file/listFiles.action?" + q.Encode()
	raw, err := tianyiGET(client, listURL, cookie, referer, ua)
	if err != nil {
		return "", err
	}
	if rc := int(getFloat(raw, "res_code")); rc != 0 {
		msg, _ := raw["res_message"].(string)
		if msg == "" {
			msg = "列目录失败"
		}
		return "", fmt.Errorf("%s", msg)
	}
	fileListAO, ok := raw["fileListAO"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("天翼目录格式异常")
	}
	files, _ := fileListAO["fileList"].([]any)
	var fileIDs []string
	for _, f := range files {
		m, ok := f.(map[string]any)
		if !ok {
			continue
		}
		isFolder := false
		if v, ok := m["isFolder"].(bool); ok {
			isFolder = v
		} else if int(getFloat(m, "isFolder")) == 1 {
			isFolder = true
		}
		if isFolder {
			continue
		}
		id := tianyiStringFromAny(m["id"])
		if id == "" {
			id = tianyiStringFromAny(m["fileId"])
		}
		if id != "" {
			fileIDs = append(fileIDs, id)
		}
		if len(fileIDs) >= want {
			break
		}
	}
	if len(fileIDs) == 0 {
		return "", fmt.Errorf("目标目录中未找到可分享的文件")
	}
	return tianyiPostCreateShareLink(client, cookie, ua, referer, fileIDs)
}

func tianyiPostCreateShareLink(client *http.Client, cookie, ua, referer string, fileIDs []string) (string, error) {
	form := url.Values{}
	form.Set("fileId", strings.Join(fileIDs, ","))
	form.Set("shareMode", "2")
	form.Set("expiresAt", "")
	form.Set("password", "")
	form.Set("desc", "")
	u := "https://cloud.189.cn/api/open/share/createShareLink.action"
	raw, err := tianyiPOSTForm(client, u, cookie, form.Encode(), referer, ua)
	if err != nil {
		return "", err
	}
	if rc := int(getFloat(raw, "res_code")); rc != 0 {
		msg, _ := raw["res_message"].(string)
		if msg == "" {
			msg = "创建分享失败"
		}
		return "", fmt.Errorf("%s", msg)
	}
	if link, ok := raw["shortShareUrl"].(string); ok && link != "" {
		return link, nil
	}
	if link, ok := raw["longShareUrl"].(string); ok && link != "" {
		return link, nil
	}
	if link, ok := raw["shareUrl"].(string); ok && link != "" {
		return link, nil
	}
	return "", fmt.Errorf("未解析到天翼分享链接")
}
