package service

import (
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

const hotKeywordMaxRunes = 64
const hotKeywordMaxBytes = 200

// NormalizeSearchKeyword 归一化热搜词（去首尾空白、合并连续空白）
func NormalizeSearchKeyword(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	fields := strings.Fields(s)
	s = strings.Join(fields, " ")
	if utf8.RuneCountInString(s) > hotKeywordMaxRunes {
		runes := []rune(s)
		s = string(runes[:hotKeywordMaxRunes])
	}
	if len(s) > hotKeywordMaxBytes {
		s = s[:hotKeywordMaxBytes]
	}
	return s
}

// RecordSearchKeyword 记录一次用户搜索（用于热搜统计；应在 goroutine 中调用）
func RecordSearchKeyword(raw string) {
	kw := NormalizeSearchKeyword(raw)
	if kw == "" {
		return
	}
	// 屏蔽词：命中则不计入热搜榜，避免前台展示
	if IsKeywordBlockedText(kw) {
		return
	}
	now := time.Now()
	err := database.DB().Exec(`
INSERT INTO search_hot_words (keyword, search_count, last_searched_at, created_at, updated_at)
VALUES (?, 1, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  search_count = search_count + 1,
  last_searched_at = VALUES(last_searched_at),
  updated_at = VALUES(updated_at)
`, kw, now, now, now).Error
	if err != nil {
		log.Printf("record search keyword: %v", err)
	}
}

// ListHotSearchKeywords 返回热搜榜（按搜索次数降序）
func ListHotSearchKeywords(limit int) ([]model.SearchHotWord, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var rows []model.SearchHotWord
	err := database.DB().Model(&model.SearchHotWord{}).
		Order("search_count DESC, last_searched_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
