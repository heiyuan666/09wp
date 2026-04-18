package service

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

// 夸克目录定时清理：对齐社区脚本列目录 + 按创建时间删旧文件的思路。
// 参考：https://github.com/NamelessClub/quark_deleter/blob/main/quark_deleter_unified.py

var (
	quarkCleanupMu          sync.Mutex
	lastQuarkCleanupRun     time.Time
)

// StartQuarkFolderCleanupWorker 后台按配置间隔扫描夸克目录并删除超时条目（仅单层目录）。
func StartQuarkFolderCleanupWorker() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var cfg model.SystemConfig
			if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
				continue
			}
			if !cfg.QuarkCleanupEnabled {
				continue
			}
			intervalMin := cfg.QuarkCleanupIntervalMinutes
			if intervalMin < 1 {
				intervalMin = 5
			}
			quarkCleanupMu.Lock()
			if !lastQuarkCleanupRun.IsZero() && time.Since(lastQuarkCleanupRun) < time.Duration(intervalMin)*time.Minute {
				quarkCleanupMu.Unlock()
				continue
			}
			lastQuarkCleanupRun = time.Now()
			quarkCleanupMu.Unlock()

			nd, err := LoadNetdiskCredentials()
			if err != nil {
				log.Printf("quark folder cleanup: load credentials: %v", err)
				continue
			}
			deleted, err := runQuarkFolderCleanupOnce(cfg, nd)
			if err != nil {
				log.Printf("quark folder cleanup: %v", err)
				continue
			}
			if deleted > 0 {
				log.Printf("quark folder cleanup: deleted %d item(s)", deleted)
			}
		}
	}()
}

func quarkCleanupRowCreatedAtMs(row map[string]any) (int64, bool) {
	if row == nil {
		return 0, false
	}
	tryMs := func(key string) (int64, bool) {
		v := getFloat(row, key)
		if v <= 0 {
			return 0, false
		}
		if v >= 1e12 { // 毫秒
			return int64(v), true
		}
		if v >= 1e9 { // 秒
			return int64(v * 1000), true
		}
		return 0, false
	}
	for _, k := range []string{"created_at", "l_created_at", "operated_at", "updated_at"} {
		if ms, ok := tryMs(k); ok {
			return ms, true
		}
	}
	if v := getFloat(row, "itime"); v > 0 {
		return int64(v * 1000), true
	}
	return 0, false
}

func quarkCleanupListFolderPage(client *http.Client, apiBase, cookie, folderID string, page, pageSize int) ([]map[string]any, int64, error) {
	q := url.Values{}
	q.Set("pr", "ucpro")
	q.Set("fr", "pc")
	q.Set("uc_param_str", "")
	q.Set("pdir_fid", folderID)
	q.Set("_page", strconv.Itoa(page))
	q.Set("_size", strconv.Itoa(pageSize))
	q.Set("_fetch_total", "1")
	q.Set("_fetch_sub_dirs", "0")
	q.Set("_sort", "created_at:asc")
	q.Set("__t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	reqURL := fmt.Sprintf("%s/1/clouddrive/file/sort?%s", strings.TrimRight(apiBase, "/"), q.Encode())
	resp, err := httpDoJSON(client, http.MethodGet, reqURL, cookie, nil, "夸克")
	if err != nil {
		return nil, 0, err
	}
	code := int(getFloat(resp, "code"))
	status := int(getFloat(resp, "status"))
	if code != 0 || status != 200 {
		msg, _ := getString(resp, "message")
		if msg == "" {
			msg = "未知错误"
		}
		return nil, 0, fmt.Errorf("%s", msg)
	}
	dataNode, _ := getAny(resp, "data").(map[string]any)
	var list []any
	if dataNode != nil {
		list, _ = dataNode["list"].([]any)
	}
	var total int64
	if meta, ok := resp["metadata"].(map[string]any); ok {
		total = int64(getFloat(meta, "_total"))
	}
	out := make([]map[string]any, 0, len(list))
	for _, it := range list {
		row, ok := it.(map[string]any)
		if ok {
			out = append(out, row)
		}
	}
	return out, total, nil
}

func runQuarkFolderCleanupOnce(cfg model.SystemConfig, nd model.NetdiskCredential) (deleted int, err error) {
	picked := PickQuarkCookie(nd)
	cookie := strings.TrimSpace(picked.Cookie)
	if cookie == "" {
		return 0, fmt.Errorf("夸克 Cookie 未配置（请在「网盘凭证」中填写）")
	}
	folder := strings.TrimSpace(cfg.QuarkCleanupFolderID)
	if folder == "" {
		folder = strings.TrimSpace(effectiveQuarkUCFolderID(picked, nd.QuarkTargetFolderID))
	}
	if folder == "" || folder == "0" {
		return 0, fmt.Errorf("未配置有效的清理目录 fid（请在系统配置「夸克清理目录」填写，或设置网盘凭证中的夸克转存目录为非根目录）；为避免误删，不支持根目录 0")
	}
	olderMin := cfg.QuarkCleanupOlderThanMinutes
	if olderMin < 1 {
		olderMin = 60
	}
	cutoff := time.Now().Add(-time.Duration(olderMin) * time.Minute)

	client := &http.Client{Timeout: 45 * time.Second}
	rawBase := "https://drive-h.quark.cn"
	apiBase := ucproShareBaseURL(rawBase)

	toDelete := make([]string, 0, 32)
	page := 1
	pageSize := 50
	for {
		rows, total, lerr := quarkCleanupListFolderPage(client, apiBase, cookie, folder, page, pageSize)
		if lerr != nil {
			return deleted, lerr
		}
		for _, row := range rows {
			ms, ok := quarkCleanupRowCreatedAtMs(row)
			if !ok {
				continue
			}
			t := time.Unix(ms/1000, (ms%1000)*1e6)
			if t.Before(cutoff) {
				fid, _ := getString(row, "fid")
				if fid != "" {
					toDelete = append(toDelete, fid)
				}
			}
		}
		if total > 0 && int64(page*pageSize) >= total {
			break
		}
		if len(rows) < pageSize {
			break
		}
		page++
		if page > 500 {
			break
		}
		time.Sleep(400 * time.Millisecond)
	}

	const batch = 80
	for i := 0; i < len(toDelete); i += batch {
		j := i + batch
		if j > len(toDelete) {
			j = len(toDelete)
		}
		if derr := quarkDeleteByFids(client, apiBase, cookie, toDelete[i:j]); derr != nil {
			return deleted, derr
		}
		deleted += j - i
	}
	return deleted, nil
}
