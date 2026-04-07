package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func httpDoJSONWithHeaders(client *http.Client, method, endpoint string, headers map[string]string, body []byte, label string) (map[string]any, error) {
	if label == "" {
		label = "接口"
	}
	var rd io.Reader
	if len(body) > 0 {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, endpoint, rd)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s错误: %s", label, string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%s返回解析失败", label)
	}
	return out, nil
}

func trimTo255(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 255 {
		return s
	}
	return s[:255]
}

func trimTo500(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 500 {
		return s
	}
	return s[:500]
}

func httpDoJSON(client *http.Client, method, endpoint, cookie string, body []byte, label string) (map[string]any, error) {
	if label == "" {
		label = "接口"
	}
	var rd io.Reader
	if len(body) > 0 {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, endpoint, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("content-type", "application/json")
	if cookie != "" {
		req.Header.Set("cookie", cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s错误: %s", label, string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%s返回解析失败", label)
	}
	return out, nil
}

// httpDoJSONUC 调用 UC 网盘与夸克同源的 drive-h 接口：须带 drive.uc.cn 的 Origin/Referer，且 API 主机使用 drive-h.quark.cn（部分服务器无法解析 drive-h.uc.cn）。
func httpDoJSONUC(client *http.Client, method, endpoint, cookie string, body []byte, label string) (map[string]any, error) {
	if label == "" {
		label = "UC网盘"
	}
	var rd io.Reader
	if len(body) > 0 {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, endpoint, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://drive.uc.cn")
	req.Header.Set("referer", "https://drive.uc.cn/")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if cookie != "" {
		req.Header.Set("cookie", cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, ucproErrorFromHTTPRaw(label, raw)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%s返回解析失败", label)
	}
	return out, nil
}

// httpDoJSONQuarkUCAuto 夸克/UC 共用逻辑：请求发往 *.uc.cn 时使用 UC 浏览器头（与 httpDoJSONUC 一致）。
func httpDoJSONQuarkUCAuto(client *http.Client, method, endpoint, cookie string, body []byte, label string) (map[string]any, error) {
	if strings.Contains(endpoint, "uc.cn") {
		return httpDoJSONUC(client, method, endpoint, cookie, body, label)
	}
	return httpDoJSON(client, method, endpoint, cookie, body, label)
}

// httpDoJSONBearerAliyun 调用阿里云盘 Open API（Bearer + 可选 x-share-token）
func httpDoJSONBearerAliyun(client *http.Client, method, endpoint, bearer string, xShareToken string, body []byte, label string) (map[string]any, error) {
	if label == "" {
		label = "阿里云盘"
	}
	var rd io.Reader
	if len(body) > 0 {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, endpoint, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	if len(body) > 0 {
		req.Header.Set("content-type", "application/json;charset=UTF-8")
	}
	req.Header.Set("authorization", "Bearer "+bearer)
	req.Header.Set("origin", "https://www.alipan.com")
	req.Header.Set("referer", "https://www.alipan.com/")
	req.Header.Set("x-canary", "client=web,app=share,version=v2.3.1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	if xShareToken != "" {
		req.Header.Set("X-Share-Token", xShareToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		snippet := strings.TrimSpace(string(raw))
		if len(snippet) > 800 {
			snippet = snippet[:800] + "...(trunc)"
		}
		if snippet == "" {
			return nil, fmt.Errorf("%s错误: HTTP %d endpoint=%s (empty body)", label, resp.StatusCode, endpoint)
		}
		return nil, fmt.Errorf("%s错误: HTTP %d endpoint=%s: %s", label, resp.StatusCode, endpoint, snippet)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%s返回解析失败", label)
	}
	if err := aliyunBizErr(out); err != nil {
		return nil, err
	}
	return out, nil
}

func aliyunBizErr(m map[string]any) error {
	if m == nil {
		return nil
	}
	c, ok := m["code"]
	if !ok {
		return nil
	}
	msg := firstNonEmptyString(
		getAnyString(m, "message"),
		getAnyString(m, "msg"),
		getAnyString(m, "error_description"),
		getAnyString(m, "error"),
	)
	switch v := c.(type) {
	case float64:
		if v == 0 {
			return nil
		}
		if msg == "" {
			msg = fmt.Sprintf("code=%.0f", v)
		}
		return fmt.Errorf("阿里云盘: %s", msg)
	case string:
		cv := strings.TrimSpace(v)
		if cv == "" || cv == "0" || strings.EqualFold(cv, "OK") {
			return nil
		}
		if msg == "" {
			msg = cv
		}
		return fmt.Errorf("阿里云盘: %s", msg)
	default:
		if msg == "" {
			msg = fmt.Sprintf("code=%v", c)
		}
		return fmt.Errorf("阿里云盘: %s", msg)
	}
}

func firstNonEmptyString(vals ...string) string {
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func getAnyString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	default:
		return fmt.Sprintf("%v", x)
	}
}

func httpDoJSONBearer(client *http.Client, method, endpoint, bearer string, body []byte, label string) (map[string]any, error) {
	if label == "" {
		label = "接口"
	}
	var rd io.Reader
	if len(body) > 0 {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, endpoint, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	if len(body) > 0 {
		req.Header.Set("content-type", "application/json;charset=UTF-8")
	}
	req.Header.Set("origin", "https://www.123pan.com")
	req.Header.Set("referer", "https://www.123pan.com/")
	req.Header.Set("authorization", bearer)
	req.Header.Set("platform", "web")
	req.Header.Set("app-version", "3")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s错误: %s", label, string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%s返回解析失败", label)
	}
	return out, nil
}

func httpDoFormPost(client *http.Client, endpoint, cookie string, formBody string, extraHeaders map[string]string) (map[string]any, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(formBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	if cookie != "" {
		req.Header.Set("cookie", cookie)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("115接口错误: %s", string(raw))
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("115返回解析失败")
	}
	return out, nil
}

func getString(m map[string]any, path ...string) (string, bool) {
	v, ok := getAny(m, path...).(string)
	return v, ok
}

func getAny(m map[string]any, path ...string) any {
	var cur any = m
	for _, key := range path {
		next, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = next[key]
	}
	return cur
}

func getFloat(m map[string]any, path ...string) float64 {
	v := getAny(m, path...)
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	default:
		return 0
	}
}
