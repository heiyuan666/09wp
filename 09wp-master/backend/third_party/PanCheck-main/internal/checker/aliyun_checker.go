package checker

import (
	"PanCheck/internal/config"
	"PanCheck/internal/model"
	apphttp "PanCheck/pkg/http"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AliyunChecker 阿里云盘检测器
type AliyunChecker struct {
	*BaseChecker
}

// NewAliyunChecker 创建阿里云盘检测器
func NewAliyunChecker(concurrencyLimit int, timeout time.Duration) *AliyunChecker {
	return &AliyunChecker{
		BaseChecker: NewBaseChecker(model.PlatformAliyun, concurrencyLimit, timeout),
	}
}

// Check 检测链接是否有效
func (c *AliyunChecker) Check(link string) (*CheckResult, error) {
	// 应用频率限制
	c.ApplyRateLimit()

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.GetTimeout())
	defer cancel()

	shareID, err := extractParamsAliPan(link)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "链接格式无效: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	_, err = aliPanRequest(ctx, shareID, c.GetRateConfig())
	duration := time.Since(start).Milliseconds()

	if err != nil {
		// 检查是否为429错误
		if apphttp.IsRateLimitError(err) {
			// 429错误返回CheckResult，标记IsRateLimited为true，以便保存到数据库
			return &CheckResult{
				Valid:         false,
				FailureReason: "API频率限制（429错误）: " + err.Error(),
				Duration:      duration,
				IsRateLimited: true,
			}, nil
		}

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

	return &CheckResult{
		Valid:         true,
		FailureReason: "",
		Duration:      duration,
	}, nil
}

// extractParamsAliPan 从URL中提取share_id
func extractParamsAliPan(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %v", err)
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) == 0 {
		return "", fmt.Errorf("URL中未找到share_id")
	}

	shareID := pathParts[len(pathParts)-1]
	if shareID == "" {
		return "", fmt.Errorf("提取的share_id为空")
	}

	return shareID, nil
}

// aliPanResp 阿里云盘API响应结构
type aliPanResp struct {
	ShareTitle string `json:"share_title"`
}

// aliPanRequest 发起API请求并获取分享信息
func aliPanRequest(ctx context.Context, shareID string, rateConfig *config.PlatformRateConfig) (*aliPanResp, error) {
	apiURL := fmt.Sprintf("https://api.aliyundrive.com/adrive/v3/share_link/get_share_by_anonymous?share_id=%s", shareID)

	requestBody := fmt.Sprintf(`{"share_id":"%s"}`, shareID)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	apphttp.SetDefaultHeaders(req)
	req.Header.Set("authorization", "")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://www.alipan.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://www.alipan.com/")
	req.Header.Set("sec-ch-ua", `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("x-canary", "client=web,app=share,version=v2.3.1")

	// 执行请求
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
		// 检查429错误
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, &apphttp.RateLimitError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body)),
			}
		}
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response aliPanResp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &response, nil
}
