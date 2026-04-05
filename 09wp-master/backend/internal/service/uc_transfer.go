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

// ucAPIBase 与 xinyue-search UcPan 一致：UC 官方 PC 接口域名（勿再用 drive-h.uc.cn / token，易 EOF 或与 pr 不匹配）
const ucAPIBase = "https://pc-api.uc.cn"

var ucSharePattern = regexp.MustCompile(`(?i)https?://(?:drive|yun)\.uc\.cn/s/([a-zA-Z0-9]+)`)

type UcTransferResult struct {
	ShareCode   string `json:"share_code"`
	Title       string `json:"title,omitempty"`
	Message     string `json:"message"`
	Raw         any    `json:"raw,omitempty"`
	OwnShareURL string `json:"own_share_url,omitempty"`
}

func ParseUcShare(link string) (shareCode string, passcode string, err error) {
	m := ucSharePattern.FindStringSubmatch(link)
	if len(m) < 2 {
		return "", "", fmt.Errorf("不是有效的 UC 网盘分享链接")
	}
	shareCode = m[1]
	u, perr := url.Parse(link)
	if perr == nil {
		passcode = strings.TrimSpace(u.Query().Get("pwd"))
	}
	return shareCode, passcode, nil
}

func ucNormalizeStoken(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), " ", "+")
}

func ucPickStoken(resp map[string]any) string {
	if resp == nil {
		return ""
	}
	if s, ok := getString(resp, "data", "token_info", "stoken"); ok && s != "" {
		return ucNormalizeStoken(s)
	}
	if s, ok := getString(resp, "data", "stoken"); ok && s != "" {
		return ucNormalizeStoken(s)
	}
	return ""
}

func ucHTTPStatus(m map[string]any) float64 {
	if m == nil {
		return 0
	}
	switch v := m["status"].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	}
	return 0
}

func ucHTTPMessage(m map[string]any) string {
	if m == nil {
		return ""
	}
	s, _ := m["message"].(string)
	return strings.TrimSpace(s)
}

// UcSaveByShareLink 使用 UC 网盘 Cookie 转存（对齐 xinyue-search UcPan：pc-api.uc.cn + pr=UCBrowser + v2/detail 取 stoken）
func UcSaveByShareLink(link string, passcodeOverride string) (UcTransferResult, error) {
	cfg, err := LoadNetdiskCredentials()
	if err != nil {
		return UcTransferResult{}, err
	}
	picked := PickUCCookie(cfg)
	cookie := picked.Cookie
	if cookie == "" {
		return UcTransferResult{}, fmt.Errorf("请先在「网盘凭证」页面填写 UC Cookie（或多账号轮询列表）")
	}

	shareCode, passcode, err := ParseUcShare(link)
	if err != nil {
		return UcTransferResult{}, err
	}
	if strings.TrimSpace(passcodeOverride) != "" {
		passcode = strings.TrimSpace(passcodeOverride)
	}
	ucproDebugf("uc transfer start shareCode=%s", shareCode)
	folderID := effectiveQuarkUCFolderID(picked, cfg.UcTargetFolderID)

	client := &http.Client{Timeout: 30 * time.Second}

	// 1) stoken：POST .../sharepage/v2/detail（UcPan#getStoken）
	v2Body, _ := json.Marshal(map[string]string{
		"pwd_id":   shareCode,
		"passcode": passcode,
	})
	v2URL := ucAPIBase + "/1/clouddrive/share/sharepage/v2/detail?pr=UCBrowser&fr=pc"
	tokenResp, err := httpDoJSONUC(client, http.MethodPost, v2URL, cookie, v2Body, "UC网盘")
	if err != nil {
		return UcTransferResult{}, ucNetErr(err)
	}
	if st := ucHTTPStatus(tokenResp); st != 0 && st != 200 {
		msg := ucHTTPMessage(tokenResp)
		if msg == "" {
			msg = fmt.Sprintf("status=%v", st)
		}
		if strings.Contains(strings.ToLower(msg), "guest") || strings.Contains(msg, "require login") {
			msg = "UC 未登录或 Cookie 失效，请重新登录 drive.uc.cn 后更新 uc_cookie"
		}
		return UcTransferResult{}, fmt.Errorf("UC：%s", msg)
	}
	stoken := ucPickStoken(tokenResp)
	if stoken == "" {
		return UcTransferResult{}, fmt.Errorf("获取 stoken 失败")
	}
	ucproDebugf("uc stoken acquired")

	// 2) 分享文件列表：GET .../sharepage/detail（UcPan#getShare）
	dq := url.Values{}
	dq.Set("pr", "UCBrowser")
	dq.Set("fr", "pc")
	dq.Set("pwd_id", shareCode)
	dq.Set("stoken", stoken)
	dq.Set("pdir_fid", "0")
	dq.Set("force", "0")
	dq.Set("_page", "1")
	dq.Set("_size", "100")
	dq.Set("_fetch_banner", "1")
	dq.Set("_fetch_share", "1")
	dq.Set("_fetch_total", "1")
	dq.Set("_sort", "file_type:asc,updated_at:desc")
	detailURL := ucAPIBase + "/1/clouddrive/share/sharepage/detail?" + dq.Encode()
	detailResp, err := httpDoJSONUC(client, http.MethodGet, detailURL, cookie, nil, "UC网盘")
	if err != nil {
		return UcTransferResult{}, ucNetErr(err)
	}
	if st := ucHTTPStatus(detailResp); st != 0 && st != 200 {
		return UcTransferResult{}, fmt.Errorf("UC：%s", ucHTTPMessage(detailResp))
	}
	list, ok := getAny(detailResp, "data", "list").([]any)
	if !ok || len(list) == 0 {
		return UcTransferResult{}, fmt.Errorf("分享内容为空")
	}
	ucproDebugf("uc share detail list count=%d", len(list))
	fids := make([]string, 0, len(list))
	tokens := make([]string, 0, len(list))
	for _, item := range list {
		m, _ := item.(map[string]any)
		fid := ucproRowFid(m)
		tk := strings.TrimSpace(ucproToString(m["share_fid_token"]))
		if fid != "" && tk != "" {
			fids = append(fids, fid)
			tokens = append(tokens, tk)
		}
	}
	if len(fids) == 0 {
		return UcTransferResult{}, fmt.Errorf("未解析到可转存文件")
	}
	ucproDebugf("uc save candidates fids=%d", len(fids))

	// 3) 转存：POST .../sharepage/save?entry=update_share（UcPan#getShareSave）
	saveReq := map[string]any{
		"fid_list":       fids,
		"fid_token_list": tokens,
		"to_pdir_fid":    folderID,
		"pwd_id":         shareCode,
		"stoken":         stoken,
		"pdir_fid":       "0",
		"scene":          "link",
	}
	saveBody, _ := json.Marshal(saveReq)
	sq := url.Values{}
	sq.Set("entry", "update_share")
	sq.Set("pr", "UCBrowser")
	sq.Set("fr", "pc")
	saveURL := ucAPIBase + "/1/clouddrive/share/sharepage/save?" + sq.Encode()
	saveResp, err := httpDoJSONUC(client, http.MethodPost, saveURL, cookie, saveBody, "UC网盘")
	if err != nil {
		return UcTransferResult{}, ucNetErr(err)
	}
	if st := ucHTTPStatus(saveResp); st != 0 && st != 200 {
		return UcTransferResult{}, fmt.Errorf("UC 转存失败：%s", ucHTTPMessage(saveResp))
	}
	msg, _ := saveResp["message"].(string)
	if msg == "" {
		msg = "转存请求已提交"
	}
	ucproDebugf("uc save accepted message=%s", strings.TrimSpace(msg))
	title := ""
	if first, ok := list[0].(map[string]any); ok {
		title, _ = first["file_name"].(string)
		if strings.TrimSpace(title) == "" {
			title, _ = first["name"].(string)
		}
	}
	out := UcTransferResult{ShareCode: shareCode, Title: strings.TrimSpace(title), Message: msg, Raw: saveResp}
	if cfg.ReplaceLinkAfterTransfer {
		// 全程 pc-api.uc.cn，与 ucproProductParam(UCBrowser) 一致
		u, err := ucproReplaceWithOwnShareLink(client, ucAPIBase, cookie, folderID, "drive.uc.cn", "UC网盘", len(fids), saveResp, "")
		if err != nil {
			ucproDebugf("uc own-share failed err=%v", err)
			out.Message = msg + "（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
		} else {
			ucproDebugf("uc own-share success")
			out.OwnShareURL = u
		}
	}
	return out, nil
}

// ucNetErr 将常见网络错误转为可读中文。
func ucNetErr(err error) error {
	if err == nil {
		return nil
	}
	s := err.Error()
	if strings.Contains(s, "no such host") {
		return fmt.Errorf("服务器无法解析网盘接口域名（请检查 DNS/网络或防火墙）。原始错误：%v", err)
	}
	if strings.Contains(s, "EOF") || strings.Contains(s, "connection reset") || strings.Contains(s, "broken pipe") {
		return fmt.Errorf("连接 UC 接口异常中断（EOF/重置），请重试；若持续出现请检查网络或代理。原始错误：%v", err)
	}
	return err
}
