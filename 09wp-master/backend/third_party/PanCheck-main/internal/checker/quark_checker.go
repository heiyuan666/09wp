package checker

import (
	"PanCheck/internal/model"
	apphttp "PanCheck/pkg/http"
	"bytes"
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

// QuarkChecker 夸克网盘检测器
type QuarkChecker struct {
	*BaseChecker
}

// NewQuarkChecker 创建夸克网盘检测器
func NewQuarkChecker(concurrencyLimit int, timeout time.Duration) *QuarkChecker {
	return &QuarkChecker{
		BaseChecker: NewBaseChecker(model.PlatformQuark, concurrencyLimit, timeout),
	}
}

// Check 检测链接是否有效
func (c *QuarkChecker) Check(link string) (*CheckResult, error) {
	// 应用频率限制
	c.ApplyRateLimit()

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.GetTimeout())
	defer cancel()

	// 提取资源ID和密码
	resourceID, passCode, err := extractParamsQuark(link)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "链接格式无效: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 发送请求
	response, err := quarkRequest(ctx, resourceID, passCode)
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

	// 检查API响应状态
	if response.Status != 200 || response.Code != 0 {
		return &CheckResult{
			Valid:         false,
			FailureReason: "分享链接失效或不存在",
			Duration:      duration,
		}, nil
	}

	// 检查是否获取到stoken
	if response.Data.Stoken == "" {
		return &CheckResult{
			Valid:         false,
			FailureReason: "分享链接无效：未获取到访问令牌",
			Duration:      duration,
		}, nil
	}

	// 使用stoken进一步验证链接有效性
	detailResponse, err := quarkDetailRequest(ctx, resourceID, response.Data.Stoken)
	if err != nil {
		if apphttp.IsTimeoutError(err) {
			return &CheckResult{
				Valid:         false,
				FailureReason: "详情验证请求超时",
				Duration:      time.Since(start).Milliseconds(),
			}, nil
		}
		return &CheckResult{
			Valid:         false,
			FailureReason: "详情验证失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 检查文件列表是否为空
	if len(detailResponse.Data.List) == 0 {
		return &CheckResult{
			Valid:         false,
			FailureReason: "分享链接无效：文件列表为空",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	return &CheckResult{
		Valid:         true,
		FailureReason: "",
		Duration:      time.Since(start).Milliseconds(),
	}, nil
}

// quarkResp 夸克API响应结构（第一个API：获取token）
type quarkResp struct {
	Status  int    `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Title       string `json:"title"`
		Stoken      string `json:"stoken"`
		Subscribed  bool   `json:"subscribed"`
		ShareType   int    `json:"share_type"`
		URLType     int    `json:"url_type"`
		ExpiredType int    `json:"expired_type"`
		ExpiredAt   int64  `json:"expired_at"`
	} `json:"data"`
}

// quarkDetailResp 夸克详情API响应结构（第二个API：获取文件列表）
type quarkDetailResp struct {
	Status  int    `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		IsOwner int           `json:"is_owner"`
		List    []interface{} `json:"list"`
	} `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
}

// quarkRequest 获取夸克网盘分享信息
func quarkRequest(ctx context.Context, resourceID string, passCode string) (*quarkResp, error) {
	apiURL := "https://drive-h.quark.cn/1/clouddrive/share/sharepage/token"

	requestBody := map[string]interface{}{
		"pwd_id":                            resourceID,
		"passcode":                          passCode,
		"support_visit_limit_private_share": true,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构造请求体失败: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	apphttp.SetDefaultHeaders(req)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://pan.quark.cn")
	req.Header.Set("referer", "https://pan.quark.cn/")

	httpClient := apphttp.GetClient()
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, &apphttp.TimeoutError{Message: "请求超时"}
		}
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer apphttp.CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response quarkResp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &response, nil
}

// quarkDetailRequest 获取夸克网盘分享详情（文件列表）
func quarkDetailRequest(ctx context.Context, resourceID string, stoken string) (*quarkDetailResp, error) {
	// 构建URL，使用url.Values确保正确的URL编码
	baseURL := "https://drive-pc.quark.cn/1/clouddrive/share/sharepage/detail"
	params := url.Values{}
	params.Set("pwd_id", resourceID)
	params.Set("stoken", stoken)
	params.Set("ver", "2")
	params.Set("pr", "ucpro")
	apiURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	apphttp.SetDefaultHeaders(req)
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("origin", "https://pan.quark.cn")
	req.Header.Set("referer", "https://pan.quark.cn/")
	req.Header.Set("pragma", "no-cache")

	httpClient := apphttp.GetClient()
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, &apphttp.TimeoutError{Message: "请求超时"}
		}
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer apphttp.CloseResponse(resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var response quarkDetailResp
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &response, nil
}

// extractParamsQuark 提取参数
// 支持多种链接格式：
// - https://pan.quark.cn/s/{pwd_id}
// - https://pan.quark.cn/s/{pwd_id}?pwd={password}
// - https://pan.quark.cn/s/{pwd_id}#/list/share
// - https://pan.quark.cn/s/{pwd_id}?pwd={password}#/list/share
// - https://pan.qoark.cn/s/{short_code} (需要重定向到pan.quark.cn获取真实pwd_id)
func extractParamsQuark(rawURL string) (resId, pwd string, err error) {
	// 支持查询参数和锚点的正则表达式
	// 匹配格式: https://pan.quark.cn/s/{pwd_id}[?pwd={password}][#{fragment}]
	// 或: https://pan.qoark.cn/s/{short_code}[?pwd={password}][#{fragment}]
	urlRegex := regexp.MustCompile(`^https://(?:pan\.quark\.cn|pan\.qoark\.cn)/s/[a-zA-Z0-9]+(?:\?[^#]*)?(?:#.*)?$`)
	if !urlRegex.MatchString(rawURL) {
		return "", "", fmt.Errorf("无效的URL格式: %s", rawURL)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("URL解析失败: %v", err)
	}

	// 支持多个域名
	supportedHosts := []string{"pan.quark.cn", "pan.qoark.cn"}
	hostSupported := false
	for _, host := range supportedHosts {
		if parsedURL.Host == host {
			hostSupported = true
			break
		}
	}
	if !hostSupported {
		return "", "", fmt.Errorf("不支持的域名: %s", parsedURL.Host)
	}

	if !strings.HasPrefix(parsedURL.Path, "/s/") {
		return "", "", fmt.Errorf("无效的路径格式: %s", parsedURL.Path)
	}

	// 如果是pan.qoark.cn域名，需要先通过重定向获取真实的pwd_id
	if parsedURL.Host == "pan.qoark.cn" {
		// 保存原始URL的查询参数（密码等）
		originalQuery := parsedURL.Query()
		pwd = strings.TrimSpace(originalQuery.Get("pwd"))

		// 创建临时context用于重定向请求
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 发起请求，跟随重定向
		redirectURL, err := followRedirect(ctx, rawURL)
		if err != nil {
			return "", "", fmt.Errorf("重定向失败: %v", err)
		}

		// 从重定向后的URL中提取真实的pwd_id
		redirectParsedURL, err := url.Parse(redirectURL)
		if err != nil {
			return "", "", fmt.Errorf("解析重定向URL失败: %v", err)
		}

		// 验证重定向后的域名是pan.quark.cn
		if redirectParsedURL.Host != "pan.quark.cn" {
			return "", "", fmt.Errorf("重定向后的域名不正确: %s", redirectParsedURL.Host)
		}

		// 从重定向后的URL路径中提取真实的pwd_id
		if !strings.HasPrefix(redirectParsedURL.Path, "/s/") {
			return "", "", fmt.Errorf("重定向后的URL路径格式无效: %s", redirectParsedURL.Path)
		}

		pathPart := strings.TrimPrefix(redirectParsedURL.Path, "/s/")
		resId = strings.TrimSpace(strings.Split(pathPart, "/")[0])
		if resId == "" {
			return "", "", fmt.Errorf("无法从重定向URL中提取有效的pwd_id")
		}

		// 如果重定向后的URL也有密码参数，优先使用重定向后的
		redirectQuery := redirectParsedURL.Query()
		if redirectPwd := strings.TrimSpace(redirectQuery.Get("pwd")); redirectPwd != "" {
			pwd = redirectPwd
		}

		return resId, pwd, nil
	}

	// 对于pan.quark.cn域名，直接提取pwd_id
	// 从路径中提取 pwd_id（资源ID）
	// 移除 "/s/" 前缀，获取 pwd_id
	// 例如: "/s/39749e1fb630" -> "39749e1fb630"
	pathPart := strings.TrimPrefix(parsedURL.Path, "/s/")
	// 移除路径中可能存在的额外部分（如斜杠等）
	resId = strings.TrimSpace(strings.Split(pathPart, "/")[0])
	if resId == "" {
		return "", "", fmt.Errorf("无法从URL路径中提取有效的pwd_id")
	}

	// 从查询参数中提取密码（如果有）
	queryParams := parsedURL.Query()
	pwd = strings.TrimSpace(queryParams.Get("pwd"))

	// 如果存在密码，验证其格式
	if pwd != "" && (len(pwd) < 2 || len(pwd) > 50) {
		return "", "", fmt.Errorf("pwd参数长度无效: %d，应在2-50字符之间", len(pwd))
	}

	return resId, pwd, nil
}

// followRedirect 跟随HTTP重定向，返回最终的URL
func followRedirect(ctx context.Context, urlStr string) (string, error) {
	// 创建一个会跟随重定向的HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
		// 默认会跟随最多10次重定向
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 如果重定向次数过多，返回错误
			if len(via) >= 10 {
				return fmt.Errorf("重定向次数过多")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	apphttp.SetDefaultHeaders(req)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", &apphttp.TimeoutError{Message: "重定向请求超时"}
		}
		return "", fmt.Errorf("请求失败: %v", err)
	}
	defer apphttp.CloseResponse(resp)

	// 返回最终重定向后的URL
	return resp.Request.URL.String(), nil
}
