package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

type TelegramImageDownloadOptions struct {
	// ChannelChatID 支持：
	// - @username
	// - -100xxxxxxxxxx（频道ID）
	ChannelChatID string

	OutDir string

	// Limit <=0 表示不限制（最多拉到 Telegram 侧限制）
	Limit int

	// MinMessageID >0 时仅下载 msgID >= MinMessageID
	MinMessageID int

	// Threads 下载并发（downloader threads）
	Threads int
}

type TelegramImageDownloadResult struct {
	ScannedMessages int `json:"scanned_messages"`
	Downloaded      int `json:"downloaded"`
	Skipped         int `json:"skipped"`
	Failed          int `json:"failed"`
}

func DownloadTelegramChannelImages(ctx context.Context, opt TelegramImageDownloadOptions) (TelegramImageDownloadResult, error) {
	opt.ChannelChatID = strings.TrimSpace(opt.ChannelChatID)
	if opt.ChannelChatID == "" {
		return TelegramImageDownloadResult{}, fmt.Errorf("channel_chat_id 不能为空")
	}
	opt.OutDir = strings.TrimSpace(opt.OutDir)
	if opt.OutDir == "" {
		opt.OutDir = "./storage/tg-images"
	}
	if opt.Threads <= 0 {
		opt.Threads = 4
	}
	if err := os.MkdirAll(opt.OutDir, 0o755); err != nil {
		return TelegramImageDownloadResult{}, err
	}

	cfg, err := getSystemConfig()
	if err != nil {
		return TelegramImageDownloadResult{}, err
	}
	if cfg.TgAPIID <= 0 || strings.TrimSpace(cfg.TgAPIHash) == "" {
		return TelegramImageDownloadResult{}, fmt.Errorf("请先在系统配置填写 tg_api_id / tg_api_hash")
	}
	if strings.TrimSpace(cfg.TgSession) == "" {
		return TelegramImageDownloadResult{}, fmt.Errorf("请先完成 MTProto 登录，写入 tg_session")
	}
	st := &mtStorage{}
	if buf, decErr := base64.StdEncoding.DecodeString(strings.TrimSpace(cfg.TgSession)); decErr == nil {
		st.data = buf
	} else {
		return TelegramImageDownloadResult{}, fmt.Errorf("tg_session 不是有效 base64")
	}

	client, err := newMTProtoClient(cfg.TgAPIID, strings.TrimSpace(cfg.TgAPIHash), strings.TrimSpace(cfg.TgProxyURL), st)
	if err != nil {
		return TelegramImageDownloadResult{}, err
	}

	res := TelegramImageDownloadResult{}
	runErr := client.Run(ctx, func(ctx context.Context) error {
		peer, err := resolveChannelPeer(ctx, client.API(), opt.ChannelChatID)
		if err != nil {
			return err
		}

		dl := downloader.NewDownloader()
		// 批量拉历史消息
		offsetID := 0
		remain := opt.Limit
		for {
			limit := 100
			if remain > 0 && remain < limit {
				limit = remain
			}
			history, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:     peer,
				Limit:    limit,
				OffsetID: offsetID,
			})
			if err != nil {
				return err
			}
			msgs := extractHistoryMessages(history)
			if len(msgs) == 0 {
				break
			}
			for _, m := range msgs {
				res.ScannedMessages++
				if opt.MinMessageID > 0 && m.ID < opt.MinMessageID {
					continue
				}

				ok, saved, err := downloadPhotoFromMessage(ctx, client, dl, opt.OutDir, opt.Threads, m)
				if err != nil {
					res.Failed++
					continue
				}
				if !ok {
					res.Skipped++
				} else if saved {
					res.Downloaded++
				}
			}
			// 下一页：用最小 message id 做 offset
			last := msgs[len(msgs)-1]
			offsetID = last.ID
			if remain > 0 {
				remain -= len(msgs)
				if remain <= 0 {
					break
				}
			}
			// 小睡避免触发 flood
			time.Sleep(250 * time.Millisecond)
		}
		return nil
	})
	if runErr != nil {
		return res, runErr
	}
	return res, nil
}

func downloadPhotoFromMessage(
	ctx context.Context,
	client *telegram.Client,
	dl *downloader.Downloader,
	outDir string,
	threads int,
	m *tg.Message,
) (hasPhoto bool, saved bool, err error) {
	if m == nil || m.Media == nil {
		return false, false, nil
	}
	media, ok := m.Media.(*tg.MessageMediaPhoto)
	if !ok || media.Photo == nil {
		return false, false, nil
	}
	photo, ok := media.Photo.(*tg.Photo)
	if !ok {
		return false, false, nil
	}
	if len(photo.Sizes) == 0 {
		return true, false, nil
	}
	// 选择最大的 size（按 W*H）
	best := pickLargestPhotoSize(photo.Sizes)
	_ = best // 目前仅用于决定扩展名/存在性，下载 location 用 Photo 即可

	ext := ".jpg"
	filename := fmt.Sprintf("msg_%d_photo_%d%s", m.ID, photo.ID, ext)
	dst := filepath.Join(outDir, filename)
	if _, statErr := os.Stat(dst); statErr == nil {
		return true, false, nil
	}
	loc := &tg.InputPhotoFileLocation{
		ID:            photo.ID,
		AccessHash:    photo.AccessHash,
		FileReference: photo.FileReference,
		ThumbSize:     "w", // 让服务端选原图/较大图
	}
	b := dl.Download(client.API(), loc).WithThreads(threads)
	if _, err := b.ToPath(ctx, dst); err != nil {
		return true, false, err
	}
	return true, true, nil
}

func pickLargestPhotoSize(sizes []tg.PhotoSizeClass) tg.PhotoSizeClass {
	var best tg.PhotoSizeClass
	bestScore := int64(-1)
	for _, s := range sizes {
		ps, ok := s.(*tg.PhotoSize)
		if !ok {
			continue
		}
		score := int64(ps.W) * int64(ps.H)
		if score > bestScore {
			bestScore = score
			best = s
		}
	}
	if best != nil {
		return best
	}
	// 兜底返回第一个
	if len(sizes) > 0 {
		return sizes[0]
	}
	return nil
}

