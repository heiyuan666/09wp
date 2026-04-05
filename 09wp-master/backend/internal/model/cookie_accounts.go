package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// NetdiskCookieAccount 网盘多账号轮询项（Cookie / Token 文本）
type NetdiskCookieAccount struct {
	Name     string `json:"name"`
	Cookie   string `json:"cookie"`
	Disabled bool   `json:"disabled"`
	// TargetFolderID 可选：该账号单独的转存目录 ID（夸克/UC/115/天翼/123/阿里云 parent_file_id/迅雷 folder_id）；空则使用页面全局「转存目录」配置
	TargetFolderID string `json:"target_folder_id,omitempty"`
	// TargetPath 可选：百度网盘转存路径；空则用全局 baidu_target_path
	TargetPath string `json:"target_path,omitempty"`
}

// JSONCookieAccounts 存 MySQL TEXT，API 与前端为 JSON 数组
type JSONCookieAccounts []NetdiskCookieAccount

// ActiveCookies 返回未禁用且 cookie 非空的条目（保持配置顺序）
func (j JSONCookieAccounts) ActiveCookies() []NetdiskCookieAccount {
	if len(j) == 0 {
		return nil
	}
	var out []NetdiskCookieAccount
	for _, a := range j {
		if a.Disabled {
			continue
		}
		if strings.TrimSpace(a.Cookie) == "" {
			continue
		}
		out = append(out, a)
	}
	return out
}

// PrimaryCookie 列表中第一个可用 cookie；无则空串
func (j JSONCookieAccounts) PrimaryCookie() string {
	for _, a := range j.ActiveCookies() {
		return strings.TrimSpace(a.Cookie)
	}
	return ""
}

// Scan 实现 sql.Scanner
func (j *JSONCookieAccounts) Scan(src interface{}) error {
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
		return fmt.Errorf("JSONCookieAccounts: unsupported %T", src)
	}
	if len(data) == 0 || string(data) == "null" {
		*j = nil
		return nil
	}
	var out []NetdiskCookieAccount
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*j = JSONCookieAccounts(out)
	return nil
}

// Value 实现 driver.Valuer
func (j JSONCookieAccounts) Value() (driver.Value, error) {
	if j == nil || len(j) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal([]NetdiskCookieAccount(j))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// MarshalJSON HTTP JSON 响应
func (j JSONCookieAccounts) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]NetdiskCookieAccount(j))
}

// UnmarshalJSON 请求体绑定
func (j *JSONCookieAccounts) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*j = nil
		return nil
	}
	var out []NetdiskCookieAccount
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	*j = JSONCookieAccounts(out)
	return nil
}
