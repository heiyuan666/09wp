package service

import (
	"log"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

// StartTelegramSyncWorker 启动 TG 自动同步任务
func StartTelegramSyncWorker() {
	go func() {
		for {
			var channels []model.TelegramChannel
			if err := database.DB().Where("enabled = 1").Find(&channels).Error; err != nil {
				time.Sleep(30 * time.Second)
				continue
			}

			now := time.Now()
			for _, ch := range channels {
				interval := ch.SyncInterval
				if interval < 30 {
					interval = 30
				}
				if ch.LastSyncAt != nil && now.Sub(*ch.LastSyncAt) < time.Duration(interval)*time.Second {
					continue
				}
				added, skipped, err := SyncTelegramChannelByID(ch.ID)
				if err != nil {
					log.Printf("tg channel sync failed (id=%d): %v", ch.ID, err)
				} else if added > 0 || skipped > 0 {
					log.Printf("tg channel sync done (id=%d): added=%d skipped=%d", ch.ID, added, skipped)
				}
			}
			time.Sleep(30 * time.Second)
		}
	}()
}

