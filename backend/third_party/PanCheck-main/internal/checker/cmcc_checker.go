package checker

import (
	"PanCheck/internal/model"
	apphttp "PanCheck/pkg/http"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// CMCCChecker 中国移动云盘检测器
type CMCCChecker struct {
	*BaseChecker
	sharePattern *regexp.Regexp
}

// NewCMCCChecker 创建中国移动云盘检测器
func NewCMCCChecker(concurrencyLimit int, timeout time.Duration) *CMCCChecker {
	// 支持两种格式的分享链接：
	// 1. https://yun.139.com/shareweb/#/w/i/([^&]+)
	// 2. https://caiyun.139.com/m/i?([^&]+)
	return &CMCCChecker{
		BaseChecker:  NewBaseChecker(model.PlatformCMCC, concurrencyLimit, timeout),
		sharePattern: regexp.MustCompile(`https://(?:yun\.139\.com/shareweb/#/w/i/|caiyun\.139\.com/m/i\?)([^&]+)`),
	}
}

// extractShareID 从链接中提取分享ID
func (c *CMCCChecker) extractShareID(shareURL string) string {
	matches := c.sharePattern.FindStringSubmatch(shareURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// Check 检测链接是否有效
func (c *CMCCChecker) Check(link string) (*CheckResult, error) {
	// 应用频率限制
	c.ApplyRateLimit()

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.GetTimeout())
	defer cancel()

	shareID := c.extractShareID(link)
	if shareID == "" {
		return &CheckResult{
			Valid:         false,
			FailureReason: "链接格式无效：无法提取分享ID",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	fmt.Printf("[CMCCChecker] 开始检测链接: %s, shareID: %s\n", link, shareID)

	response, err := c.getShareInfo(ctx, shareID)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		fmt.Printf("[CMCCChecker] 请求失败 (shareID=%s): %v\n", shareID, err)
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

	// 打印响应中的关键字段
	resultCode, _ := response["resultCode"].(string)
	desc, _ := response["desc"].(string)
	data := response["data"]
	fmt.Printf("[CMCCChecker] 响应结果 (shareID=%s): resultCode=%s, desc=%s, data=%v\n", shareID, resultCode, desc, data != nil)

	// 判断链接是否有效：
	// 1. resultCode 必须为 "0"（成功）
	// 2. data 不能为 null（必须包含分享信息）
	if resultCode == "0" && data != nil {
		fmt.Printf("[CMCCChecker] 链接有效 (shareID=%s)\n", shareID)
		return &CheckResult{
			Valid:         true,
			FailureReason: "",
			Duration:      duration,
		}, nil
	}

	// 获取失败原因
	failureMessage := "获取分享信息失败"
	if desc != "" {
		failureMessage = desc
	} else if resultCode != "" {
		failureMessage = fmt.Sprintf("错误码: %s", resultCode)
	}

	fmt.Printf("[CMCCChecker] 链接无效 (shareID=%s): %s\n", shareID, failureMessage)
	return &CheckResult{
		Valid:         false,
		FailureReason: failureMessage,
		Duration:      duration,
	}, nil
}

// getShareInfo 获取分享信息
func (c *CMCCChecker) getShareInfo(ctx context.Context, shareID string) (map[string]interface{}, error) {
	// 构建请求体数据
	requestData := map[string]interface{}{
		"getOutLinkInfoReq": map[string]interface{}{
			"account": "",
			"linkID":  shareID,
			"passwd":  "",
			"caSrt":   1,
			"coSrt":   1,
			"srtDr":   0,
			"bNum":    1,
			"pCaID":   "root",
			"eNum":    200,
		},
		"commonAccountInfo": map[string]interface{}{
			"account":     "",
			"accountType": 1,
		},
	}

	// 将请求数据转换为JSON字符串
	jsonStr, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %v", err)
	}

	// 加密请求数据
	encryptedData, err := chinaMobileCloudEncrypt(string(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("加密请求数据失败: %v", err)
	}

	// 将加密后的数据包装为JSON字符串
	encryptedJSON, err := json.Marshal(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("序列化加密数据失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://share-kd-njs.yun.139.com/yun-share/richlifeApp/devapp/IOutLink/getOutLinkInfoV6", strings.NewReader(string(encryptedJSON)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("hcy-cool-flag", "1")
	req.Header.Set("x-deviceinfo", "||3|12.27.0|chrome|131.0.0.0|5c7c68368f048245e1ce47f1c0f8f2d0||windows 10|1536X695|zh-CN|||")

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

	// 读取响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解密响应数据
	decryptedData, err := chinaMobileCloudDecrypt(string(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("解密响应数据失败: %v", err)
	}

	// 打印解密后的原始响应数据用于调试
	fmt.Printf("[CMCCChecker] 解密后的响应数据 (shareID=%s): %s\n", shareID, decryptedData)

	// 解析JSON响应
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(decryptedData), &response); err != nil {
		return nil, fmt.Errorf("解析JSON响应失败: %v", err)
	}

	// 打印解析后的响应结构用于调试
	fmt.Printf("[CMCCChecker] 解析后的响应结构 (shareID=%s): %+v\n", shareID, response)

	return response, nil
}
