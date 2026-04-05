package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

type telegramGetUpdatesResp struct {
	OK     bool `json:"ok"`
	Result []struct {
		UpdateID      int64 `json:"update_id"`
		ChannelPost   *telegramMessage `json:"channel_post"`
		EditedChannelPost *telegramMessage `json:"edited_channel_post"`
	} `json:"result"`
}

type telegramMessage struct {
	MessageID int64 `json:"message_id"`
	Date      int64 `json:"date"`
	Text      string `json:"text"`
	Caption   string `json:"caption"`
	Chat      struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
		Title string `json:"title"`
	} `json:"chat"`
}

var urlReg = regexp.MustCompile(`https?://[^\s]+`)

// SyncTelegramResources 拉取 Telegram 频道消息并入库
func SyncTelegramResources() (int, int, error) {
	db := database.DB()

	var cfg model.SystemConfig
	if err := db.Order("id ASC").First(&cfg).Error; err != nil {
		return 0, 0, err
	}
	if cfg.TgBotToken == "" || cfg.TgChannelChatID == "" {
		return 0, 0, fmt.Errorf("telegram 配置不完整")
	}

	chatID, err := parseChatID(cfg.TgChannelChatID)
	if err != nil {
		return 0, 0, err
	}

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?offset=%d&limit=100&timeout=0",
		cfg.TgBotToken,
		cfg.TgLastUpdateID+1,
	)
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var payload telegramGetUpdatesResp
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, 0, err
	}
	if !payload.OK {
		return 0, 0, fmt.Errorf("telegram getUpdates 返回失败")
	}

	added := 0
	skipped := 0
	maxUpdateID := cfg.TgLastUpdateID

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

		link := extractFirstURL(text)
		if link == "" {
			skipped++
			continue
		}

		title := buildTitle(text)
		if title == "" {
			title = "TG 频道资源"
		}

		categoryID := cfg.TgDefaultCatID
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
			continue
		}

		item := model.Resource{
			Title:       title,
			Link:        link,
			CategoryID:  categoryID,
			Source:      "telegram",
			ExternalID:  externalID,
			Description: trimText(text, 2000),
			Status:      1,
			SortOrder:   0,
		}
		if err := db.Create(&item).Error; err != nil {
			skipped++
			continue
		}
		added++
	}

	if maxUpdateID > cfg.TgLastUpdateID {
		_ = db.Model(&cfg).Update("tg_last_update_id", maxUpdateID).Error
	}

	return added, skipped, nil
}

func parseChatID(v string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(strings.TrimSpace(v), "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("tg_channel_chat_id 格式错误")
	}
	return id, nil
}

func extractFirstURL(text string) string {
	m := urlReg.FindString(text)
	return strings.TrimSpace(m)
}

func buildTitle(text string) string {
	line := strings.TrimSpace(strings.Split(text, "\n")[0])
	line = strings.ReplaceAll(line, "\r", "")
	// 如果第一行是纯链接，则尝试第二行
	if extractFirstURL(line) == line {
		lines := strings.Split(text, "\n")
		if len(lines) > 1 {
			line = strings.TrimSpace(lines[1])
		}
	}
	return trimText(line, 120)
}

func trimText(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n]
}

