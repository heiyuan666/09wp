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

// Pan115Checker 115网盘检测器
type Pan115Checker struct {
	*BaseChecker
}

// NewPan115Checker 创建115网盘检测器
func NewPan115Checker(concurrencyLimit int, timeout time.Duration) *Pan115Checker {
	return &Pan115Checker{
		BaseChecker: NewBaseChecker(model.PlatformPan115, concurrencyLimit, timeout),
	}
}

// Check 检测链接是否有效
func (c *Pan115Checker) Check(link string) (*CheckResult, error) {
	// 应用频率限制
	c.ApplyRateLimit()

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.GetTimeout())
	defer cancel()

	shareCode, receiveCode, err := extractParams115(link)
	if err != nil || shareCode == "" || receiveCode == "" {
		failureReason := "链接格式无效"
		if err != nil {
			failureReason += ": " + err.Error()
		} else if shareCode == "" {
			failureReason += ": 缺少分享码"
		} else if receiveCode == "" {
			failureReason += ": 缺少提取码"
		}
		return &CheckResult{
			Valid:         false,
			FailureReason: failureReason,
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	response, err := pan115Request(ctx, shareCode, receiveCode)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		if apphttp.IsTimeoutError(err) {
			return &CheckResult{
				Valid:         false,
				FailureReason: "请求超时",
				Duration:      duration,
			}, nil
		}
		return &CheckResult{
			Valid:         false,
			FailureReason: "检测失败: " + err.Error(),
			Duration:      duration,
		}, nil
	}

	if response.State && response.Errno == 0 {
		return &CheckResult{
			Valid:         true,
			FailureReason: "",
			Duration:      duration,
		}, nil
	}

	return &CheckResult{
		Valid:         false,
		FailureReason: response.Error,
		Duration:      duration,
	}, nil
}

// pan115Resp 115网盘API响应结构
type pan115Resp struct {
	State bool   `json:"state"`
	Error string `json:"error"`
	Errno int    `json:"errno"`
}

// pan115Request 发起请求
func pan115Request(ctx context.Context, shareCode, receiveCode string) (*pan115Resp, error) {
	apiURL := fmt.Sprintf("https://115cdn.com/webapi/share/snap?share_code=%s&offset=0&limit=20&receive_code=%s&cid=",
		shareCode, receiveCode)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	apphttp.SetDefaultHeaders(req)
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", fmt.Sprintf("https://115cdn.com/s/%s?password=%s&", shareCode, receiveCode))
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	httpClient := apphttp.GetClient()
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, &apphttp.TimeoutError{Message: "请求超时"}
		}
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer apphttp.CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response pan115Resp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &response, nil
}

// extractParams115 从URL中提取share_code和receive_code
func extractParams115(urlStr string) (shareCode, receiveCode string, err error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", err
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) > 0 {
		shareCode = pathParts[len(pathParts)-1]
	}

	receiveCode = parsedURL.Query().Get("password")

	if receiveCode == "" && parsedURL.Fragment != "" {
		if strings.Contains(parsedURL.Fragment, "password=") {
			fragmentParams, _ := url.ParseQuery(parsedURL.Fragment)
			receiveCode = fragmentParams.Get("password")
		}
	}

	return shareCode, receiveCode, nil
}

// unicodeToChinese 将Unicode转义序列转换为中文字符（未使用但保留）
func unicodeToChinese(text string) string {
	if text == "" {
		return text
	}

	re := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	result := re.ReplaceAllStringFunc(text, func(s string) string {
		var r rune
		fmt.Sscanf(s[2:], "%04x", &r)
		return string(r)
	})

	return result
}
