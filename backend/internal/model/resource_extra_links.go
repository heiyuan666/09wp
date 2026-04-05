package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// JSONStringList 资源附加网盘链接（JSON 数组），主链接仍在 resources.link
type JSONStringList []string

func (j JSONStringList) Value() (driver.Value, error) {
	if j == nil || len(j) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal([]string(j))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONStringList) Scan(src interface{}) error {
	if src == nil {
		*j = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("JSONStringList: unsupported %T", src)
	}
	if len(data) == 0 || string(data) == "null" {
		*j = nil
		return nil
	}
	var out []string
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*j = JSONStringList(out)
	return nil
}

func (j JSONStringList) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(j))
}

func (j *JSONStringList) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*j = nil
		return nil
	}
	var out []string
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	*j = JSONStringList(out)
	return nil
}

const maxExtraShareLinks = 30
const maxShareLinkLen = 500

// AllShareLinks 主链接 + 附加链接，去重保序（用于 API 展示、检测等）
func (r Resource) AllShareLinks() []string {
	seen := make(map[string]struct{}, 8)
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if len(s) > maxShareLinkLen {
			s = s[:maxShareLinkLen]
		}
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(r.Link)
	for _, u := range r.ExtraLinks {
		add(u)
	}
	return out
}

// NormalizeExtraShareLinks 裁剪长度与条数，供写入库
func NormalizeExtraShareLinks(urls []string) JSONStringList {
	if len(urls) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(urls))
	var out []string
	for _, raw := range urls {
		s := strings.TrimSpace(raw)
		if len(s) > maxShareLinkLen {
			s = s[:maxShareLinkLen]
		}
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
		if len(out) >= maxExtraShareLinks {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return JSONStringList(out)
}
