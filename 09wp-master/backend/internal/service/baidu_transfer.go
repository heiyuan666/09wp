package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func baiduTransferDebugEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("NETDISK_TRANSFER_DEBUG")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func baiduDebugf(format string, args ...any) {
	if !baiduTransferDebugEnabled() {
		return
	}
	log.Printf("[BAIDU-TRANSFER] "+format, args...)
}

func baiduCookieHasName(cookie, name string) bool {
	if cookie == "" || name == "" {
		return false
	}
	needle := strings.ToLower(name) + "="
	s := strings.ToLower(strings.TrimLeft(cookie, " "))
	if strings.HasPrefix(s, needle) {
		return true
	}
	return strings.Contains(s, "; "+needle)
}

// baiduDebugSnippet 截断字符串便于打日志（避免刷屏）
func baiduDebugSnippet(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…(trunc)"
}

// baiduSekeyForTransferQuery 拼接 share/transfer 的 sekey。verify 返回的 randsk 常常是已百分号编码的字符串，
// 若再 url.QueryEscape 一次会把「%」编成「%25」，百度会判为提取码/sekey 错误。
func baiduSekeyForTransferQuery(randsk string) string {
	r := strings.TrimSpace(randsk)
	if r == "" {
		return ""
	}
	dec, err := url.QueryUnescape(r)
	if err != nil || dec == "" {
		dec = r
	}
	return url.QueryEscape(dec)
}

const baiduAppID = "250528"
const baiduUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

type BaiduTransferResult struct {
	ShareID     string `json:"share_id"`
	Title       string `json:"title,omitempty"`
	Message     string `json:"message"`
	OwnShareURL string `json:"own_share_url,omitempty"` // 转存后本人分享链接（replace_link_after_transfer 且平台支持时）
}

type baiduShareEntry struct {
	fsid           int64
	serverFilename string
	isDir          int
}

var (
	reBdstoken = regexp.MustCompile(`"bdstoken"\s*:\s*"([^"]*)"`)
	reShareid  = regexp.MustCompile(`"shareid"\s*:\s*"?(\d+)"?`)
	reShareUK  = regexp.MustCompile(`"share_uk"\s*:\s*"?(\d+)"?`)
	reUK       = regexp.MustCompile(`(?m)"uk"\s*:\s*"?(\d+)"?`)

	// 对齐 xinyue-search BaiduWork.php parseResponse：从分享页源码抓取关键参数
	rePageShareID  = regexp.MustCompile(`"shareid":(\d+?),`)
	rePageShareUK  = regexp.MustCompile(`"share_uk":"(\d+?)",`)
	rePageFsID     = regexp.MustCompile(`"fs_id":(\d+?),`)
	rePageFileName = regexp.MustCompile(`"server_filename":"(.+?)",`)
	rePageIsDir    = regexp.MustCompile(`"isdir":(\d+?),`)
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
	shareMainURL := fmt.Sprintf("https://pan.baidu.com/s/%s", surl)
	shorturlCandidates := baiduShorturlCandidates(shareMainURL)

	baiduDebugf("开始 link_norm=%s surl=%s candidates=%v passOverride_len=%d cookie_has_BDUSS=%v cookie_has_BDCLND=%v",
		norm, surl, shorturlCandidates, len(strings.TrimSpace(passOverride)), baiduCookieHasName(cookie, "BDUSS"), baiduCookieHasName(cookie, "BDCLND"))

	pwd := strings.TrimSpace(passOverride)
	if pwd == "" {
		if u, perr := url.Parse(norm); perr == nil {
			pwd = strings.TrimSpace(u.Query().Get("pwd"))
		}
	}
	baiduDebugf("提取码解析后 pwd_len=%d（0 表示未带提取码，将跳过 verify）", len(pwd))

	client := &http.Client{Timeout: 45 * time.Second}

	// 1) 分享页 HTML，解析 bdstoken / shareid / share_uk
	pageURL := shareMainURL
	body, err := baiduHTTPGet(client, pageURL, cookie, norm)
	if err != nil {
		return BaiduTransferResult{}, err
	}
	baiduDebugf("分享页 GET 完成 url=%s body_len=%d", pageURL, len(body))
	bdstoken, shareid, shareUK := baiduParseTokens(string(body))
	baiduDebugf("首屏解析 token 非空: bdstoken=%v shareid=%v share_uk=%v", bdstoken != "", shareid != "", shareUK != "")
	if bdstoken == "" || shareid == "" || shareUK == "" {
		for _, su := range shorturlCandidates {
			initURL := fmt.Sprintf("https://pan.baidu.com/share/init?surl=%s&time=%d", su, time.Now().UnixMilli())
			body2, err2 := baiduHTTPGet(client, initURL, cookie, norm)
			if err2 == nil {
				bdstoken, shareid, shareUK = baiduParseTokens(string(body2))
			}
			if bdstoken != "" && shareid != "" && shareUK != "" {
				break
			}
		}
	}
	if bdstoken == "" || shareid == "" || shareUK == "" {
		return BaiduTransferResult{}, fmt.Errorf("无法解析分享页参数，请确认 Cookie 已登录且链接有效")
	}

	var randsk string
	// 如果 Cookie 已包含 BDCLND（通常来自浏览器手动验证提取码/验证码），可跳过 share/verify，直接用该 Cookie 访问分享。
	// 这能规避部分分享在接口层面返回“提取码验证失败”的风控场景。
	// 即使 Cookie 已带 BDCLND，也优先使用当前提取码重新 verify 一次。
	// 旧 BDCLND 可能来自其它分享，继续沿用会导致 share/list 返回 errno=-9。
	if pwd != "" {
		var verifyErr error
		for _, su := range shorturlCandidates {
			baiduDebugf("share/verify 尝试 shorturl=%s", su)
			randsk, err = baiduVerifyPwd(client, norm, su, pwd, bdstoken)
			if err == nil {
				verifyErr = nil
				baiduDebugf("share/verify 成功 shorturl=%s randsk_len=%d", su, len(randsk))
				break
			}
			verifyErr = err
			baiduDebugf("share/verify 失败 shorturl=%s err=%v", su, err)
		}
		if verifyErr != nil {
			return BaiduTransferResult{}, verifyErr
		}
	} else {
		baiduDebugf("未提供提取码：跳过 share/verify（若 share/list 报 -9，请在资源链接带 ?pwd= 或在后台填提取码）")
	}

	listCookie := cookie
	if randsk != "" {
		hadCLND := baiduCookieHasName(cookie, "BDCLND")
		listCookie = baiduMergeCookie(cookie, "BDCLND", randsk)
		baiduDebugf("合并 BDCLND: had_BDCLND_before=%v after_merge_still_has=%v randsk_len=%d cookie_len %d->%d",
			hadCLND, baiduCookieHasName(listCookie, "BDCLND"), len(randsk), len(cookie), len(listCookie))
	}

	// 2) 登录态访问分享页 -> 解析 fs_id（优先策略，降低对 share/list 的依赖）
	// 对齐 hxz393/BaiduPanFilesTransfers：verify 成功后更新 BDCLND，再次请求分享页源码解析参数。
	var fsids []int64
	var fileNames []string
	if page2, e2 := baiduHTTPGet(client, pageURL, listCookie, norm); e2 == nil {
		baiduDebugf("二次分享页 GET(带 listCookie) body_len=%d", len(page2))
		fsids, fileNames = baiduExtractFromSharePage(string(page2))
		baiduDebugf("HTML 解析 fs_id 数量=%d fileNames 数量=%d", len(fsids), len(fileNames))
	} else {
		baiduDebugf("二次分享页 GET 失败: %v", e2)
	}

	// 2.1) 回退：解析不到 fs_id 时，再调用 share/list（部分页面结构变化时用）
	if len(fsids) == 0 {
		baiduDebugf("HTML 未解析到 fs_id，fallback share/list")
		var (
			shareItems []baiduShareEntry
			listErr    error
		)
		for _, su := range shorturlCandidates {
			baiduDebugf("share/list 尝试 shorturl=%s", su)
			shareItems, err = baiduListAllFsids(client, su, norm, listCookie)
			if err == nil {
				listErr = nil
				baiduDebugf("share/list 成功 shorturl=%s entries=%d", su, len(shareItems))
				break
			}
			listErr = err
			baiduDebugf("share/list 失败 shorturl=%s err=%v", su, err)
		}
		if listErr != nil {
			return BaiduTransferResult{}, listErr
		}
		for _, it := range shareItems {
			if it.fsid != 0 {
				fsids = append(fsids, it.fsid)
			}
			if strings.TrimSpace(it.serverFilename) != "" {
				fileNames = append(fileNames, it.serverFilename)
			}
		}
	}

	if len(fsids) == 0 {
		return BaiduTransferResult{}, fmt.Errorf("分享目录为空或无法列出文件")
	}

	// 3) 转存
	transferCookie := cookie
	if randsk != "" {
		transferCookie = listCookie
	}
	sekeyPart := ""
	if randsk != "" {
		sekeyPart = "&sekey=" + baiduSekeyForTransferQuery(randsk)
	}
	baiduDebugf("share/transfer 请求 fsid_count=%d sekey_once_encoded=%v", len(fsids), strings.Contains(strings.TrimSpace(randsk), "%"))
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
		baiduDebugf("share/transfer 失败 errno=%d msg=%q body=%s", errno, msg, baiduDebugSnippet(string(raw), 500))
		return BaiduTransferResult{}, fmt.Errorf("转存失败: %s", msg)
	}
	baiduDebugf("share/transfer 成功")
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

// baiduExtractFromSharePage 兜底从分享页源码解析 fs_id / 文件名（对齐 xinyue-search BaiduWork.php parseResponse）
func baiduExtractFromSharePage(html string) (fsids []int64, fileNames []string) {
	mFs := rePageFsID.FindAllStringSubmatch(html, -1)
	if len(mFs) > 0 {
		seen := make(map[int64]struct{}, len(mFs))
		for _, m := range mFs {
			if len(m) < 2 {
				continue
			}
			id, _ := strconv.ParseInt(m[1], 10, 64)
			if id == 0 {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			fsids = append(fsids, id)
		}
	}
	mNames := rePageFileName.FindAllStringSubmatch(html, -1)
	if len(mNames) > 0 {
		seen := make(map[string]struct{}, len(mNames))
		for _, m := range mNames {
			if len(m) < 2 {
				continue
			}
			n := strings.TrimSpace(m[1])
			if n == "" {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			fileNames = append(fileNames, n)
		}
	}
	return fsids, fileNames
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

func baiduShorturl(shareMainURL string) string {
	c := baiduShorturlCandidates(shareMainURL)
	if len(c) == 0 {
		return strings.TrimSpace(shareMainURL)
	}
	return c[0]
}

func baiduShorturlCandidates(shareMainURL string) []string {
	u := strings.TrimSpace(shareMainURL)
	if u == "" {
		return nil
	}
	code := ""
	if p, err := url.Parse(u); err == nil {
		path := strings.TrimSpace(p.Path)
		if strings.HasPrefix(path, "/s/") {
			code = strings.TrimSpace(strings.TrimPrefix(path, "/s/"))
		}
	}
	if code == "" {
		if idx := strings.Index(u, "/s/"); idx >= 0 {
			code = strings.TrimSpace(u[idx+3:])
		}
	}
	if code == "" {
		return []string{u}
	}
	code = strings.Trim(code, "/")
	out := []string{code}
	// 某些接口要求不带前导 "1" 的 surl，做兼容回退。
	if strings.HasPrefix(code, "1") && len(code) > 1 {
		out = append(out, code[1:])
	}
	return out
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

func baiduVerifyPwd(client *http.Client, shareURL, shorturl, pwd, bdstoken string) (string, error) {
	// 对齐 xinyue-search 的 BaiduWork.php verifyPassCode：
	// /share/verify?surl=<surl>&bdstoken=<bdstoken>&t=<ts>&channel=chunlei&web=1&clienttype=0
	params := url.Values{}
	params.Set("surl", shorturl)
	params.Set("bdstoken", bdstoken)
	params.Set("t", fmt.Sprint(time.Now().UnixMilli()))
	params.Set("channel", "chunlei")
	params.Set("web", "1")
	params.Set("clienttype", "0")
	apiURL := "https://pan.baidu.com/share/verify?" + params.Encode()

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
		baiduDebugf("share/verify 响应非 JSON: %s", baiduDebugSnippet(string(raw), 400))
		return "", fmt.Errorf("验证提取码响应解析失败")
	}
	errno0 := int(getFloat(result, "errno"))
	baiduDebugf("share/verify 响应 errno=%d body_snip=%s", errno0, baiduDebugSnippet(string(raw), 500))
	if errno0 != 0 {
		msg, _ := result["errmsg"].(string)
		if msg == "" {
			msg = "提取码验证失败"
		}
		// 常见 errno 解释（参考 xinyue-search 的错误码映射）
		switch errno0 {
		case -9, -12:
			return "", fmt.Errorf("%s（errno=%d）：请确认提取码正确", msg, errno0)
		case -62:
			return "", fmt.Errorf("%s（errno=%d）：链接访问次数过多/触发风控，请稍后再试或手动转存", msg, errno0)
		case -6:
			return "", fmt.Errorf("%s（errno=%d）：建议用浏览器无痕模式重新获取 Cookie，并完成可能出现的验证码（Cookie 需包含 BDUSS/BDCLND）", msg, errno0)
		default:
			return "", fmt.Errorf("%s（errno=%d）：若该分享页需要验证码/密享校验，请先在浏览器打开分享完成验证，然后把 Cookie 中的 BDCLND 一并填写到系统百度 Cookie 里", msg, errno0)
		}
	}
	randsk, _ := result["randsk"].(string)
	// 优先使用响应 Set-Cookie 下发的 BDCLND，兼容 randsk 与 cookie 编码不一致导致的 errno=-9。
	for _, ck := range resp.Cookies() {
		if strings.EqualFold(strings.TrimSpace(ck.Name), "BDCLND") && strings.TrimSpace(ck.Value) != "" {
			return strings.TrimSpace(ck.Value), nil
		}
	}
	if randsk == "" {
		return "", fmt.Errorf("验证成功但未返回 BDCLND/randsk")
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
			baiduDebugf("share/list 解析 JSON 失败 page=%d shorturl=%s raw=%s", page, shorturl, baiduDebugSnippet(string(body), 400))
			return nil, fmt.Errorf("分享列表解析失败")
		}
		if errno := int(getFloat(result, "errno")); errno != 0 {
			msg, _ := result["errmsg"].(string)
			if msg == "" && errno == -9 {
				msg = "提取码未验证或已失效（errno=-9）"
			}
			if msg == "" {
				msg = fmt.Sprintf("errno=%d", errno)
			}
			baiduDebugf("share/list 业务错误 page=%d shorturl=%s errno=%d errmsg=%q raw=%s",
				page, shorturl, errno, msg, baiduDebugSnippet(string(body), 600))
			return nil, fmt.Errorf("列举分享文件失败: %s", msg)
		}
		list, _ := result["list"].([]any)
		baiduDebugf("share/list page=%d shorturl=%s list_len=%d", page, shorturl, len(list))
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
				fsid:           fsid,
				serverFilename: name,
				isDir:          isDir,
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
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "" {
		return strings.TrimSpace(base)
	}
	base = strings.TrimSpace(base)
	if base == "" {
		return name + "=" + value
	}
	type pair struct{ k, v string }
	var pairs []pair
	seen := map[string]int{} // lower(name) -> index
	for _, chunk := range strings.Split(base, ";") {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		eq := strings.Index(chunk, "=")
		if eq <= 0 {
			continue
		}
		k := strings.TrimSpace(chunk[:eq])
		v := strings.TrimSpace(chunk[eq+1:])
		kl := strings.ToLower(k)
		if idx, ok := seen[kl]; ok {
			pairs[idx].v = v
			continue
		}
		seen[kl] = len(pairs)
		pairs = append(pairs, pair{k: k, v: v})
	}
	kl := strings.ToLower(name)
	if idx, ok := seen[kl]; ok {
		pairs[idx].k = name
		pairs[idx].v = value
	} else {
		pairs = append(pairs, pair{k: name, v: value})
	}
	var b strings.Builder
	for i, p := range pairs {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(p.k)
		b.WriteString("=")
		b.WriteString(p.v)
	}
	return b.String()
}
