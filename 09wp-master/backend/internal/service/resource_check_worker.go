package service

import (
	"log"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

// StartResourceCheckWorker 启动资源链接自动检测任务
func StartResourceCheckWorker() {
	go func() {
		var lastRun time.Time
		for {
			var cfg model.SystemConfig
			if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
				time.Sleep(30 * time.Second)
				continue
			}
			if !cfg.LinkCheckEnabled {
				time.Sleep(30 * time.Second)
				continue
			}

			interval := cfg.LinkCheckInterval
			if interval < 60 {
				interval = 60
			}
			if !lastRun.IsZero() && time.Since(lastRun) < time.Duration(interval)*time.Second {
				time.Sleep(30 * time.Second)
				continue
			}

			stats, err := CheckResourceLinks(nil, nil, false)
			lastRun = time.Now()
			if err != nil {
				log.Printf("resource auto check failed: %v", err)
			} else {
				log.Printf(
					"resource auto check done: checked=%d valid=%d invalid=%d unknown=%d",
					stats.Checked, stats.Valid, stats.Invalid, stats.Unknown,
				)
			}
			time.Sleep(30 * time.Second)
		}
	}()
}

