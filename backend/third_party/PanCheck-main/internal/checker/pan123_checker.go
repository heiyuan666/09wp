package checker

import (
	"PanCheck/internal/model"
	apphttp "PanCheck/pkg/http"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Pan123Checker 123网盘检测器
type Pan123Checker struct {
	*BaseChecker
}

// NewPan123Checker 创建123网盘检测器
func NewPan123Checker(concurrencyLimit int, timeout time.Duration) *Pan123Checker {
	return &Pan123Checker{
		BaseChecker: NewBaseChecker(model.PlatformPan123, concurrencyLimit, timeout),
	}
}

// Check 检测链接是否有效
func (c *Pan123Checker) Check(link string) (*CheckResult, error) {
	// 应用频率限制
	c.ApplyRateLimit()

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.GetTimeout())
	defer cancel()

	// 提取shareKey
	shareKey, err := extractShareKey123(link)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "链接格式无效: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 调用API
	apiURL := fmt.Sprintf("https://www.123pan.com/api/share/info?shareKey=%s", shareKey)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "创建请求失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	httpClient := apphttp.GetClient()
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		if apphttp.IsTimeoutError(err) {
			return &CheckResult{
				Valid:         true, // 超时视为有效，避免误判
				FailureReason: "",
				Duration:      time.Since(start).Milliseconds(),
			}, nil
		}
		return &CheckResult{
			Valid:         true, // 请求错误也视为有效，避免误判
			FailureReason: "",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}
	defer apphttp.CloseResponse(resp)

	// 403状态码视为有效（可能是访问限制，但不是链接失效）
	if resp.StatusCode == 403 {
		return &CheckResult{
			Valid:         true,
			FailureReason: "",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &CheckResult{
			Valid:         true, // 读取错误视为有效
			FailureReason: "",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	var response pan123Resp
	if err = json.Unmarshal(body, &response); err != nil {
		return &CheckResult{
			Valid:         true, // JSON解析错误视为有效
			FailureReason: "",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 检查响应
	if response.Code == 0 || response.Data.HasPwd {
		return &CheckResult{
			Valid:         true,
			FailureReason: "",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	return &CheckResult{
		Valid:         false,
		FailureReason: "链接已失效",
		Duration:      time.Since(start).Milliseconds(),
	}, nil
}

// pan123Resp 123网盘API响应结构
type pan123Resp struct {
	Code int `json:"code"`
	Data struct {
		HasPwd bool `json:"HasPwd"`
	} `json:"data"`
}

// extractShareKey123 从URL中提取shareKey
func extractShareKey123(urlStr string) (string, error) {
	// 支持多种123网盘域名
	patterns := []string{
		`https?://(?:www\.)?(?:123684|123685|123912|123pan|123592|123865)\.com/s/([a-zA-Z0-9-]+)`,
		`https?://(?:www\.)?123pan\.cn/s/([a-zA-Z0-9-]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(urlStr)
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}

	// 如果正则匹配失败，尝试从URL路径中提取
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("URL解析失败: %v", err)
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) > 0 && pathParts[len(pathParts)-1] != "" {
		shareKey := pathParts[len(pathParts)-1]
		// 验证shareKey格式
		if len(shareKey) > 0 {
			return shareKey, nil
		}
	}

	return "", fmt.Errorf("无法从URL中提取shareKey")
}
