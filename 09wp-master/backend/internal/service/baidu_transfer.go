package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const baiduAppID = "250528"
const baiduUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type BaiduTransferResult struct {
	ShareID     string `json:"share_id"`
	Title       string `json:"title,omitempty"`
	Message     string `json:"message"`
	OwnShareURL string `json:"own_share_url,omitempty"` // 转存后本人分享链接（replace_link_after_transfer 且平台支持时）
}

type baiduShareEntry struct {
	fsid            int64
	serverFilename string
	isDir           int
}

var (
	reBdstoken = regexp.MustCompile(`"bdstoken"\s*:\s*"([^"]*)"`)
	reShareid  = regexp.MustCompile(`"shareid"\s*:\s*"?(\d+)"?`)
	reShareUK  = regexp.MustCompile(`"share_uk"\s*:\s*"?(\d+)"?`)
	reUK       = regexp.MustCompile(`(?m)"uk"\s*:\s*"?(\d+)"?`)
)

// BaiduSaveByShareLink 使用登录后的百度网盘 Cookie 将分享转存到指定目录
func BaiduSaveByShareLink(link string, passOverride string) (BaiduTransferResult, error) {
	cred, err := LoadNetdiskCredentials()
	if err != nil {
		return BaiduTransferResult{}, err
	}
	picked := PickBaiduCookie(cred)
	cookie := picked.Cookie
	if cookie == "" {
		return BaiduTransferResult{}, fmt.Errorf("请先在「网盘凭证」页面填写百度 Cookie（或多账号轮询列表）")
	}
	targetPath := effectiveBaiduPath(picked, cred.BaiduTargetPath)
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = "/" + targetPath
	}

	norm, err := baiduNormalizeURL(link)
	if err != nil {
		return BaiduTransferResult{}, err
	}
	surl := baiduExtractSur(norm)
	if surl == "" {
		return BaiduTransferResult{}, fmt.Errorf("无效的百度分享链接")
	}
	shorturl := baiduShorturl(surl)

	pwd := strings.TrimSpace(passOverride)
	if pwd == "" {
		if u, perr := url.Parse(norm); perr == nil {
			pwd = strings.TrimSpace(u.Query().Get("pwd"))
		}
	}

	client := &http.Client{Timeout: 45 * time.Second}

	// 1) 分享页 HTML，解析 bdstoken / shareid / share_uk
	pageURL := fmt.Sprintf("https://pan.baidu.com/s/%s", surl)
	body, err := baiduHTTPGet(client, pageURL, cookie, norm)
	if err != nil {
		return BaiduTransferResult{}, err
	}
	bdstoken, shareid, shareUK := baiduParseTokens(string(body))
	if bdstoken == "" || shareid == "" || shareUK == "" {
		initURL := fmt.Sprintf("https://pan.baidu.com/share/init?surl=%s&time=%d", shorturl, time.Now().UnixMilli())
		body2, err2 := baiduHTTPGet(client, initURL, cookie, norm)
		if err2 == nil {
			bdstoken, shareid, shareUK = baiduParseTokens(string(body2))
		}
	}
	if bdstoken == "" || shareid == "" || shareUK == "" {
		return BaiduTransferResult{}, fmt.Errorf("无法解析分享页参数，请确认 Cookie 已登录且链接有效")
	}

	var randsk string
	if pwd != "" {
		randsk, err = baiduVerifyPwd(client, norm, shorturl, pwd)
		if err != nil {
			return BaiduTransferResult{}, err
		}
	}

	listCookie := cookie
	if randsk != "" {
		listCookie = baiduMergeCookie(cookie, "BDCLND", randsk)
	}

	// 2) 列举分享根目录文件 fs_id
	shareItems, err := baiduListAllFsids(client, shorturl, norm, listCookie)
	if err != nil {
		return BaiduTransferResult{}, err
	}
	if len(shareItems) == 0 {
		return BaiduTransferResult{}, fmt.Errorf("分享目录为空或无法列出文件")
	}
	var fsids []int64
	var fileNames []string
	for _, it := range shareItems {
		if it.fsid != 0 {
			fsids = append(fsids, it.fsid)
		}
		if strings.TrimSpace(it.serverFilename) != "" {
			fileNames = append(fileNames, it.serverFilename)
		}
	}
	if len(fsids) == 0 {
		return BaiduTransferResult{}, fmt.Errorf("分享目录fs_id为空")
	}

	// 3) 转存
	transferCookie := cookie
	if randsk != "" {
		transferCookie = listCookie
	}
	sekeyPart := ""
	if randsk != "" {
		sekeyPart = "&sekey=" + url.QueryEscape(randsk)
	}
	transferBase := fmt.Sprintf(
		"https://pan.baidu.com/share/transfer?shareid=%s&from=%s&bdstoken=%s&app_id=%s&channel=chunlei&web=1&clienttype=0%s",
		url.QueryEscape(shareid), url.QueryEscape(shareUK), url.QueryEscape(bdstoken), baiduAppID, sekeyPart,
	)
	fsidJSON, _ := json.Marshal(fsids)
	form := url.Values{}
	form.Set("fsidlist", string(fsidJSON))
	form.Set("path", targetPath)
	req, err := http.NewRequest(http.MethodPost, transferBase, strings.NewReader(form.Encode()))
	if err != nil {
		return BaiduTransferResult{}, err
	}
	req.Header.Set("User-Agent", baiduUA)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Referer", norm)
	req.Header.Set("Cookie", transferCookie)

	resp, err := client.Do(req)
	if err != nil {
		return BaiduTransferResult{}, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var tr map[string]any
	if err := json.Unmarshal(raw, &tr); err != nil {
		return BaiduTransferResult{}, fmt.Errorf("转存返回解析失败: %s", string(raw))
	}
	errno := int(getFloat(tr, "errno"))
	if errno != 0 {
		msg, _ := tr["show_msg"].(string)
		if msg == "" {
			msg, _ = tr["errmsg"].(string)
		}
		if msg == "" {
			msg = fmt.Sprintf("errno=%d", errno)
		}
		return BaiduTransferResult{}, fmt.Errorf("转存失败: %s", msg)
	}
	title := ""
	if len(fileNames) > 0 {
		title = strings.TrimSpace(fileNames[0])
	}
	out := BaiduTransferResult{ShareID: shareid, Title: title, Message: "转存成功"}
	if cred.ReplaceLinkAfterTransfer {
		u, err := baiduReplaceWithOwnShareLink(client, transferCookie, bdstoken, targetPath, norm, fileNames)
		if err != nil {
			out.Message = "转存成功（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
		} else {
			out.OwnShareURL = u
		}
	}
	return out, nil
}

func baiduNormalizeURL(link string) (string, error) {
	cleaned := strings.TrimSpace(link)
	startIdx := strings.Index(cleaned, "https://pan.baidu.com/s/")
	if startIdx == -1 {
		startIdx = strings.Index(cleaned, "http://pan.baidu.com/s/")
	}
	if startIdx == -1 {
		return "", fmt.Errorf("未找到有效的百度网盘分享 URL")
	}
	endIdx := startIdx
	for endIdx < len(cleaned) {
		ch := cleaned[endIdx]
		if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
			break
		}
		remaining := cleaned[endIdx:]
		if strings.HasPrefix(remaining, "提取码") || strings.HasPrefix(remaining, "密码") {
			break
		}
		endIdx++
	}
	urlStr := strings.TrimSpace(cleaned[startIdx:endIdx])
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	normalized := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
	if parsedURL.RawQuery != "" {
		normalized += "?" + parsedURL.Query().Encode()
	}
	if parsedURL.Fragment != "" {
		normalized += "#" + parsedURL.Fragment
	}
	return normalized, nil
}

func baiduExtractSur(shareURL string) string {
	parsedURL, err := url.Parse(shareURL)
	if err != nil {
		return ""
	}
	path := parsedURL.Path
	if strings.HasPrefix(path, "/s/") {
		s := strings.TrimPrefix(path, "/s/")
		if idx := strings.Index(s, "?"); idx != -1 {
			s = s[:idx]
		}
		return s
	}
	if strings.HasPrefix(path, "/share/init") {
		return parsedURL.Query().Get("surl")
	}
	return ""
}

func baiduShorturl(surl string) string {
	if len(surl) > 1 {
		return surl[1:]
	}
	return surl
}

func baiduParseTokens(html string) (bdstoken, shareid, shareUK string) {
	if m := reBdstoken.FindStringSubmatch(html); len(m) > 1 {
		bdstoken = m[1]
	}
	if m := reShareid.FindStringSubmatch(html); len(m) > 1 {
		shareid = m[1]
	}
	if m := reShareUK.FindStringSubmatch(html); len(m) > 1 {
		shareUK = m[1]
	}
	if shareUK == "" {
		if m := reUK.FindStringSubmatch(html); len(m) > 1 {
			shareUK = m[1]
		}
	}
	return bdstoken, shareid, shareUK
}

func baiduHTTPGet(client *http.Client, reqURL, cookie, referer string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", baiduUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Cookie", cookie)
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func baiduVerifyPwd(client *http.Client, shareURL, shorturl, pwd string) (string, error) {
	apiURL := fmt.Sprintf("https://pan.baidu.com/share/verify?surl=%s&pwd=%s&t=%d",
		url.QueryEscape(shorturl), url.QueryEscape(pwd), rand.Int63())
	form := url.Values{}
	form.Set("pwd", pwd)
	form.Set("vcode", "")
	form.Set("vcode_str", "")
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", baiduUA)
	req.Header.Set("Referer", shareURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("验证提取码响应解析失败")
	}
	if errno := int(getFloat(result, "errno")); errno != 0 {
		msg, _ := result["errmsg"].(string)
		if msg == "" {
			msg = "提取码验证失败"
		}
		return "", fmt.Errorf("%s", msg)
	}
	randsk, _ := result["randsk"].(string)
	if randsk == "" {
		return "", fmt.Errorf("验证成功但未返回 randsk")
	}
	return randsk, nil
}

func baiduListAllFsids(client *http.Client, shorturl, referer, cookie string) ([]baiduShareEntry, error) {
	var all []baiduShareEntry
	page := 1
	for page < 50 {
		apiURL := fmt.Sprintf(
			"https://pan.baidu.com/share/list?web=5&app_id=%s&desc=1&showempty=0&page=%d&num=100&order=time&shorturl=%s&root=1&view_mode=1&channel=chunlei&web=1&clienttype=0&_=%d",
			baiduAppID, page, url.QueryEscape(shorturl), time.Now().UnixMilli(),
		)
		req, err := http.NewRequest(http.MethodGet, apiURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", baiduUA)
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Cookie", cookie)
		req.Header.Set("Referer", referer)
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var result map[string]any
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("分享列表解析失败")
		}
		if errno := int(getFloat(result, "errno")); errno != 0 {
			msg, _ := result["errmsg"].(string)
			if msg == "" {
				msg = fmt.Sprintf("errno=%d", errno)
			}
			return nil, fmt.Errorf("列举分享文件失败: %s", msg)
		}
		list, _ := result["list"].([]any)
		for _, it := range list {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			fsid := int64(getFloat(m, "fs_id"))
			if fsid == 0 {
				continue
			}
			name, _ := m["server_filename"].(string)
			isDir := int(getFloat(m, "isdir"))
			all = append(all, baiduShareEntry{
				fsid:            fsid,
				serverFilename: name,
				isDir:           isDir,
			})
		}
		if len(list) < 100 {
			break
		}
		page++
	}
	return all, nil
}

func baiduMergeCookie(base, name, value string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return name + "=" + value
	}
	if strings.Contains(strings.ToLower(base), strings.ToLower(name)+"=") {
		return base
	}
	return base + "; " + name + "=" + value
}
