package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"

	"gorm.io/gorm"
)

type transferSlot struct {
	isPrimary bool
	extraIdx  int // 仅附加链接时有效，>=0
	link      string
}

// transferSingleShareLink 转存单条分享链接，写 transfer_log，不更新 resources 行。
func transferSingleShareLink(resourceID uint64, link string, pass string, maxRetry int) (ownShareURL string, platform TransferPlatform, resultMsg string, err error) {
	link = strings.TrimSpace(link)
	if link == "" {
		return "", PlatformUnknown, "", fmt.Errorf("链接为空")
	}
	platform = DetectTransferPlatform(link)
	if platform == PlatformUnknown {
		return "", platform, "", fmt.Errorf("无法识别或暂不支持的网盘链接")
	}

	var lastErr error
	for i := 0; i < maxRetry; i++ {
		var e error
		var own string
		var msg string
		var filterLogJSON string
		var savedFileIDsJSON string

		switch platform {
		case PlatformBaidu:
			var r BaiduTransferResult
			r, e = BaiduSaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
			if len(r.SavedPaths) > 0 {
				if b, er := json.Marshal(r.SavedPaths); er == nil {
					savedFileIDsJSON = sanitizeForLegacyUTF8(string(b))
				}
			}
		case PlatformQuark:
			var r QuarkTransferResult
			r, e = QuarkSaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
			if r.FilterLog != nil {
				if b, er := json.Marshal(r.FilterLog); er == nil {
					filterLogJSON = sanitizeForLegacyUTF8(string(b))
				}
			}
			if len(r.SavedFids) > 0 {
				if b, er := json.Marshal(r.SavedFids); er == nil {
					savedFileIDsJSON = sanitizeForLegacyUTF8(string(b))
				}
			}
		case PlatformUC:
			var r UcTransferResult
			r, e = UcSaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
			if len(r.SavedFids) > 0 {
				if b, er := json.Marshal(r.SavedFids); er == nil {
					savedFileIDsJSON = sanitizeForLegacyUTF8(string(b))
				}
			}
		case PlatformPan115:
			var r Pan115TransferResult
			r, e = Pan115SaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
		case PlatformTianyi:
			var r TianyiTransferResult
			r, e = TianyiSaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
		case PlatformPan123:
			var r Pan123TransferResult
			r, e = Pan123SaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
		case PlatformAliyun:
			var r AliyunTransferResult
			r, e = AliyunSaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
		case PlatformXunlei:
			var r XunleiTransferResult
			r, e = XunleiSaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
			if len(r.SavedFileIDs) > 0 {
				if b, er := json.Marshal(r.SavedFileIDs); er == nil {
					savedFileIDsJSON = sanitizeForLegacyUTF8(string(b))
				}
			}
		default:
			e = fmt.Errorf("暂未实现该网盘转存")
		}

		attempt := i + 1
		if e == nil {
			_ = database.DB().Create(&model.ResourceTransferLog{
				ResourceID:   resourceID,
				Attempt:      attempt,
				Platform:     platform.String(),
				Status:       "success",
				Message:      strings.TrimSpace(msg),
				OldLink:      link,
				NewLink:      strings.TrimSpace(own),
				OwnShareURL:  strings.TrimSpace(own),
				FilterLog:    filterLogJSON,
				SavedFileIDs: savedFileIDsJSON,
			}).Error
			return strings.TrimSpace(own), platform, strings.TrimSpace(msg), nil
		}

		lastErr = e
		_ = database.DB().Create(&model.ResourceTransferLog{
			ResourceID:  resourceID,
			Attempt:     attempt,
			Platform:    platform.String(),
			Status:      "failed",
			Message:     "转存失败",
			ErrorDetail: trimTo255(e.Error()),
			OldLink:     link,
		}).Error
		time.Sleep(1200 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("转存失败")
	}
	return "", platform, "", lastErr
}

// TransferResourceWithRetry 按主链接 + extra_links 逐项转存（仅处理已开启对应网盘「自动转存」的链接）。
func TransferResourceWithRetry(resourceID uint64, maxRetry int) error {
	return transferResourceMulti(resourceID, maxRetry, true)
}

// TransferResourceWithRetryForce 忽略自动转存开关，凡可识别网盘链接均尝试（管理端手动重试）。
func TransferResourceWithRetryForce(resourceID uint64, maxRetry int) error {
	return transferResourceMulti(resourceID, maxRetry, false)
}

func transferResourceMulti(resourceID uint64, maxRetry int, onlyIfAutoSave bool) error {
	if maxRetry < 1 {
		maxRetry = 1
	}

	var res model.Resource
	if err := database.DB().First(&res, resourceID).Error; err != nil {
		return err
	}

	// 快照首次转存前的原始分享链接，供「详情页每次重新生成本人分享」使用。
	if strings.TrimSpace(res.TransferSourceLink) == "" {
		el := make([]string, 0, len(res.ExtraLinks))
		for _, u := range res.ExtraLinks {
			if t := strings.TrimSpace(u); t != "" {
				el = append(el, t)
			}
		}
		_ = database.DB().Model(&model.Resource{}).Where("id = ?", resourceID).Updates(map[string]any{
			"transfer_source_link":  trimTo500(strings.TrimSpace(res.Link)),
			"transfer_source_extra":   model.NormalizeExtraShareLinks(el),
		}).Error
	}

	cred, credErr := LoadNetdiskCredentials()
	if credErr != nil {
		return fmt.Errorf("读取网盘凭证失败: %v", credErr)
	}

	pass := strings.TrimSpace(res.ExtractCode)
	primary := strings.TrimSpace(res.Link)
	extrasWork := make([]string, len(res.ExtraLinks))
	for i, u := range res.ExtraLinks {
		extrasWork[i] = strings.TrimSpace(u)
	}

	var slots []transferSlot
	if primary != "" {
		slots = append(slots, transferSlot{isPrimary: true, extraIdx: -1, link: primary})
	}
	for i, u := range extrasWork {
		if strings.TrimSpace(u) == "" {
			continue
		}
		slots = append(slots, transferSlot{isPrimary: false, extraIdx: i, link: strings.TrimSpace(u)})
	}

	if len(slots) == 0 {
		now := time.Now()
		_ = database.DB().Model(&model.Resource{}).Where("id = ?", resourceID).Updates(map[string]any{
			"transfer_status":  "failed",
			"transfer_msg":     "资源无有效分享链接",
			"transfer_last_at": &now,
		}).Error
		return fmt.Errorf("资源无有效分享链接")
	}

	var jobs []transferSlot
	for _, s := range slots {
		if onlyIfAutoSave {
			if !ShouldAutoTransferOnCreate(cred, s.link) {
				continue
			}
		} else if DetectTransferPlatform(s.link) == PlatformUnknown {
			continue
		}
		jobs = append(jobs, s)
	}

	now := time.Now()
	if len(jobs) == 0 {
		msg := "无需转存（未开启对应网盘的自动转存）"
		if !onlyIfAutoSave {
			msg = "无可用网盘链接（无法识别平台）"
		}
		_ = database.DB().Model(&model.Resource{}).Where("id = ?", resourceID).Updates(map[string]any{
			"transfer_status":  "success",
			"transfer_msg":     trimTo255(msg),
			"transfer_last_at": &now,
		}).Error
		return nil
	}

	var failMsgs []string
	okCount := 0
	for _, job := range jobs {
		newURL, _, msg, err := transferSingleShareLink(resourceID, job.link, pass, maxRetry)
		if err != nil {
			failMsgs = append(failMsgs, trimTo255(err.Error()))
			continue
		}
		okCount++
		if strings.TrimSpace(newURL) != "" {
			if job.isPrimary {
				primary = trimTo500(newURL)
			} else if job.extraIdx >= 0 && job.extraIdx < len(extrasWork) {
				extrasWork[job.extraIdx] = trimTo500(newURL)
			}
		}
		_ = msg
	}

	extraJSON := model.NormalizeExtraShareLinks(extrasWork)
	updates := map[string]any{
		"link":                 trimTo500(primary),
		"extra_links":          extraJSON,
		"transfer_last_at":     &now,
		"transfer_retry_count": gorm.Expr("transfer_retry_count + 1"),
		"transfer_status":      "success",
		"transfer_msg":         "",
	}

	var sysCfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&sysCfg).Error; err == nil && sysCfg.ResourceDetailEachClickFreshShare {
		// 详情页「每次重新分享」：不在库里覆盖为本人链接，仅更新转存状态；对外仍展示原始分享。
		delete(updates, "link")
		delete(updates, "extra_links")
	}

	if okCount == 0 {
		updates["transfer_status"] = "failed"
		updates["transfer_msg"] = trimTo255(strings.Join(failMsgs, "; "))
		if updates["transfer_msg"] == "" {
			updates["transfer_msg"] = "转存失败"
		}
	} else if okCount < len(jobs) {
		updates["transfer_msg"] = trimTo255(fmt.Sprintf("已转存 %d/%d 条链接", okCount, len(jobs)))
	} else {
		updates["transfer_msg"] = trimTo255(fmt.Sprintf("已全部转存 %d 条链接", okCount))
	}

	_ = database.DB().Model(&model.Resource{}).Where("id = ?", resourceID).Updates(updates).Error
	if okCount == 0 {
		return fmt.Errorf("%s", updates["transfer_msg"])
	}
	return nil
}

// SyncTimeAutoTransferAllowed 判定 TG/RSS 等资源入库后是否应立即触发转存。
// 开启「详情页自动转存」时为 false：转存改在用户访问详情页时触发，避免入库与详情页重复抢跑。
func SyncTimeAutoTransferAllowed() (cred model.NetdiskCredential, allow bool) {
	cred, err := LoadNetdiskCredentials()
	if err != nil {
		return cred, false
	}
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		return cred, true
	}
	if cfg.ResourceDetailAutoTransfer {
		return cred, false
	}
	return cred, true
}

// MarkResourceTransferPending 将资源转存状态置为 pending。
func MarkResourceTransferPending(resourceID uint64, msg string) {
	if strings.TrimSpace(msg) == "" {
		msg = "等待转存"
	}
	_ = database.DB().Model(&model.Resource{}).Where("id = ?", resourceID).Updates(map[string]any{
		"transfer_status": "pending",
		"transfer_msg":    trimTo255(msg),
	}).Error
}

// ShouldAutoTransferOnCreate 根据当前网盘凭证判断是否支持自动转存。
func ShouldAutoTransferOnCreate(cred model.NetdiskCredential, link string) bool {
	link = strings.TrimSpace(link)
	if link == "" {
		return false
	}

	switch DetectTransferPlatform(link) {
	case PlatformBaidu:
		return cred.BaiduAutoSave
	case PlatformQuark:
		return cred.QuarkAutoSave
	case PlatformUC:
		return cred.UcAutoSave
	case PlatformPan115:
		return cred.Pan115AutoSave
	case PlatformTianyi:
		return cred.TianyiAutoSave
	case PlatformPan123:
		return cred.Pan123AutoSave
	case PlatformAliyun:
		return cred.AliyunAutoSave
	case PlatformXunlei:
		return cred.XunleiAutoSave
	default:
		return false
	}
}

// ShouldAutoTransferOnCreateMulti 任一分享链接匹配已开启自动转存的网盘即返回 true
func ShouldAutoTransferOnCreateMulti(cred model.NetdiskCredential, primary string, extras model.JSONStringList) bool {
	if ShouldAutoTransferOnCreate(cred, primary) {
		return true
	}
	for _, u := range extras {
		if ShouldAutoTransferOnCreate(cred, u) {
			return true
		}
	}
	return false
}

// resolveTransferSourceLinks 解析用于转存的原始分享链接（优先库内快照，其次成功日志中的 old_link，最后当前行）。
func resolveTransferSourceLinks(res *model.Resource) (primary string, extras []string) {
	if s := strings.TrimSpace(res.TransferSourceLink); s != "" {
		ex := make([]string, 0, len(res.TransferSourceExtra))
		for _, u := range res.TransferSourceExtra {
			if t := strings.TrimSpace(u); t != "" {
				ex = append(ex, t)
			}
		}
		return s, ex
	}
	var logs []model.ResourceTransferLog
	_ = database.DB().
		Where("resource_id = ? AND status = ? AND old_link != ?", res.ID, "success", "").
		Order("id ASC").
		Find(&logs).Error
	olds := make([]string, 0, len(logs))
	seen := map[string]bool{}
	for _, lg := range logs {
		u := strings.TrimSpace(lg.OldLink)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		olds = append(olds, u)
	}
	if len(olds) > 0 {
		if len(olds) == 1 {
			return olds[0], nil
		}
		return olds[0], olds[1:]
	}
	primary = strings.TrimSpace(res.Link)
	for _, u := range res.ExtraLinks {
		if t := strings.TrimSpace(u); t != "" {
			extras = append(extras, t)
		}
	}
	return primary, extras
}

// DetailPageEachClickOwnShare 详情页每次点击：从原始分享重新转存，返回新的本人分享链接，不写回 resources。
func DetailPageEachClickOwnShare(resourceID uint64, maxRetry int) (links []string, displayMsg string, err error) {
	if maxRetry < 1 {
		maxRetry = 1
	}
	var res model.Resource
	if err := database.DB().Where("id = ? AND status = 1", resourceID).First(&res).Error; err != nil {
		return nil, "", err
	}
	primary, extras := resolveTransferSourceLinks(&res)
	if strings.TrimSpace(primary) == "" {
		return nil, "", fmt.Errorf("资源无有效分享链接")
	}
	pass := strings.TrimSpace(res.ExtractCode)

	extrasWork := make([]string, len(extras))
	copy(extrasWork, extras)

	var slots []transferSlot
	if primary != "" {
		slots = append(slots, transferSlot{isPrimary: true, extraIdx: -1, link: primary})
	}
	for i, u := range extrasWork {
		if strings.TrimSpace(u) == "" {
			continue
		}
		slots = append(slots, transferSlot{isPrimary: false, extraIdx: i, link: strings.TrimSpace(u)})
	}
	var jobs []transferSlot
	for _, s := range slots {
		if DetectTransferPlatform(s.link) == PlatformUnknown {
			continue
		}
		jobs = append(jobs, s)
	}
	if len(jobs) == 0 {
		return nil, "", fmt.Errorf("无可用网盘链接（无法识别平台）")
	}

	outPrimary := ""
	extrasOut := make([]string, len(extrasWork))
	var failMsgs []string
	okN := 0
	for _, job := range jobs {
		newURL, _, msg, e := transferSingleShareLink(resourceID, job.link, pass, maxRetry)
		if e != nil {
			failMsgs = append(failMsgs, trimTo255(e.Error()))
			continue
		}
		okN++
		u := strings.TrimSpace(newURL)
		if job.isPrimary {
			outPrimary = u
		} else if job.extraIdx >= 0 && job.extraIdx < len(extrasOut) {
			extrasOut[job.extraIdx] = u
		}
		_ = msg
	}
	if okN == 0 {
		return nil, "", fmt.Errorf("%s", strings.Join(failMsgs, "; "))
	}

	links = []string{}
	if strings.TrimSpace(outPrimary) != "" {
		links = append(links, outPrimary)
	}
	for _, u := range extrasOut {
		if strings.TrimSpace(u) != "" {
			links = append(links, u)
		}
	}
	if len(links) == 0 {
		return nil, "", fmt.Errorf("未获得本人分享链接")
	}
	displayMsg = "已为你生成本次分享链接（每次点击重新转存）"
	if okN < len(jobs) {
		displayMsg = fmt.Sprintf("部分成功 %d/%d；%s", okN, len(jobs), displayMsg)
	}
	return links, displayMsg, nil
}
