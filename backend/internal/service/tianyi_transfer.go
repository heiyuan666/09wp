package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

)

type TianyiTransferResult struct {
	ShareID     int64  `json:"share_id"`
	Message     string `json:"message"`
	OwnShareURL string `json:"own_share_url,omitempty"`
}

var errTianyiNoCode = errors.New("无法从天翼分享链接解析分享码")

var reTianyiAccessInURL = regexp.MustCompile(`（访问码[：:]\s*([a-zA-Z0-9]+)）`)

// TianyiSaveByShareLink 使用浏览器登录后的 Cookie 将分享批量保存到个人云
func TianyiSaveByShareLink(link string, extractCode string) (TianyiTransferResult, error) {
	cfg, err := LoadNetdiskCredentials()
	if err != nil {
		return TianyiTransferResult{}, err
	}
	picked := PickTianyiCookie(cfg)
	cookie := picked.Cookie
	if cookie == "" {
		return TianyiTransferResult{}, fmt.Errorf("请先在「网盘凭证」页面填写天翼 Cookie（或多账号轮询列表）")
	}
	codeValue, accessFromURL, err := tianyiExtractCode(link)
	if err != nil {
		return TianyiTransferResult{}, err
	}
	access := strings.TrimSpace(extractCode)
	if access == "" {
		access = accessFromURL
	}

	client := &http.Client{Timeout: 45 * time.Second}
	ua := "Mozilla/5.0 (Linux; U; Android 11; KB2000 Build/RP1A.201005.001) AppleWebKit/537.36 Chrome/74.0.3729.136 Mobile Safari/537.36 Ecloud/9.0.6"

	shareParam := codeValue
	if access != "" {
		shareParam = fmt.Sprintf("%s（访问码：%s）", codeValue, access)
	}
	noCache := fmt.Sprintf("%f", rand.Float64())
	infoURL := fmt.Sprintf("https://cloud.189.cn/api/open/share/getShareInfoByCodeV2.action?noCache=%s&shareCode=%s",
		url.QueryEscape(noCache), url.QueryEscape(shareParam))
	infoRaw, err := tianyiGET(client, infoURL, cookie, link, ua)
	if err != nil {
		return TianyiTransferResult{}, err
	}
	if rc := int(getFloat(infoRaw, "res_code")); rc != 0 {
		msg, _ := infoRaw["res_message"].(string)
		if msg == "" {
			msg = "获取分享信息失败"
		}
		return TianyiTransferResult{}, fmt.Errorf("%s", msg)
	}
	shareID := int64(getFloat(infoRaw, "shareId"))
	shareDirFileID := tianyiStringFromAny(infoRaw["fileId"])
	shareMode := tianyiStringFromAny(infoRaw["shareMode"])
	if shareMode == "" {
		shareMode = "1"
	}
	if shareID <= 0 || shareDirFileID == "" {
		return TianyiTransferResult{}, fmt.Errorf("分享信息不完整")
	}

	targetFolder := effectiveTianyiFolder(picked, cfg.TianyiTargetFolderID)

	tasks, err := tianyiCollectRootFileTasks(client, cookie, ua, link, shareID, shareDirFileID, shareMode)
	if err != nil {
		return TianyiTransferResult{}, err
	}
	if len(tasks) == 0 {
		return TianyiTransferResult{}, fmt.Errorf("分享根目录没有可转存的文件")
	}

	wrap := func(msg string) TianyiTransferResult {
		out := TianyiTransferResult{ShareID: shareID, Message: msg}
		if cfg.ReplaceLinkAfterTransfer && len(tasks) > 0 {
			if u, err := tianyiTryOwnShareLink(client, cookie, ua, link, targetFolder, len(tasks)); err != nil {
				out.Message = msg + "（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
			} else {
				out.OwnShareURL = u
			}
		}
		return out
	}

	taskInfosJSON, _ := json.Marshal(tasks)
	form := url.Values{}
	form.Set("type", "SHARE_SAVE")
	form.Set("taskInfos", string(taskInfosJSON))
	form.Set("targetFolderId", targetFolder)
	form.Set("shareId", strconv.FormatInt(shareID, 10))

	batchURL := "https://cloud.189.cn/api/open/batch/createBatchTask.action"
	batchRaw, err := tianyiPOSTForm(client, batchURL, cookie, form.Encode(), link, ua)
	if err != nil {
		return TianyiTransferResult{}, err
	}
	if rc := int(getFloat(batchRaw, "res_code")); rc != 0 {
		msg, _ := batchRaw["res_message"].(string)
		if msg == "" {
			msg = "创建转存任务失败"
		}
		return TianyiTransferResult{}, fmt.Errorf("%s", msg)
	}
	taskID := tianyiStringFromAny(batchRaw["taskId"])
	if taskID == "" {
		return wrap("转存任务已提交"), nil
	}

	for i := 0; i < 120; i++ {
		time.Sleep(1 * time.Second)
		chk := url.Values{}
		chk.Set("taskId", taskID)
		chk.Set("type", "SHARE_SAVE")
		chkURL := "https://cloud.189.cn/api/open/batch/checkBatchTask.action"
		st, err := tianyiPOSTForm(client, chkURL, cookie, chk.Encode(), link, ua)
		if err != nil {
			return TianyiTransferResult{}, err
		}
		taskStatus := int(getFloat(st, "taskStatus"))
		errCode := tianyiStringFromAny(st["errorCode"])
		if taskStatus != 3 || errCode != "" {
			if errCode != "" && errCode != "null" {
				return TianyiTransferResult{}, fmt.Errorf("转存失败: %s", errCode)
			}
			return wrap("转存完成"), nil
		}
	}
	return wrap("转存任务已提交（仍在处理中）"), nil
}

func tianyiExtractCode(urlStr string) (codeValue, accessFromURL string, err error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", err
	}
	codeValue = parsedURL.Query().Get("code")
	if codeValue == "" {
		path := parsedURL.Path
		if strings.HasPrefix(path, "/t/") {
			codeValue = strings.TrimPrefix(path, "/t/")
			if idx := strings.Index(codeValue, "/"); idx != -1 {
				codeValue = codeValue[:idx]
			}
		}
	}
	if codeValue == "" && parsedURL.Fragment != "" {
		fragment := parsedURL.Fragment
		if strings.HasPrefix(fragment, "/t/") {
			codeValue = strings.TrimPrefix(fragment, "/t/")
			if idx := strings.Index(codeValue, "/"); idx != -1 {
				codeValue = codeValue[:idx]
			}
		} else if strings.HasPrefix(fragment, "#/t/") {
			codeValue = strings.TrimPrefix(fragment, "#/t/")
			if idx := strings.Index(codeValue, "/"); idx != -1 {
				codeValue = codeValue[:idx]
			}
		}
	}
	if codeValue == "" {
		return "", "", errTianyiNoCode
	}
	if m := reTianyiAccessInURL.FindStringSubmatch(urlStr); len(m) >= 2 {
		accessFromURL = m[1]
	}
	return codeValue, accessFromURL, nil
}

func tianyiCollectRootFileTasks(client *http.Client, cookie, ua, referer string, shareID int64, shareDirFileID, shareMode string) ([]map[string]any, error) {
	listURL := "https://cloud.189.cn/api/open/share/listShareDir.action"
	q := url.Values{}
	q.Set("pageNum", "1")
	q.Set("pageSize", "10000")
	q.Set("fileId", shareDirFileID)
	q.Set("shareDirFileId", shareDirFileID)
	q.Set("isFolder", "true")
	q.Set("shareId", strconv.FormatInt(shareID, 10))
	q.Set("shareMode", shareMode)
	q.Set("iconOption", "5")
	q.Set("orderBy", "lastOpTime")
	q.Set("descending", "true")
	q.Set("accessCode", "")
	raw, err := tianyiGET(client, listURL+"?"+q.Encode(), cookie, referer, ua)
	if err != nil {
		return nil, err
	}
	if rc := int(getFloat(raw, "res_code")); rc != 0 {
		msg, _ := raw["res_message"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	fileListAO, ok := raw["fileListAO"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("天翼目录列表格式异常")
	}
	files, _ := fileListAO["fileList"].([]any)
	var tasks []map[string]any
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
		fid := tianyiStringFromAny(m["id"])
		if fid == "" {
			fid = tianyiStringFromAny(m["fileId"])
		}
		name, _ := m["name"].(string)
		if fid == "" {
			continue
		}
		tasks = append(tasks, map[string]any{
			"fileId":   fid,
			"fileName": name,
			"isFolder": 0,
		})
	}
	return tasks, nil
}

func tianyiStringFromAny(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return fmt.Sprintf("%.0f", x)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case json.Number:
		return x.String()
	default:
		return ""
	}
}

func tianyiGET(client *http.Client, endpoint, cookie, referer, ua string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json;charset=UTF-8")
	req.Header.Set("cookie", cookie)
	req.Header.Set("referer", referer)
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("天翼接口错误: %s", string(body))
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("天翼返回解析失败")
	}
	return out, nil
}

func tianyiPOSTForm(client *http.Client, endpoint, cookie, body, referer, ua string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json;charset=UTF-8")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("cookie", cookie)
	req.Header.Set("referer", referer)
	req.Header.Set("User-Agent", ua)
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
		return nil, fmt.Errorf("天翼接口错误: %s", string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("天翼返回解析失败")
	}
	return out, nil
}
