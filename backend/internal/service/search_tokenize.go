package service

import (
	"sort"
	"strings"
	"unicode"
)

// NormalizeSearchQuery 将搜索词做基础归一化：
// - trim
// - 将常见标点/空白折叠成空格
// - 英文转小写（便于 LIKE 匹配）
func NormalizeSearchQuery(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(q))
	lastSpace := false
	for _, r := range q {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || (r >= 0x4E00 && r <= 0x9FFF) {
			lastSpace = false
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		// 其他字符（空格/标点/符号）统一当空格
		if !lastSpace {
			b.WriteRune(' ')
			lastSpace = true
		}
	}

	out := strings.Join(strings.Fields(b.String()), " ")
	return out
}

// TokenizeSearchQuery 分词搜索（轻量实现）：
// - 若用户输入里含空格：以空格分段，并对每段生成 bigram
// - 若没有空格：对整段按汉字/字母数字生成 bigram
// - 去重 + 降噪 + 按长度排序，并限制数量
func TokenizeSearchQuery(q string) []string {
	q = NormalizeSearchQuery(q)
	if q == "" {
		return nil
	}

	parts := strings.Split(q, " ")
	tokenSet := map[string]struct{}{}

	addToken := func(t string) {
		t = strings.TrimSpace(t)
		if t == "" {
			return
		}
		// 跳过过短 token（单字在 title LIKE 里命中太泛）
		if len([]rune(t)) < 2 {
			return
		}
		tokenSet[t] = struct{}{}
	}

	genBigrams := func(seg string) {
		rs := []rune(seg)
		if len(rs) < 2 {
			return
		}
		for i := 0; i < len(rs)-1; i++ {
			addToken(string(rs[i : i+2]))
		}
	}

	for _, p := range parts {
		if p == "" {
			continue
		}
		// bigram
		genBigrams(p)
		// 原段（有助于命中整串）
		addToken(p)
	}

	// 若 bigram 为空，兜底返回原 q
	if len(tokenSet) == 0 {
		addToken(q)
	}

	tokens := make([]string, 0, len(tokenSet))
	for t := range tokenSet {
		tokens = append(tokens, t)
	}
	// 按长度降序，长 token 更“精准”
	sort.Slice(tokens, func(i, j int) bool {
		return len([]rune(tokens[i])) > len([]rune(tokens[j]))
	})

	// 限制 token 数，避免 SQL 参数过多
	if len(tokens) > 8 {
		tokens = tokens[:8]
	}
	return tokens
}
