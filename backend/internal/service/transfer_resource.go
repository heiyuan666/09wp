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

		switch platform {
		case PlatformBaidu:
			var r BaiduTransferResult
			r, e = BaiduSaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
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
		case PlatformUC:
			var r UcTransferResult
			r, e = UcSaveByShareLink(link, pass)
			own = r.OwnShareURL
			msg = r.Message
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
		default:
			e = fmt.Errorf("暂未实现该网盘转存")
		}

		attempt := i + 1
		if e == nil {
			_ = database.DB().Create(&model.ResourceTransferLog{
				ResourceID:  resourceID,
				Attempt:     attempt,
				Platform:    platform.String(),
				Status:      "success",
				Message:     strings.TrimSpace(msg),
				OldLink:     link,
				NewLink:     strings.TrimSpace(own),
				OwnShareURL: strings.TrimSpace(own),
				FilterLog:   filterLogJSON,
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
