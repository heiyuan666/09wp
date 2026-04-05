package service

import (
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

const (
	keywordBlockCacheTTL = 10 * time.Second
	keywordBlockMaxRunes = 64
	keywordBlockMaxBytes = 200
)

type keywordBlockCacheState struct {
	words    []string
	loadedAt time.Time
}

var (
	keywordBlockCacheMu sync.RWMutex
	keywordBlockCache   keywordBlockCacheState
)

func normalizeKeywordBlock(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// 合并连续空白，避免录入时“看似相同但不一致”
	fields := strings.Fields(s)
	if len(fields) > 0 {
		s = strings.Join(fields, " ")
	}
	// 统一大小写
	s = strings.ToLower(s)

	if utf8.RuneCountInString(s) > keywordBlockMaxRunes {
		rs := []rune(s)
		s = string(rs[:keywordBlockMaxRunes])
	}
	if len(s) > keywordBlockMaxBytes {
		s = s[:keywordBlockMaxBytes]
	}
	return strings.TrimSpace(s)
}

// NormalizeKeywordBlockForStore 用于入库/更新时的归一化（trim / 合并空白 / toLower / 长度裁剪）
func NormalizeKeywordBlockForStore(s string) string {
	return normalizeKeywordBlock(s)
}

func ClearKeywordBlockCache() {
	keywordBlockCacheMu.Lock()
	defer keywordBlockCacheMu.Unlock()
	keywordBlockCache = keywordBlockCacheState{}
}

func ListEnabledKeywordBlocks() ([]string, error) {
	keywordBlockCacheMu.RLock()
	if keywordBlockCache.words != nil && time.Since(keywordBlockCache.loadedAt) < keywordBlockCacheTTL {
		words := append([]string(nil), keywordBlockCache.words...)
		keywordBlockCacheMu.RUnlock()
		return words, nil
	}
	keywordBlockCacheMu.RUnlock()

	keywordBlockCacheMu.Lock()
	defer keywordBlockCacheMu.Unlock()
	// Double check
	if keywordBlockCache.words != nil && time.Since(keywordBlockCache.loadedAt) < keywordBlockCacheTTL {
		return append([]string(nil), keywordBlockCache.words...), nil
	}

	var rows []model.KeywordBlock
	if err := database.DB().
		Model(&model.KeywordBlock{}).
		Where("enabled = ?", true).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		n := normalizeKeywordBlock(r.Keyword)
		if n == "" {
			continue
		}
		out = append(out, n)
	}
	keywordBlockCache = keywordBlockCacheState{
		words:    out,
		loadedAt: time.Now(),
	}
	return append([]string(nil), out...), nil
}

// IsKeywordBlockedText 判断文本是否命中屏蔽词（子串匹配）
func IsKeywordBlockedText(text string) bool {
	kw := normalizeKeywordBlock(text)
	if kw == "" {
		return false
	}
	words, err := ListEnabledKeywordBlocks()
	if err != nil {
		// 兜底：查询失败则不拦截，避免误杀
		return false
	}
	for _, w := range words {
		if w == "" {
			continue
		}
		if strings.Contains(kw, w) {
			return true
		}
	}
	return false
}

func EscapeLikePattern(s string) string {
	// 用于 SQL LIKE / NOT LIKE，搭配 ESCAPE '\'，把 %/_ 转义
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

