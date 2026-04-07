package service

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const pan123MainAPI = "https://www.123pan.com/b/api"

type Pan123TransferResult struct {
	ShareKey    string `json:"share_key"`
	Message     string `json:"message"`
	OwnShareURL string `json:"own_share_url,omitempty"`
}

var re123ShareKey = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?(?:123684|123685|123912|123pan|123592|123865)\.(?:com|cn)/s/([a-zA-Z0-9-]+)`)

// Pan123SaveByShareLink 使用登录后的 Bearer Token 将分享文件转存到网盘目录
func Pan123SaveByShareLink(link string, sharePwd string) (Pan123TransferResult, error) {
	cfg, err := LoadNetdiskCredentials()
	if err != nil {
		return Pan123TransferResult{}, err
	}
	p123 := PickPan123Cookie(cfg)
	bearer := NormalizePan123Bearer(p123.Cookie)
	if bearer == "" || bearer == "Bearer " {
		return Pan123TransferResult{}, fmt.Errorf("请先在「网盘凭证」页面填写 123 Token（或多账号轮询列表）")
	}
	m := re123ShareKey.FindStringSubmatch(link)
	if len(m) < 2 {
		return Pan123TransferResult{}, fmt.Errorf("不是有效的123云盘分享链接")
	}
	shareKey := m[1]
	targetParent := effectivePan123Parent(p123, cfg.Pan123TargetFolderID)

	client := &http.Client{Timeout: 45 * time.Second}
	ids, err := pan123CollectShareFileIDs(client, bearer, shareKey, strings.TrimSpace(sharePwd))
	if err != nil {
		return Pan123TransferResult{}, err
	}
	if len(ids) == 0 {
		return Pan123TransferResult{}, fmt.Errorf("分享目录下没有可转存的文件")
	}

	var lastMsg string
	var newFileIDs []int64
	for _, fid := range ids {
		msg, newID, err := pan123CopyOne(client, bearer, shareKey, fid, targetParent)
		if err != nil {
			return Pan123TransferResult{}, err
		}
		lastMsg = msg
		if newID != 0 {
			newFileIDs = append(newFileIDs, newID)
		}
	}
	if lastMsg == "" {
		lastMsg = "转存完成"
	}
	out := Pan123TransferResult{ShareKey: shareKey, Message: lastMsg}
	if cfg.ReplaceLinkAfterTransfer && len(newFileIDs) > 0 {
		u, err := pan123CreateOwnShare(client, bearer, newFileIDs)
		if err != nil {
			out.Message = lastMsg + "（未生成本人分享链接：" + trimTo255(err.Error()) + "）"
		} else {
			out.OwnShareURL = u
		}
	}
	return out, nil
}

func pan123CollectShareFileIDs(client *http.Client, bearer, shareKey, sharePwd string) ([]int64, error) {
	var out []int64
	var walk func(parentID string, depth int) error
	walk = func(parentID string, depth int) error {
		if depth > 12 {
			return nil
		}
		page := 1
		for {
			q := url.Values{}
			q.Set("limit", "100")
			q.Set("next", "0")
			q.Set("orderBy", "file_id")
			q.Set("orderDirection", "desc")
			q.Set("parentFileId", parentID)
			q.Set("Page", strconv.Itoa(page))
			q.Set("shareKey", shareKey)
			q.Set("SharePwd", sharePwd)
			rawURL := pan123MainAPI + "/share/get?" + q.Encode()
			signed := pan123SignURL(rawURL)
			resp, err := httpDoJSONBearer(client, http.MethodGet, signed, bearer, nil, "123云盘")
			if err != nil {
				return err
			}
			if code := int(getFloat(resp, "code")); code != 0 {
				msg, _ := resp["message"].(string)
				if msg == "" {
					msg = "获取分享列表失败"
				}
				return fmt.Errorf("%s", msg)
			}
			data, _ := resp["data"].(map[string]any)
			list, _ := data["InfoList"].([]any)
			if len(list) == 0 {
				list, _ = data["infoList"].([]any)
			}
			next, _ := data["Next"].(string)
			if next == "" {
				next, _ = data["next"].(string)
			}
			for _, it := range list {
				row, ok := it.(map[string]any)
				if !ok {
					continue
				}
				typ := int(getFloat(row, "Type"))
				fid := int64(getFloat(row, "FileId"))
				if fid == 0 {
					continue
				}
				switch typ {
				case 0:
					out = append(out, fid)
				case 1:
					_ = walk(strconv.FormatInt(fid, 10), depth+1)
				}
			}
			if len(list) == 0 || next == "-1" {
				break
			}
			page++
		}
		return nil
	}
	if err := walk("0", 0); err != nil {
		return nil, err
	}
	return out, nil
}

func pan123CopyOne(client *http.Client, bearer, shareKey string, fileID int64, parentFileID string) (string, int64, error) {
	body := map[string]any{
		"driveId":      0,
		"shareKey":     shareKey,
		"fileIdList":   []int64{fileID},
		"parentFileId": parentFileID,
		"duplicate":    2,
	}
	raw, _ := json.Marshal(body)
	rawURL := pan123MainAPI + "/file/copy"
	signed := pan123SignURL(rawURL)
	resp, err := httpDoJSONBearer(client, http.MethodPost, signed, bearer, raw, "123云盘")
	if err != nil {
		return "", 0, err
	}
	if code := int(getFloat(resp, "code")); code != 0 {
		msg, _ := resp["message"].(string)
		if msg == "" {
			msg = "转存失败"
		}
		return "", 0, fmt.Errorf("%s", msg)
	}
	var newID int64
	if data, ok := resp["data"].(map[string]any); ok {
		newID = int64(getFloat(data, "FileId"))
		if newID == 0 {
			newID = int64(getFloat(data, "DestFileId"))
		}
		if newID == 0 {
			if arr, ok := data["InfoList"].([]any); ok {
				for _, it := range arr {
					row, ok := it.(map[string]any)
					if !ok {
						continue
					}
					newID = int64(getFloat(row, "FileId"))
					if newID != 0 {
						break
					}
				}
			}
			if newID == 0 {
				if arr, ok := data["fileList"].([]any); ok {
					for _, it := range arr {
						row, ok := it.(map[string]any)
						if !ok {
							continue
						}
						newID = int64(getFloat(row, "FileId"))
						if newID != 0 {
							break
						}
					}
				}
			}
		}
	}
	msg, _ := resp["message"].(string)
	if msg == "" {
		msg = "ok"
	}
	return msg, newID, nil
}

func pan123CreateOwnShare(client *http.Client, bearer string, fileIDs []int64) (string, error) {
	if len(fileIDs) == 0 {
		return "", fmt.Errorf("无文件 ID")
	}
	body := map[string]any{
		"DriveId":     0,
		"FileIdList":  fileIDs,
		"ShareName":   "",
		"SharePwd":    "",
		"ExpiredTime": "",
	}
	raw, _ := json.Marshal(body)
	var lastErr string
	for _, path := range []string{"/share/create", "/file/share_create"} {
		rawURL := pan123MainAPI + path
		signed := pan123SignURL(rawURL)
		resp, err := httpDoJSONBearer(client, http.MethodPost, signed, bearer, raw, "123云盘")
		if err != nil {
			lastErr = err.Error()
			continue
		}
		if code := int(getFloat(resp, "code")); code != 0 {
			lastErr, _ = resp["message"].(string)
			continue
		}
		if data, ok := resp["data"].(map[string]any); ok {
			if u, ok := data["ShareUrl"].(string); ok && u != "" {
				return u, nil
			}
			if u, ok := data["shareUrl"].(string); ok && u != "" {
				return u, nil
			}
			if key, ok := data["ShareKey"].(string); ok && key != "" {
				return "https://www.123pan.com/s/" + key, nil
			}
			if key, ok := data["shareKey"].(string); ok && key != "" {
				return "https://www.123pan.com/s/" + key, nil
			}
		}
	}
	if lastErr != "" {
		return "", fmt.Errorf("%s", lastErr)
	}
	return "", fmt.Errorf("未解析到分享链接")
}

var pan123Table = []byte("adegfhlmynijopkqrstubcvwsz")

func pan123SignURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	path := u.Path
	k, v := pan123SignPath(path)
	q := u.Query()
	q.Set(k, v)
	u.RawQuery = q.Encode()
	return u.String()
}

func pan123SignPath(path string) (k, v string) {
	random := fmt.Sprintf("%.f", math.Round(1e7*rand.Float64()))
	now := time.Now().In(time.FixedZone("CST", 8*3600))
	timestamp := fmt.Sprint(now.Unix())
	nowStr := []byte(now.Format("200601021504"))
	for i := 0; i < len(nowStr); i++ {
		nowStr[i] = pan123Table[nowStr[i]-48]
	}
	timeSign := fmt.Sprint(crc32.ChecksumIEEE(nowStr))
	data := strings.Join([]string{timestamp, random, path, "web", "3", timeSign}, "|")
	dataSign := fmt.Sprint(crc32.ChecksumIEEE([]byte(data)))
	return timeSign, strings.Join([]string{timestamp, random, dataSign}, "-")
}
