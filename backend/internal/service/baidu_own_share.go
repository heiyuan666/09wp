package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type baiduListEntry struct {
	fsid  int64
	mtime int64
	name  string
	isDir int
}

// baiduReplaceWithOwnShareLink 转存完成后：在目标目录按“转存源文件名”匹配 fs_id，并调用 share/set 创建本人分享
func baiduReplaceWithOwnShareLink(client *http.Client, cookie, bdstoken, targetPath, referer string, fileNames []string) (string, error) {
	const sharePwd = "6666"
	if strings.TrimSpace(targetPath) == "" {
		targetPath = "/"
	}
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = "/" + targetPath
	}
	time.Sleep(450 * time.Millisecond)

	entries, err := baiduListOwnDirEntries(client, cookie, bdstoken, targetPath, 2000)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("目标目录无文件，无法创建分享")
	}

	// fileNames 为空时，退化成取最新的若干文件
	nameSet := map[string]struct{}{}
	for _, n := range fileNames {
		n = strings.TrimSpace(n)
		if n != "" {
			nameSet[n] = struct{}{}
		}
	}
	want := len(nameSet)
	if want < 1 {
		want = 1
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].mtime > entries[j].mtime })
	var pick []int64
	for _, it := range entries {
		if it.fsid == 0 {
			continue
		}
		if len(nameSet) > 0 {
			if _, ok := nameSet[it.name]; !ok {
				continue
			}
		}
		pick = append(pick, it.fsid)
		if len(pick) >= want {
			break
		}
	}
	if len(pick) == 0 {
		return "", fmt.Errorf("未匹配到可分享的 fs_id")
	}

	link, err := baiduCreateShareSet(client, cookie, bdstoken, pick, sharePwd, referer)
	if err != nil {
		return "", err
	}
	if sharePwd != "" && !strings.Contains(link, "pwd=") {
		if strings.Contains(link, "?") {
			return link + "&pwd=" + sharePwd, nil
		}
		return link + "?pwd=" + sharePwd, nil
	}
	return link, nil
}

func baiduListOwnDirEntries(client *http.Client, cookie, bdstoken, dir string, limit int) ([]baiduListEntry, error) {
	if limit < 1 {
		limit = 100
	}
	if dir == "" {
		dir = "/"
	}
	q := url.Values{}
	q.Set("dir", dir)
	q.Set("bdstoken", bdstoken)
	q.Set("web", "1")
	q.Set("order", "time")
	q.Set("desc", "1")
	q.Set("page", "1")
	q.Set("num", fmt.Sprintf("%d", limit))
	q.Set("showempty", "0")
	q.Set("web", "1")
	q.Set("page", "1")
	q.Set("num", fmt.Sprintf("%d", limit))
	apiURL := "https://pan.baidu.com/api/list?" + q.Encode()
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", baiduUA)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", "https://pan.baidu.com/disk/home")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("列目录解析失败")
	}
	if errno := int(getFloat(result, "errno")); errno != 0 {
		msg, _ := result["errmsg"].(string)
		if msg == "" {
			msg = fmt.Sprintf("errno=%d", errno)
		}
		return nil, fmt.Errorf("列举网盘目录失败: %s", msg)
	}
	list, _ := result["list"].([]any)
	var out []baiduListEntry
	for _, it := range list {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		fsid := int64(getFloat(m, "fs_id"))
		name, _ := m["server_filename"].(string)
		isDir := int(getFloat(m, "isdir"))
		mtime := int64(getFloat(m, "server_mtime"))
		if fsid == 0 {
			continue
		}
		out = append(out, baiduListEntry{fsid: fsid, mtime: mtime, name: name, isDir: isDir})
	}
	return out, nil
}

func baiduCreateShareSet(client *http.Client, cookie, bdstoken string, fsids []int64, password, referer string) (string, error) {
	if len(fsids) == 0 {
		return "", fmt.Errorf("fsids 为空")
	}
	joined := make([]string, 0, len(fsids))
	for _, id := range fsids {
		if id != 0 {
			joined = append(joined, fmt.Sprintf("%d", id))
		}
	}
	fidStr := strings.Join(joined, ",")

	params := url.Values{}
	params.Set("period", "0")
	params.Set("pwd", password)
	params.Set("eflag_disable", "true")
	params.Set("channel_list", "[]")
	params.Set("schannel", "4")
	// xinyue-search：fid_list => '['.$fsId.']'，其中 $fsId 是 "1,2,3"
	params.Set("fid_list", "["+fidStr+"]")

	createURL := fmt.Sprintf(
		"https://pan.baidu.com/share/set?channel=chunlei&bdstoken=%s&clienttype=0&app_id=%s&web=1",
		url.QueryEscape(bdstoken),
		url.QueryEscape(baiduAppID),
	)

	req, err := http.NewRequest(http.MethodPost, createURL, strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", baiduUA)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Cookie", cookie)
	if referer != "" {
		req.Header.Set("Referer", referer)
	} else {
		req.Header.Set("Referer", "https://pan.baidu.com/disk/home")
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var tr map[string]any
	if err := json.Unmarshal(raw, &tr); err != nil {
		return "", fmt.Errorf("创建分享解析失败: %s", string(raw))
	}
	if errno := int(getFloat(tr, "errno")); errno != 0 {
		msg, _ := tr["show_msg"].(string)
		if msg == "" {
			msg, _ = tr["errmsg"].(string)
		}
		if msg == "" {
			msg = fmt.Sprintf("errno=%d", errno)
		}
		return "", fmt.Errorf("创建分享失败: %s", msg)
	}
	if link, ok := tr["link"].(string); ok && link != "" {
		return link, nil
	}
	if s, ok := tr["shorturl"].(string); ok && s != "" {
		if strings.HasPrefix(s, "http") {
			return s, nil
		}
		return "https://pan.baidu.com/s/" + s, nil
	}
	if su, ok := tr["surl"].(string); ok && su != "" {
		return "https://pan.baidu.com/s/" + su, nil
	}
	return "", fmt.Errorf("创建分享成功但未解析到链接")
}
