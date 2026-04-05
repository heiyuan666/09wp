package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"dfan-netdisk-backend/internal/config"
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/service"
)

func main() {
	chat := flag.String("chat", "", "频道/群组：支持 @username 或 -100xxxxxxxxxx")
	out := flag.String("out", "./storage/tg-images", "输出目录")
	limit := flag.Int("limit", 500, "最多扫描消息数（<=0 不限制）")
	minID := flag.Int("min_id", 0, "只下载 message_id >= min_id")
	threads := flag.Int("threads", 4, "下载并发线程数")
	timeoutSec := flag.Int("timeout", 600, "超时时间(秒)")
	flag.Parse()

	if *chat == "" {
		log.Fatalf("缺少参数 -chat")
	}

	cfg := config.Load()
	if err := database.InitMySQL(cfg.MySQLDSN); err != nil {
		log.Fatalf("init mysql failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSec)*time.Second)
	defer cancel()

	res, err := service.DownloadTelegramChannelImages(ctx, service.TelegramImageDownloadOptions{
		ChannelChatID: *chat,
		OutDir:        *out,
		Limit:         *limit,
		MinMessageID:  *minID,
		Threads:       *threads,
	})
	if err != nil {
		log.Fatalf("download failed: %v", err)
	}
	fmt.Printf("done. scanned=%d downloaded=%d skipped=%d failed=%d\n",
		res.ScannedMessages, res.Downloaded, res.Skipped, res.Failed,
	)
}

