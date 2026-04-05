package service

import (
	"log"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

// StartRSSSyncWorker 启动 RSS 自动抓取任务
func StartRSSSyncWorker() {
	go func() {
		for {
			var subs []model.RSSSubscription
			if err := database.DB().Where("enabled = 1").Find(&subs).Error; err != nil {
				time.Sleep(30 * time.Second)
				continue
			}

			now := time.Now()
			for _, sub := range subs {
				interval := sub.SyncInterval
				if interval < 60 {
					interval = 60
				}
				if sub.LastSyncAt != nil && now.Sub(*sub.LastSyncAt) < time.Duration(interval)*time.Second {
					continue
				}
				added, skipped, err := SyncRSSSubscriptionByID(sub.ID)
				if err != nil {
					log.Printf("rss subscription sync failed (id=%d): %v", sub.ID, err)
				} else if added > 0 || skipped > 0 {
					log.Printf("rss subscription sync done (id=%d): added=%d skipped=%d", sub.ID, added, skipped)
				}
			}
			time.Sleep(30 * time.Second)
		}
	}()
}
