package service

import (
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

// StartTelegramAutoTransferWorker：
// 补偿 TG 同步历史资源“未自动转存”的问题。
// 定时扫描 transfer_status 为空的 telegram 资源，若当前凭证开启了对应平台自动转存，则触发转存。
func StartTelegramAutoTransferWorker() {
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		// 立即跑一次，减少等待时间
		_ = scanAndTriggerTelegramAutoTransfer(5)

		for range ticker.C {
			_ = scanAndTriggerTelegramAutoTransfer(5)
		}
	}()
}

func scanAndTriggerTelegramAutoTransfer(limit int) error {
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil && cfg.ResourceDetailAutoTransfer {
		return nil
	}

	cred, err := LoadNetdiskCredentials()
	if err != nil {
		return err
	}

	db := database.DB().Model(&model.Resource{})
	var list []model.Resource
	if err := db.
		Where("status = 1").
		Where("source = ?", "telegram").
		Where("(transfer_status = '' OR transfer_status IS NULL)").
		Order("id ASC").
		Limit(limit).
		Find(&list).Error; err != nil {
		return err
	}

	for _, r := range list {
		if !ShouldAutoTransferOnCreateMulti(cred, r.Link, r.ExtraLinks) {
			continue
		}
		MarkResourceTransferPending(r.ID, "TG 扫描等待自动转存")
		rid := r.ID
		go func() {
			defer func() { recover() }()
			_ = TransferResourceWithRetry(rid, 3)
		}()
	}

	return nil
}
