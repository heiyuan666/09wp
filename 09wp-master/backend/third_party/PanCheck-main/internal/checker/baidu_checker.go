package checker

import (
	"PanCheck/internal/model"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BaiduChecker 百度网盘检测器
type BaiduChecker struct {
	*BaseChecker
}

// NewBaiduChecker 创建百度网盘检测器
func NewBaiduChecker(concurrencyLimit int, timeout time.Duration) *BaiduChecker {
	return &BaiduChecker{
		BaseChecker: NewBaseChecker(model.PlatformBaidu, concurrencyLimit, timeout),
	}
}

// normalizeBaiduURL 规范化百度网盘URL，提取有效部分并进行编码
func normalizeBaiduURL(link string) (string, error) {
	cleaned := strings.TrimSpace(link)

	// 找到 https://pan.baidu.com/s/ 的位置
	startIdx := strings.Index(cleaned, "https://pan.baidu.com/s/")
	if startIdx == -1 {
		startIdx = strings.Index(cleaned, "http://pan.baidu.com/s/")
	}
	if startIdx == -1 {
		return "", fmt.Errorf("未找到有效的百度网盘URL")
	}

	// 从URL起始位置开始，找到URL结束位置（第一个空格、换行或"提取码"等关键词）
	endIdx := startIdx
	for endIdx < len(cleaned) {
		char := cleaned[endIdx]
		// 遇到空白字符，停止
		if char == ' ' || char == '\n' || char == '\r' || char == '\t' {
			break
		}
		// 检查是否遇到"提取码"等关键词
		remaining := cleaned[endIdx:]
		if strings.HasPrefix(remaining, "提取码") || strings.HasPrefix(remaining, "密码") {
			break
		}
		endIdx++
	}

	// 提取URL部分（不包含后面的额外文本）
	urlStr := cleaned[startIdx:endIdx]
	urlStr = strings.TrimSpace(urlStr)

	// 解析URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %v", err)
	}

	// 重新构建URL，确保查询参数正确编码
	// 这样即使原始URL包含未编码的字符，也会被正确编码
	normalized := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
	if parsedURL.RawQuery != "" {
		normalized += "?" + parsedURL.Query().Encode()
	}
	if parsedURL.Fragment != "" {
		normalized += "#" + parsedURL.Fragment
	}

	return normalized, nil
}

// Check 检测链接是否有效
func (c *BaiduChecker) Check(link string) (*CheckResult, error) {
	// 规范化URL（提取有效部分并进行编码）
	normalizedLink, err := normalizeBaiduURL(link)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "URL规范化失败: " + err.Error(),
			Duration:      0,
		}, nil
	}

	// 应用频率限制
	c.ApplyRateLimit()

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.GetTimeout())
	defer cancel()

	// 提取分享ID (surl)
	surl := extractBaiduShareID(normalizedLink)
	if surl == "" {
		return &CheckResult{
			Valid:         false,
			FailureReason: "无效的分享链接格式",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 提取密码
	parsedURL, err := url.Parse(normalizedLink)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "解析URL失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}
	password := parsedURL.Query().Get("pwd")

	// 计算shorturl（去掉首字符）
	shorturl := getShorturl(surl)

	// 构建/share/list API URL
	apiURL := fmt.Sprintf("https://pan.baidu.com/share/list?web=5&app_id=250528&desc=1&showempty=0&page=1&num=20&order=time&shorturl=%s&root=1&view_mode=1&channel=chunlei&web=1&clienttype=0",
		url.QueryEscape(shorturl))

	var bdclnd string

	// 如果URL中带有提取码，先验证提取码
	if password != "" {
		randsk, err := verifyPassCode(ctx, normalizedLink, shorturl, password)
		if err != nil {
			failureReason := fmt.Sprintf("验证提取码失败: %v", err)
			return &CheckResult{
				Valid:         false,
				FailureReason: failureReason,
				Duration:      time.Since(start).Milliseconds(),
				IsRateLimited: false,
			}, nil
		}
		bdclnd = randsk
	}

	// 调用/share/list API（如果有提取码则带上BDCLND，否则不带）
	result, err := callShareListAPI(ctx, apiURL, normalizedLink, bdclnd)
	if err != nil {
		log.Printf("[BaiduChecker] /share/list API请求失败: %v", err)
		return &CheckResult{
			Valid:         false,
			FailureReason: "请求失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
			IsRateLimited: true,
		}, nil
	}

	// 检查errno
	errno := result.Errno
	errMsg := result.ErrMsg

	// 如果errno=0，表示链接有效
	if errno == 0 {
		return &CheckResult{
			Valid:         true,
			FailureReason: "",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 根据errno判断失败原因
	failureReason := getFailureReason(errno, errMsg)
	isRateLimited := errno == -62 // -62表示请求接口受限

	return &CheckResult{
		Valid:         false,
		FailureReason: failureReason,
		Duration:      time.Since(start).Milliseconds(),
		IsRateLimited: isRateLimited,
	}, nil
}

// ShareListResponse /share/list API响应结构体
type ShareListResponse struct {
	Errno  float64 `json:"errno"`
	ErrMsg string  `json:"errmsg"`
	Title  string  `json:"title,omitempty"`
}

// extractBaiduShareID 从URL中提取分享ID (surl)
func extractBaiduShareID(shareURL string) string {
	parsedURL, err := url.Parse(shareURL)
	if err != nil {
		return ""
	}

	// 处理 /s/ 格式
	path := parsedURL.Path
	if strings.HasPrefix(path, "/s/") {
		surl := strings.TrimPrefix(path, "/s/")
		// 移除可能的查询参数部分（如果URL格式不规范）
		if idx := strings.Index(surl, "?"); idx != -1 {
			surl = surl[:idx]
		}
		return surl
	}

	// 处理 /share/init?surl= 格式
	if strings.HasPrefix(path, "/share/init") {
		surl := parsedURL.Query().Get("surl")
		if surl != "" {
			return surl
		}
	}

	return ""
}

// getShorturl 获取shorturl（去掉首字符的surl）
func getShorturl(surl string) string {
	if len(surl) > 1 {
		return surl[1:]
	}
	return surl
}

// callShareListAPI 调用/share/list API
func callShareListAPI(ctx context.Context, apiURL, refererURL, bdclnd string) (*ShareListResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头（参考baidu.txt）
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh,en-GB;q=0.9,en-US;q=0.8,en;q=0.7,zh-CN;q=0.6")

	// 如果提供了bdclnd，设置Cookie
	if bdclnd != "" {
		req.Header.Set("Cookie", fmt.Sprintf("BDCLND=%s", bdclnd))
	}

	// 设置Referer
	if refererURL != "" {
		req.Header.Set("Referer", refererURL)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	var result ShareListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		// 如果JSON解析失败，尝试获取errno字段
		var jsonMap map[string]interface{}
		if err2 := json.Unmarshal(body, &jsonMap); err2 == nil {
			if errno, ok := jsonMap["errno"].(float64); ok {
				result.Errno = errno
			}
			if errmsg, ok := jsonMap["errmsg"].(string); ok {
				result.ErrMsg = errmsg
			} else if errmsg, ok := jsonMap["err_msg"].(string); ok {
				result.ErrMsg = errmsg
			}
		} else {
			return nil, fmt.Errorf("解析JSON响应失败: %v, Body: %s", err, string(body))
		}
	}

	return &result, nil
}

// verifyPassCode 验证提取码
func verifyPassCode(ctx context.Context, shareURL, shorturl, password string) (string, error) {
	// 构建验证URL
	apiURL := fmt.Sprintf("https://pan.baidu.com/share/verify?surl=%s&pwd=%s", url.QueryEscape(shorturl), url.QueryEscape(password))
	reqBody := url.Values{
		"pwd":       {password},
		"vcode":     {""},
		"vcode_str": {""},
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBufferString(reqBody.Encode()))
	if err != nil {
		return "", fmt.Errorf("创建验证请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	req.Header.Set("Referer", shareURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("验证请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析验证响应失败: %v, Body: %s", err, string(body))
	}

	if errno, ok := result["errno"].(float64); !ok || errno != 0 {
		errmsg := "未知错误"
		if msg, exists := result["errmsg"].(string); exists {
			errmsg = msg
		} else if msg, exists := result["err_msg"].(string); exists {
			errmsg = msg
		}
		return "", fmt.Errorf("验证提取码失败: errno=%v, errmsg=%s", errno, errmsg)
	}

	if randsk, ok := result["randsk"].(string); ok && randsk != "" {
		return randsk, nil
	}

	return "", fmt.Errorf("验证响应格式错误，没有randsk字段")
}

// getFailureReason 根据errno获取失败原因
func getFailureReason(errno float64, errMsg string) string {
	if errMsg != "" {
		return fmt.Sprintf("分享链接无效 (errno: %.0f, err_msg: %s)", errno, errMsg)
	}

	// 根据常见错误码提供更友好的提示
	switch int(errno) {
	case -12:
		return "缺少提取码 (errno: -12)"
	case -9:
		return "提取码错误 (errno: -9)"
	case -62:
		return "请求接口受限 (errno: -62)"
	case -8:
		return "分享文件已过期 (errno: -8)"
	default:
		return fmt.Sprintf("分享链接无效 (errno: %.0f)", errno)
	}
}
