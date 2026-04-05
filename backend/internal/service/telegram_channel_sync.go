package service

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

type tgGetUpdatesResp struct {
	OK     bool `json:"ok"`
	Result []struct {
		UpdateID          int64      `json:"update_id"`
		ChannelPost       *tgMessage `json:"channel_post"`
		EditedChannelPost *tgMessage `json:"edited_channel_post"`
	} `json:"result"`
}

type tgMessage struct {
	MessageID int64  `json:"message_id"`
	Text      string `json:"text"`
	Caption   string `json:"caption"`
	Photo     []struct {
		FileID string `json:"file_id"`
	} `json:"photo"`
	Chat struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	} `json:"chat"`
}

var tgURLReg = regexp.MustCompile(`https?://[^\s]+`)
var tgKVLineReg = regexp.MustCompile(`^\s*([^\s：:]+)\s*[：:]\s*(.*)$`)
var tgImageURLReg = regexp.MustCompile(`https?://[^\s]+?\.(?:jpg|jpeg|png|webp|gif)(?:\?[^\s]*)?`)
// Telegram / tgme 网页里常见整段 HTML，图片只在 <img src="..."> 中，裸露 URL 正则扫不到。
var tgImgSrcReg = regexp.MustCompile(`(?is)<img\b[^>]*\bsrc\s*=\s*["']([^"']+)["']`)

// SyncTelegramChannelByID 同步指定频道
func SyncTelegramChannelByID(channelID uint64) (int, int, error) {
	var ch model.TelegramChannel
	if err := database.DB().First(&ch, channelID).Error; err != nil {
		return 0, 0, err
	}
	return syncTelegramChannel(&ch)
}

// BackfillTelegramChannelByID 回溯同步历史消息（不受 last_update_id 限制）。
func BackfillTelegramChannelByID(channelID uint64, limit int) (int, int, int, error) {
	var ch model.TelegramChannel
	if err := database.DB().First(&ch, channelID).Error; err != nil {
		return 0, 0, 0, err
	}
	if strings.TrimSpace(ch.ChannelChatID) == "" {
		return 0, 0, 0, fmt.Errorf("频道配置不完整")
	}
	if !canUseMTProtoSync() {
		return 0, 0, 0, fmt.Errorf("回溯同步仅支持 MTProto，请先完成 api_id/api_hash + session 配置")
	}
	if limit <= 0 {
		limit = 2000
	}
	if limit > 5000 {
		limit = 5000
	}
	return backfillTelegramChannelByMTProto(&ch, limit)
}

// SyncAllEnabledTelegramChannels 同步所有启用频道
func SyncAllEnabledTelegramChannels() (int, int, int, error) {
	var channels []model.TelegramChannel
	if err := database.DB().Where("enabled = 1").Order("id ASC").Find(&channels).Error; err != nil {
		return 0, 0, 0, err
	}
	totalAdded := 0
	totalSkipped := 0
	synced := 0
	for i := range channels {
		added, skipped, err := syncTelegramChannel(&channels[i])
		if err != nil {
			continue
		}
		totalAdded += added
		totalSkipped += skipped
		synced++
	}
	return synced, totalAdded, totalSkipped, nil
}

func syncTelegramChannel(ch *model.TelegramChannel) (int, int, error) {
	if ch.ChannelChatID == "" {
		return 0, 0, fmt.Errorf("频道配置不完整")
	}

	// 优先 MTProto（api_id/api_hash + session）
	if canUseMTProtoSync() {
		added, skipped, err := syncTelegramChannelByMTProto(ch)
		if err == nil {
			return added, skipped, nil
		}
		// MTProto失败时，继续尝试 Bot API 回退
		botToken, proxyURL, botErr := resolveTelegramConnConfig(ch.BotToken, ch.ProxyURL)
		if botErr != nil {
			markChannelSync(ch.ID, "failed", err.Error(), ch.LastUpdateID)
			return 0, 0, err
		}
		return syncTelegramChannelByBot(ch, botToken, proxyURL)
	}

	// 未配置 MTProto 时走 Bot API
	botToken, proxyURL, err := resolveTelegramConnConfig(ch.BotToken, ch.ProxyURL)
	if err != nil {
		markChannelSync(ch.ID, "failed", err.Error(), ch.LastUpdateID)
		return 0, 0, err
	}
	return syncTelegramChannelByBot(ch, botToken, proxyURL)
}

func syncTelegramChannelByBot(ch *model.TelegramChannel, botToken, proxyURL string) (int, int, error) {
	chatID, err := parseTGChatID(ch.ChannelChatID)
	if err != nil {
		return 0, 0, err
	}

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?offset=%d&limit=100&timeout=0",
		botToken,
		ch.LastUpdateID+1,
	)
	client, err := newTelegramHTTPClient(proxyURL, 20*time.Second)
	if err != nil {
		markChannelSync(ch.ID, "failed", err.Error(), ch.LastUpdateID)
		return 0, 0, err
	}
	resp, err := client.Get(apiURL)
	if err != nil {
		markChannelSync(ch.ID, "failed", err.Error(), ch.LastUpdateID)
		return 0, 0, err
	}
	defer resp.Body.Close()

	var payload tgGetUpdatesResp
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		markChannelSync(ch.ID, "failed", err.Error(), ch.LastUpdateID)
		return 0, 0, err
	}
	if !payload.OK {
		err := fmt.Errorf("telegram getUpdates 返回失败")
		markChannelSync(ch.ID, "failed", err.Error(), ch.LastUpdateID)
		return 0, 0, err
	}

	added := 0
	skipped := 0
	maxUpdateID := ch.LastUpdateID

	db := database.DB()
	cred, autoTransferOnTG := SyncTimeAutoTransferAllowed()

	for _, up := range payload.Result {
		if up.UpdateID > maxUpdateID {
			maxUpdateID = up.UpdateID
		}

		msg := up.ChannelPost
		if msg == nil {
			msg = up.EditedChannelPost
		}
		if msg == nil || msg.Chat.Type != "channel" || msg.Chat.ID != chatID {
			continue
		}

		text := strings.TrimSpace(msg.Text)
		if text == "" {
			text = strings.TrimSpace(msg.Caption)
		}
		if text == "" {
			skipped++
			continue
		}

		parsed := parseTGResourceContent(text)
		link := parsed.Link
		if link == "" {
			skipped++
			continue
		}

		title := parsed.Title
		if title == "" {
			title = "TG 频道资源"
		}

		categoryID := ch.DefaultCatID
		if categoryID == 0 {
			var cat model.Category
			if err := db.Where("status = 1").Order("sort_order DESC, id ASC").First(&cat).Error; err == nil {
				categoryID = cat.ID
			}
		}
		if categoryID == 0 {
			skipped++
			continue
		}

		externalID := fmt.Sprintf("tg:%d:%d", msg.Chat.ID, msg.MessageID)
		var exists int64
		_ = db.Model(&model.Resource{}).Where("external_id = ?", externalID).Count(&exists).Error
		if exists > 0 {
			// 兼容历史数据：如果之前 TG 同步创建但未自动转存，
			// 且当前开启了自动转存，则在这里补触发一次。
			if autoTransferOnTG && ShouldAutoTransferOnCreateMulti(cred, link, model.NormalizeExtraShareLinks(parsed.ExtraLinks)) {
				var old model.Resource
				if err := db.Where("external_id = ?", externalID).First(&old).Error; err == nil {
					if strings.TrimSpace(old.TransferStatus) == "" {
						MarkResourceTransferPending(old.ID, "TG 历史等待自动转存")
						rid := old.ID
						go func() {
							defer func() { recover() }()
							_ = TransferResourceWithRetry(rid, 3)
						}()
					}
				}
			}
			continue
		}

		item := model.Resource{
			Title:       title,
			Link:        link,
			ExtraLinks:  model.NormalizeExtraShareLinks(parsed.ExtraLinks),
			CategoryID:  categoryID,
			Source:      "telegram",
			ExternalID:  externalID,
			Description: trimTGText(parsed.Description, 2000),
			ExtractCode: trimTGText(parsed.ExtractCode, 50),
			Cover: trimTGText(
				localizeCoverURL(
					client,
					firstNonEmpty(parsed.Cover, resolveBotMessageCover(client, botToken, msg, text)),
					externalID,
				),
				255,
			),
			Tags:      trimTGText(parsed.Tags, 255),
			Status:    1,
			SortOrder: 0,
		}
		if err := db.Create(&item).Error; err != nil {
			skipped++
			continue
		}
		added++
		// TG 同步新增资源后，如果开启“自动转存”，则立即标记 pending 并触发转存。
		if autoTransferOnTG && ShouldAutoTransferOnCreateMulti(cred, item.Link, item.ExtraLinks) {
			MarkResourceTransferPending(item.ID, "TG 同步等待自动转存")
			rid := item.ID
			go func() {
				defer func() { recover() }()
				_ = TransferResourceWithRetry(rid, 3)
			}()
		}
	}

	markChannelSync(ch.ID, "success", "", maxUpdateID)
	return added, skipped, nil
}

func syncTelegramChannelByMTProto(ch *model.TelegramChannel) (int, int, error) {
	cfg, err := getSystemConfig()
	if err != nil {
		return 0, 0, err
	}
	if cfg.TgAPIID <= 0 || strings.TrimSpace(cfg.TgAPIHash) == "" || strings.TrimSpace(cfg.TgSession) == "" {
		return 0, 0, fmt.Errorf("MTProto 未配置完整")
	}
	sess, err := base64.StdEncoding.DecodeString(strings.TrimSpace(cfg.TgSession))
	if err != nil {
		return 0, 0, fmt.Errorf("MTProto session 无效")
	}
	st := &mtStorage{data: sess}
	client, err := newMTProtoClient(cfg.TgAPIID, strings.TrimSpace(cfg.TgAPIHash), strings.TrimSpace(cfg.TgProxyURL), st)
	if err != nil {
		return 0, 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	added := 0
	skipped := 0
	maxMessageID := ch.LastUpdateID
	cred, autoTransferOnTG := SyncTimeAutoTransferAllowed()
	runErr := client.Run(ctx, func(ctx context.Context) error {
		peer, err := resolveChannelPeer(ctx, client.API(), strings.TrimSpace(ch.ChannelChatID))
		if err != nil {
			return err
		}
		history, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:  peer,
			Limit: 100,
		})
		if err != nil {
			return err
		}
		msgs := extractHistoryMessages(history)
		sort.Slice(msgs, func(i, j int) bool { return msgs[i].ID < msgs[j].ID })

		for _, m := range msgs {
			msgID := int64(m.ID)
			if msgID <= ch.LastUpdateID {
				continue
			}
			if msgID > maxMessageID {
				maxMessageID = msgID
			}
			text := strings.TrimSpace(m.Message)
			if text == "" {
				skipped++
				continue
			}
			parsed := parseTGResourceContent(text)
			link := parsed.Link
			if link == "" {
				skipped++
				continue
			}
			title := parsed.Title
			if title == "" {
				title = "TG 频道资源"
			}
			categoryID := resolveCategoryID(ch.DefaultCatID)
			if categoryID == 0 {
				skipped++
				continue
			}

			externalID := fmt.Sprintf("tg:mt:%s:%d", normalizeChatKey(ch.ChannelChatID), m.ID)
			var exists int64
			_ = database.DB().Model(&model.Resource{}).Where("external_id = ?", externalID).Count(&exists).Error
			if exists > 0 {
				// 兼容历史数据：如果之前 TG 同步创建但未自动转存，
				// 且当前开启了自动转存，则在这里补触发一次。
				if autoTransferOnTG && ShouldAutoTransferOnCreateMulti(cred, link, model.NormalizeExtraShareLinks(parsed.ExtraLinks)) {
					var old model.Resource
					if err := database.DB().Where("external_id = ?", externalID).First(&old).Error; err == nil {
						if strings.TrimSpace(old.TransferStatus) == "" {
							MarkResourceTransferPending(old.ID, "TG 历史等待自动转存")
							rid := old.ID
							go func() {
								defer func() { recover() }()
								_ = TransferResourceWithRetry(rid, 3)
							}()
						}
					}
				}
				continue
			}
			item := model.Resource{
				Title:       title,
				Link:        link,
				ExtraLinks:  model.NormalizeExtraShareLinks(parsed.ExtraLinks),
				CategoryID:  categoryID,
				Source:      "telegram",
				ExternalID:  externalID,
				Description: trimTGText(parsed.Description, 2000),
				ExtractCode: trimTGText(parsed.ExtractCode, 50),
				Cover: trimTGText(
					localizeCoverURL(
						nil,
						firstNonEmpty(
							resolveMTProtoMessageCover(ctx, client, m, externalID),
							parsed.Cover,
							extractTGImageURL(text),
						),
						externalID,
					),
					255,
				),
				Tags:        trimTGText(parsed.Tags, 255),
				Status:      1,
				SortOrder:   0,
			}
			if err := database.DB().Create(&item).Error; err != nil {
				skipped++
				continue
			}
			added++
			// TG 同步新增资源后，如果开启“自动转存”，则立即标记 pending 并触发转存。
			if autoTransferOnTG && ShouldAutoTransferOnCreateMulti(cred, item.Link, item.ExtraLinks) {
				MarkResourceTransferPending(item.ID, "TG 同步等待自动转存")
				rid := item.ID
				go func() {
					defer func() { recover() }()
					_ = TransferResourceWithRetry(rid, 3)
				}()
			}
		}
		return nil
	})
	if runErr != nil {
		return 0, 0, runErr
	}
	markChannelSync(ch.ID, "success", "", maxMessageID)
	return added, skipped, nil
}

func backfillTelegramChannelByMTProto(ch *model.TelegramChannel, maxCount int) (int, int, int, error) {
	cfg, err := getSystemConfig()
	if err != nil {
		return 0, 0, 0, err
	}
	if cfg.TgAPIID <= 0 || strings.TrimSpace(cfg.TgAPIHash) == "" || strings.TrimSpace(cfg.TgSession) == "" {
		return 0, 0, 0, fmt.Errorf("MTProto 未配置完整")
	}
	sess, err := base64.StdEncoding.DecodeString(strings.TrimSpace(cfg.TgSession))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("MTProto session 无效")
	}
	st := &mtStorage{data: sess}
	client, err := newMTProtoClient(cfg.TgAPIID, strings.TrimSpace(cfg.TgAPIHash), strings.TrimSpace(cfg.TgProxyURL), st)
	if err != nil {
		return 0, 0, 0, err
	}

	added := 0
	skipped := 0
	scanned := 0
	cred, autoTransferOnTG := SyncTimeAutoTransferAllowed()
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	runErr := client.Run(ctx, func(ctx context.Context) error {
		peer, err := resolveChannelPeer(ctx, client.API(), strings.TrimSpace(ch.ChannelChatID))
		if err != nil {
			return err
		}

		offsetID := 0
		for scanned < maxCount {
			batchLimit := 100
			remain := maxCount - scanned
			if remain < batchLimit {
				batchLimit = remain
			}
			history, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:     peer,
				Limit:    batchLimit,
				OffsetID: offsetID,
			})
			if err != nil {
				return err
			}
			msgs := extractHistoryMessages(history)
			if len(msgs) == 0 {
				break
			}
			sort.Slice(msgs, func(i, j int) bool { return msgs[i].ID < msgs[j].ID })

			minID := msgs[0].ID
			for _, m := range msgs {
				if m.ID < minID {
					minID = m.ID
				}
				scanned++
				text := strings.TrimSpace(m.Message)
				if text == "" {
					skipped++
					if scanned >= maxCount {
						break
					}
					continue
				}
				parsed := parseTGResourceContent(text)
				if strings.TrimSpace(parsed.Link) == "" {
					skipped++
					if scanned >= maxCount {
						break
					}
					continue
				}
				title := parsed.Title
				if title == "" {
					title = "TG 频道资源"
				}
				categoryID := resolveCategoryID(ch.DefaultCatID)
				if categoryID == 0 {
					skipped++
					if scanned >= maxCount {
						break
					}
					continue
				}
				externalID := fmt.Sprintf("tg:mt:%s:%d", normalizeChatKey(ch.ChannelChatID), m.ID)
				var exists int64
				_ = database.DB().Model(&model.Resource{}).Where("external_id = ?", externalID).Count(&exists).Error
				if exists > 0 {
					skipped++
					if scanned >= maxCount {
						break
					}
					continue
				}
				item := model.Resource{
					Title:       title,
					Link:        parsed.Link,
					ExtraLinks:  model.NormalizeExtraShareLinks(parsed.ExtraLinks),
					CategoryID:  categoryID,
					Source:      "telegram",
					ExternalID:  externalID,
					Description: trimTGText(parsed.Description, 2000),
					ExtractCode: trimTGText(parsed.ExtractCode, 50),
					Cover: trimTGText(
						localizeCoverURL(
							nil,
							firstNonEmpty(
								resolveMTProtoMessageCover(ctx, client, m, externalID),
								parsed.Cover,
								extractTGImageURL(text),
							),
							externalID,
						),
						255,
					),
					Tags:        trimTGText(parsed.Tags, 255),
					Status:      1,
					SortOrder:   0,
				}
				if err := database.DB().Create(&item).Error; err != nil {
					skipped++
					if scanned >= maxCount {
						break
					}
					continue
				}
				added++
				// TG 回溯新增资源后，如果开启“自动转存”，则立即标记 pending 并触发转存。
				if autoTransferOnTG && ShouldAutoTransferOnCreateMulti(cred, item.Link, item.ExtraLinks) {
					MarkResourceTransferPending(item.ID, "TG 回溯等待自动转存")
					rid := item.ID
					go func() {
						defer func() { recover() }()
						_ = TransferResourceWithRetry(rid, 3)
					}()
				}
				if scanned >= maxCount {
					break
				}
			}
			offsetID = minID
			if len(msgs) < batchLimit {
				break
			}
		}
		return nil
	})
	if runErr != nil {
		return 0, 0, scanned, runErr
	}
	// 回溯同步不推进 last_update_id，避免影响增量同步游标。
	markChannelSync(ch.ID, "success", "", ch.LastUpdateID)
	return added, skipped, scanned, nil
}

func markChannelSync(channelID uint64, status, msg string, lastUpdateID int64) {
	now := time.Now()
	_ = database.DB().Model(&model.TelegramChannel{}).Where("id = ?", channelID).Updates(map[string]interface{}{
		"last_sync_status": status,
		"last_sync_msg":    trimTGText(msg, 255),
		"last_update_id":   lastUpdateID,
		"last_sync_at":     &now,
	}).Error
}

func parseTGChatID(v string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("channel_chat_id 格式错误")
	}
	return id, nil
}

func extractTGFirstURL(text string) string {
	m := tgURLReg.FindString(text)
	return strings.TrimSpace(m)
}

func extractTGImageURL(text string) string {
	m := tgImageURLReg.FindString(text)
	return strings.TrimSpace(m)
}

func extractTGImgSrcURL(text string) string {
	if m := tgImgSrcReg.FindStringSubmatch(text); len(m) > 1 {
		return strings.TrimSpace(html.UnescapeString(m[1]))
	}
	return ""
}

func resolveBotMessageCover(client *http.Client, botToken string, msg *tgMessage, text string) string {
	if img := extractTGImgSrcURL(text); img != "" {
		return img
	}
	if img := extractTGImageURL(text); img != "" {
		return img
	}
	if msg == nil || len(msg.Photo) == 0 {
		return ""
	}
	fileID := strings.TrimSpace(msg.Photo[len(msg.Photo)-1].FileID)
	if fileID == "" {
		return ""
	}
	fileURL, err := getTelegramBotFileURL(client, botToken, fileID)
	if err != nil {
		return ""
	}
	return fileURL
}

// resolveMTProtoMessageCover 使用 MTProto 媒体直接下载频道消息图片到本地，并返回可访问路径
func resolveMTProtoMessageCover(ctx context.Context, client *telegram.Client, msg *tg.Message, externalID string) string {
	if client == nil || msg == nil || msg.Media == nil {
		return ""
	}
	mm, ok := msg.Media.(*tg.MessageMediaPhoto)
	if !ok || mm.Photo == nil {
		return ""
	}
	photo, ok := mm.Photo.(*tg.Photo)
	if !ok {
		return ""
	}
	base := strings.TrimSpace(externalID)
	if base == "" {
		base = fmt.Sprintf("tg:mt:%d", msg.ID)
	}
	sum := sha1.Sum([]byte(base + ":" + strconv.FormatInt(photo.ID, 10)))
	fileName := fmt.Sprintf("%x.jpg", sum[:])
	saveDir := filepath.Join("storage", "covers")
	savePath := filepath.Join(saveDir, fileName)
	relPath := "/public/covers/" + fileName
	if _, err := os.Stat(savePath); err == nil {
		return relPath
	}
	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		return ""
	}

	// 直接按 photo location 下载；FILE_REFERENCE_EXPIRED 时上层后续同步会再次尝试。
	loc := &tg.InputPhotoFileLocation{
		ID:            photo.ID,
		AccessHash:    photo.AccessHash,
		FileReference: photo.FileReference,
		ThumbSize:     "w",
	}
	dl := downloader.NewDownloader()
	dctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	if _, err := dl.Download(client.API(), loc).WithThreads(4).ToPath(dctx, savePath); err != nil {
		_ = os.Remove(savePath)
		return ""
	}
	return relPath
}

func localizeCoverURL(client *http.Client, rawURL, externalID string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" || strings.HasPrefix(strings.ToLower(rawURL), "blob:") {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}

	sum := sha1.Sum([]byte(externalID + "|" + rawURL))
	baseName := hex.EncodeToString(sum[:])
	ext := strings.ToLower(filepath.Ext(parsed.Path))
	if ext == "" {
		ext = ".jpg"
	}
	fileName := baseName + ext
	relPath := "/public/covers/" + fileName
	saveDir := filepath.Join("storage", "covers")
	savePath := filepath.Join(saveDir, fileName)

	if _, err := os.Stat(savePath); err == nil {
		return relPath
	}
	if err := os.MkdirAll(saveDir, 0o755); err != nil {
		return rawURL
	}

	httpClient := client
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := httpClient.Get(rawURL)
	if err != nil {
		return rawURL
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return rawURL
	}

	if ct := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type"))); ct != "" {
		if exts, _ := mime.ExtensionsByType(strings.Split(ct, ";")[0]); len(exts) > 0 {
			norm := strings.ToLower(exts[0])
			if norm != "" && norm != ext {
				ext = norm
				fileName = baseName + ext
				relPath = "/public/covers/" + fileName
				savePath = filepath.Join(saveDir, fileName)
			}
		}
	}
	if _, err := os.Stat(savePath); err == nil {
		return relPath
	}

	buf, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil || len(buf) == 0 {
		return rawURL
	}
	if err := os.WriteFile(savePath, buf, 0o644); err != nil {
		return rawURL
	}
	return relPath
}

func getTelegramBotFileURL(client *http.Client, botToken, fileID string) (string, error) {
	if client == nil || strings.TrimSpace(botToken) == "" || strings.TrimSpace(fileID) == "" {
		return "", fmt.Errorf("invalid args")
	}
	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getFile?file_id=%s",
		strings.TrimSpace(botToken),
		url.QueryEscape(strings.TrimSpace(fileID)),
	)
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		OK     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if !payload.OK || strings.TrimSpace(payload.Result.FilePath) == "" {
		if payload.Description != "" {
			return "", errors.New(payload.Description)
		}
		return "", fmt.Errorf("telegram getFile failed")
	}
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", strings.TrimSpace(botToken), payload.Result.FilePath), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func buildTGTitle(text string) string {
	line := strings.TrimSpace(strings.Split(text, "\n")[0])
	line = strings.ReplaceAll(line, "\r", "")
	if extractTGFirstURL(line) == line {
		lines := strings.Split(text, "\n")
		if len(lines) > 1 {
			line = strings.TrimSpace(lines[1])
		}
	}
	return trimTGText(line, 120)
}

type tgParsedResource struct {
	Title       string
	Description string
	Link        string
	ExtraLinks  []string
	ExtractCode string
	Cover       string
	Tags        string
}

// parseTGResourceContent 从频道消息中提取标题、描述、链接。
// 支持“名称/描述/链接(夸克)”键值风格，提取失败时回退到旧逻辑。
func parseTGResourceContent(text string) tgParsedResource {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")

	title := ""
	desc := ""
	link := ""
	cover := ""
	extractCode := ""
	tags := ""
	descCollecting := false
	descLines := make([]string, 0, 16)

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			if descCollecting {
				descLines = append(descLines, "")
			}
			continue
		}

		// 描述段结束标记
		if descCollecting && (isTGMetaLine(line) || tgKVLineReg.MatchString(line)) {
			descCollecting = false
		}
		if descCollecting {
			descLines = append(descLines, line)
			continue
		}

		m := tgKVLineReg.FindStringSubmatch(line)
		if len(m) == 3 {
			key := strings.TrimSpace(m[1])
			val := strings.TrimSpace(m[2])
			switch {
			case key == "名称" || key == "标题":
				if title == "" && val != "" {
					title = val
				}
			case key == "描述" || key == "简介":
				descCollecting = true
				if val != "" {
					descLines = append(descLines, val)
				}
			case key == "链接" || key == "夸克" || key == "url" || key == "URL":
				if link == "" {
					link = extractTGFirstURL(val)
				}
			case key == "提取码" || key == "提取码/密码" || key == "密码":
				if extractCode == "" && val != "" {
					extractCode = val
				}
			case key == "封面":
				if cover == "" {
					cover = extractTGImageURL(val)
				}
			case key == "标签" || key == "tags" || key == "Tags":
				if tags == "" && val != "" {
					tags = normalizeTGTags(val)
				}
			}
		}
	}

	desc = strings.TrimSpace(strings.Join(descLines, "\n"))
	if title == "" {
		title = buildTGTitle(text)
	}
	if link == "" {
		link = extractTGFirstURL(text)
	}
	if cover == "" {
		cover = extractTGImageURL(text)
	}
	if cover == "" {
		cover = extractTGImgSrcURL(text)
	}
	if tags == "" {
		tags = normalizeTGTags(strings.Join(extractHashTags(text), " "))
	}
	if desc == "" {
		desc = text
	}
	// 设计上「描述」应对应简介正文；整段 tgme HTML 入库会重复链接/图片标签，与 RSS 一致转为纯文本。
	if strings.Contains(desc, "<") {
		if plain := strings.TrimSpace(htmlToText(desc)); plain != "" {
			desc = plain
		}
	}

	ordered := mergeTGShareLinkOrder(strings.TrimSpace(link), collectTGNetdiskURLs(text))
	if len(ordered) == 0 {
		if u := strings.TrimSpace(extractTGFirstURL(text)); u != "" {
			ordered = []string{u}
		}
	}
	var extras []string
	if len(ordered) > 0 {
		link = ordered[0]
		if len(ordered) > 1 {
			extras = ordered[1:]
		}
	} else {
		link = ""
	}

	return tgParsedResource{
		Title:       trimTGText(title, 120),
		Description: desc,
		Link:        strings.TrimSpace(link),
		ExtraLinks:  extras,
		ExtractCode: trimTGText(extractCode, 50),
		Cover:       trimTGText(cover, 2048),
		Tags:        trimTGText(tags, 255),
	}
}

func trimTGShareURL(u string) string {
	u = strings.TrimSpace(strings.TrimRight(u, `,.;:!?)"'）】，。`))
	if len(u) > 500 {
		u = u[:500]
	}
	return u
}

func collectTGNetdiskURLs(text string) []string {
	matches := tgURLReg.FindAllString(text, -1)
	out := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		m = trimTGShareURL(m)
		if m == "" || !isNetdiskURL(m) {
			continue
		}
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	return out
}

// mergeTGShareLinkOrder：键值解析到的链接优先，再合并全文中的其它网盘链（去重保序）
func mergeTGShareLinkOrder(kvLink string, textNetdisks []string) []string {
	seen := make(map[string]struct{}, 8)
	var out []string
	add := func(u string) {
		u = trimTGShareURL(u)
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	if kvLink != "" {
		add(kvLink)
	}
	for _, u := range textNetdisks {
		add(u)
	}
	return out
}

func isTGMetaLine(line string) bool {
	return strings.HasPrefix(line, "📁") ||
		strings.HasPrefix(line, "🏷") ||
		strings.HasPrefix(line, "大小：") ||
		strings.HasPrefix(line, "大小:") ||
		strings.HasPrefix(line, "标签：") ||
		strings.HasPrefix(line, "标签:")
}

func extractHashTags(text string) []string {
	parts := strings.Fields(strings.ReplaceAll(text, "\n", " "))
	out := make([]string, 0, 8)
	seen := map[string]struct{}{}
	for _, p := range parts {
		if !strings.HasPrefix(p, "#") || len(p) <= 1 {
			continue
		}
		tag := strings.Trim(p, "#,.;:![](){}\"'`")
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func normalizeTGTags(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parts := strings.Fields(strings.ReplaceAll(raw, "，", " "))
	tags := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		p = strings.TrimSpace(strings.Trim(p, ","))
		p = strings.TrimPrefix(p, "#")
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		tags = append(tags, p)
	}
	return strings.Join(tags, ",")
}

func trimTGText(s string, n int) string {
	s = strings.TrimSpace(s)
	s = sanitizeForLegacyUTF8(s)
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// sanitizeForLegacyUTF8 过滤 4 字节字符（如 emoji），兼容 utf8_general_ci.
func sanitizeForLegacyUTF8(s string) string {
	if s == "" {
		return s
	}
	return strings.Map(func(r rune) rune {
		// utf8mb3/utf8_general_ci 无法存储 U+10000 以上字符。
		if r > 0xFFFF {
			return -1
		}
		return r
	}, s)
}

// TestTelegramChannelConfig 测试频道配置是否可用
func TestTelegramChannelConfig(botToken, channelChatID, proxyURL string) error {
	if strings.TrimSpace(channelChatID) == "" {
		return fmt.Errorf("channel_chat_id 不能为空")
	}
	botToken = strings.TrimSpace(botToken)

	// 优先 MTProto 测试
	if canUseMTProtoSync() {
		if err := testChannelByMTProto(strings.TrimSpace(channelChatID)); err == nil {
			return nil
		} else {
			// 若未配置任何 Bot Token，则直接返回 MTProto 真实错误，避免被 Bot 回退错误覆盖
			cfg, cfgErr := getSystemConfig()
			if cfgErr != nil || strings.TrimSpace(cfg.TgBotToken) == "" {
				return err
			}
		}
	}

	var err error
	botToken, proxyURL, err = resolveTelegramConnConfig(botToken, proxyURL)
	if err != nil {
		return err
	}
	chatID, err := parseTGChatID(channelChatID)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getChat?chat_id=%d", strings.TrimSpace(botToken), chatID)
	client, err := newTelegramHTTPClient(proxyURL, 15*time.Second)
	if err != nil {
		return err
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("telegram 响应解析失败")
	}
	if !payload.OK {
		if payload.Description != "" {
			return errors.New(payload.Description)
		}
		return fmt.Errorf("telegram 校验失败")
	}
	return nil
}

func testChannelByMTProto(channelChatID string) error {
	cfg, err := getSystemConfig()
	if err != nil {
		return err
	}
	if cfg.TgAPIID <= 0 || strings.TrimSpace(cfg.TgAPIHash) == "" || strings.TrimSpace(cfg.TgSession) == "" {
		return fmt.Errorf("MTProto 未配置完整")
	}
	sess, err := base64.StdEncoding.DecodeString(strings.TrimSpace(cfg.TgSession))
	if err != nil {
		return fmt.Errorf("MTProto session 无效")
	}
	st := &mtStorage{data: sess}
	client, err := newMTProtoClient(cfg.TgAPIID, strings.TrimSpace(cfg.TgAPIHash), strings.TrimSpace(cfg.TgProxyURL), st)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	return client.Run(ctx, func(ctx context.Context) error {
		peer, err := resolveChannelPeer(ctx, client.API(), channelChatID)
		if err != nil {
			return err
		}
		_, err = client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{Peer: peer, Limit: 1})
		return err
	})
}

func resolveTelegramConnConfig(channelBotToken, channelProxyURL string) (string, string, error) {
	botToken := strings.TrimSpace(channelBotToken)
	proxyURL := strings.TrimSpace(channelProxyURL)

	if botToken == "" || proxyURL == "" {
		var cfg model.SystemConfig
		if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil {
			if botToken == "" {
				botToken = strings.TrimSpace(cfg.TgBotToken)
			}
			if proxyURL == "" {
				proxyURL = strings.TrimSpace(cfg.TgProxyURL)
			}
		}
	}

	if botToken == "" {
		return "", "", fmt.Errorf("请先在频道填写 Bot Token 或在系统配置填写 TG 全局 Bot Token")
	}
	return botToken, proxyURL, nil
}

func canUseMTProtoSync() bool {
	cfg, err := getSystemConfig()
	if err != nil {
		return false
	}
	return cfg.TgAPIID > 0 && strings.TrimSpace(cfg.TgAPIHash) != "" && strings.TrimSpace(cfg.TgSession) != ""
}

func resolveChannelPeer(ctx context.Context, api *tg.Client, channelChatID string) (*tg.InputPeerChannel, error) {
	channelChatID = strings.TrimSpace(channelChatID)
	if channelChatID == "" {
		return nil, fmt.Errorf("channel_chat_id 不能为空")
	}
	// @username 模式
	if strings.HasPrefix(channelChatID, "@") || isLikelyUsername(channelChatID) {
		username := strings.TrimPrefix(channelChatID, "@")
		res, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: username})
		if err != nil {
			return nil, err
		}
		for _, chat := range res.Chats {
			if c, ok := chat.(*tg.Channel); ok {
				accessHash, ok := c.GetAccessHash()
				if !ok {
					continue
				}
				return &tg.InputPeerChannel{ChannelID: c.ID, AccessHash: accessHash}, nil
			}
		}
	}

	// -100xxxx 模式（从 dialogs 中匹配）
	chatID, err := parseTGChatID(channelChatID)
	if err != nil {
		return nil, err
	}
	targetID := normalizeChannelNumericID(chatID)
	dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      200,
	})
	if err != nil {
		return nil, err
	}
	for _, c := range extractDialogChannels(dialogs) {
		if c.ID != targetID {
			continue
		}
		accessHash, ok := c.GetAccessHash()
		if !ok {
			continue
		}
		return &tg.InputPeerChannel{ChannelID: c.ID, AccessHash: accessHash}, nil
	}
	return nil, fmt.Errorf("未找到频道，请确认账号已加入频道，或改用 @username")
}

func extractDialogChannels(dialogs tg.MessagesDialogsClass) []*tg.Channel {
	out := make([]*tg.Channel, 0)
	appendChats := func(chats []tg.ChatClass) {
		for _, c := range chats {
			if ch, ok := c.(*tg.Channel); ok {
				out = append(out, ch)
			}
		}
	}
	switch d := dialogs.(type) {
	case *tg.MessagesDialogs:
		appendChats(d.Chats)
	case *tg.MessagesDialogsSlice:
		appendChats(d.Chats)
	}
	return out
}

func extractHistoryMessages(history tg.MessagesMessagesClass) []*tg.Message {
	out := make([]*tg.Message, 0, 32)
	var items []tg.MessageClass
	switch h := history.(type) {
	case *tg.MessagesMessages:
		items = h.Messages
	case *tg.MessagesMessagesSlice:
		items = h.Messages
	case *tg.MessagesChannelMessages:
		items = h.Messages
	}
	for _, m := range items {
		if mm, ok := m.(*tg.Message); ok {
			out = append(out, mm)
		}
	}
	return out
}

func resolveCategoryID(defaultID uint64) uint64 {
	if defaultID != 0 {
		return defaultID
	}
	var cat model.Category
	if err := database.DB().Where("status = 1").Order("sort_order DESC, id ASC").First(&cat).Error; err == nil {
		return cat.ID
	}
	return 0
}

func isLikelyUsername(v string) bool {
	if strings.Contains(v, " ") || strings.HasPrefix(v, "-") {
		return false
	}
	_, err := strconv.ParseInt(v, 10, 64)
	return err != nil
}

func normalizeChannelNumericID(chatID int64) int64 {
	abs := int64(math.Abs(float64(chatID)))
	// Telegram channel peer id 形如 -1001234567890，真实 channel id 为 1234567890
	if abs > 1_000_000_000_000 {
		return abs - 1_000_000_000_000
	}
	return abs
}

func normalizeChatKey(v string) string {
	return strings.ReplaceAll(strings.TrimSpace(v), " ", "")
}

func newTelegramHTTPClient(proxyValue string, timeout time.Duration) (*http.Client, error) {
	proxyValue = strings.TrimSpace(proxyValue)
	if proxyValue == "" {
		return &http.Client{Timeout: timeout}, nil
	}
	parsed, err := url.Parse(proxyValue)
	if err != nil {
		return nil, fmt.Errorf("代理地址格式错误")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "socks5", "socks5h":
	default:
		return nil, fmt.Errorf("代理协议仅支持 http/https/socks5")
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(parsed),
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}, nil
}
