package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

)

var re115SharePath = regexp.MustCompile(`(?i)(?:115\.com|115cdn\.com)/s/([^/?#]+)`)

type Pan115TransferResult struct {
	ShareCode   string `json:"share_code"`
	Message     string `json:"message"`
	OwnShareURL string `json:"own_share_url,omitempty"`
}

// ParsePan115Share 解析 115 分享码与提取码（password 查询参数或资源表 extract_code）
func ParsePan115Share(link string, passOverride string) (shareCode, receiveCode string, err error) {
	m := re115SharePath.FindStringSubmatch(link)
	if len(m) < 2 {
		return "", "", fmt.Errorf("不是有效的115分享链接")
	}
	shareCode = m[1]
	u, perr := url.Parse(link)
	if perr == nil {
		receiveCode = strings.TrimSpace(u.Query().Get("password"))
	}
	if strings.TrimSpace(passOverride) != "" {
		receiveCode = strings.TrimSpace(passOverride)
	}
	return shareCode, receiveCode, nil
}

// Pan115SaveByShareLink 使用系统配置中的 115 Cookie 将分享转存到目标目录
func Pan115SaveByShareLink(link string, passOverride string) (Pan115TransferResult, error) {
	cfg, err := LoadNetdiskCredentials()
	if err != nil {
		return Pan115TransferResult{}, err
	}
	picked := PickPan115Cookie(cfg)
	cookie := picked.Cookie
	if cookie == "" {
		return Pan115TransferResult{}, fmt.Errorf("请先在「网盘凭证」页面填写 115 Cookie（或多账号轮询列表）")
	}
	shareCode, receiveCode, err := ParsePan115Share(link, passOverride)
	if err != nil {
		return Pan115TransferResult{}, err
	}
	targetCid := effectivePan115Cid(picked, cfg.Pan115TargetFolderID)

	client := &http.Client{Timeout: 30 * time.Second}
	base := "https://webapi.115.com"

	snapURL := fmt.Sprintf("%s/share/snap?share_code=%s&receive_code=%s&offset=0&limit=50&cid=",
		base, url.QueryEscape(shareCode), url.QueryEscape(receiveCode))
	snapRaw, err := http115GET(client, snapURL, cookie)
	if err != nil {
		return Pan115TransferResult{}, err
	}
	state, _ := snapRaw["state"].(bool)
	if !state {
		msg, _ := snapRaw["error"].(string)
		if msg == "" {
			msg = "获取分享文件列表失败"
		}
		return Pan115TransferResult{}, fmt.Errorf("%s", msg)
	}
	data, ok := snapRaw["data"].(map[string]any)
	if !ok {
		return Pan115TransferResult{}, fmt.Errorf("115分享数据异常")
	}
	list, _ := data["list"].([]any)
	if len(list) == 0 {
		return Pan115TransferResult{}, fmt.Errorf("分享内容为空")
	}

	extra := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36",
		"Referer":    fmt.Sprintf("https://115.com/s/%s", shareCode),
	}

	var lastMsg string
	received := 0
	for _, it := range list {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		cid := stringFromAny(m["cid"])
		if cid == "" {
			continue
		}
		form := url.Values{}
		form.Set("cid", targetCid)
		form.Set("share_code", shareCode)
		form.Set("receive_code", receiveCode)
		form.Set("file_id", cid)
		recvURL := base + "/share/receive"
		resp, err := httpDoFormPost(client, recvURL, cookie, form.Encode(), extra)
		if err != nil {
			return Pan115TransferResult{}, err
		}
		ok2, _ := resp["state"].(bool)
		if !ok2 {
			e, _ := resp["error"].(string)
			if e == "" {
				e = "转存失败"
			}
			return Pan115TransferResult{}, fmt.Errorf("%s", e)
		}
		lastMsg, _ = resp["error"].(string)
		if lastMsg == "" {
			lastMsg = "转存成功"
		}
		received++
	}
	if lastMsg == "" {
		lastMsg = "转存完成"
	}
	out := Pan115TransferResult{ShareCode: shareCode, Message: lastMsg}
	if cfg.ReplaceLinkAfterTransfer && received > 0 {
		u, err := pan115ReplaceWithCookieShare(client, cookie, targetCid, received)
		if err != nil {
			out.Message = lastMsg + "（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
		} else {
			out.OwnShareURL = u
		}
	}
	return out, nil
}

func stringFromAny(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return fmt.Sprintf("%.0f", x)
	case int:
		return fmt.Sprintf("%d", x)
	case int64:
		return fmt.Sprintf("%d", x)
	default:
		return ""
	}
}

func http115GET(client *http.Client, endpoint, cookie string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("cookie", cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("115接口错误: %s", string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("115返回解析失败")
	}
	return out, nil
}
