package checker

import (
	"PanCheck/internal/model"
	apphttp "PanCheck/pkg/http"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// XunleiChecker 迅雷云盘检测器
type XunleiChecker struct {
	*BaseChecker
}

// NewXunleiChecker 创建迅雷云盘检测器
func NewXunleiChecker(concurrencyLimit int, timeout time.Duration) *XunleiChecker {
	return &XunleiChecker{
		BaseChecker: NewBaseChecker(model.PlatformXunlei, concurrencyLimit, timeout),
	}
}

// extractShareID 从链接中提取 share_id
func extractShareID(shareURL string) string {
	// https://pan.xunlei.com/s/{share_id}?pwd=xxxx
	re := regexp.MustCompile(`pan\.xunlei\.com/s/([^?/#]+)`)
	m := re.FindStringSubmatch(shareURL)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// getCaptchaSign 获取验证码签名（参考参考项目的实现）
func getCaptchaSign(clientID, clientVersion, packageName, deviceID string) (timestamp, sign string) {
	timestamp = fmt.Sprint(time.Now().UnixMilli())
	str := fmt.Sprint(clientID, clientVersion, packageName, deviceID, timestamp)

	// 使用多个算法字符串进行多次 MD5 哈希
	algorithms := []string{
		"uWRwO7gPfdPB/0NfPtfQO+71",
		"F93x+qPluYy6jdgNpq+lwdH1ap6WOM+nfz8/V",
		"0HbpxvpXFsBK5CoTKam",
		"dQhzbhzFRcawnsZqRETT9AuPAJ+wTQso82mRv",
		"SAH98AmLZLRa6DB2u68sGhyiDh15guJpXhBzI",
		"unqfo7Z64Rie9RNHMOB",
		"7yxUdFADp3DOBvXdz0DPuKNVT35wqa5z0DEyEvf",
		"RBG",
		"ThTWPG5eC0UBqlbQ+04nZAptqGCdpv9o55A",
	}

	for _, algorithm := range algorithms {
		hash := md5.Sum([]byte(str + algorithm))
		str = fmt.Sprintf("%x", hash)
	}
	sign = "1." + str
	return
}

// getCaptchaToken 获取 captcha token（参考参考项目的实现）
func (c *XunleiChecker) getCaptchaToken(ctx context.Context, action string, metas map[string]interface{}) (string, error) {
	// 使用固定的设备ID和配置
	deviceID := "5505bd0cab8c9469b98e5891d9fb3e0d"
	clientID := "ZUBzD9J_XPXfn7f7"
	clientVersion := "1.10.0.2633"
	packageName := "com.xunlei.browser"
	userAgent := "ANDROID-com.xunlei.browser/1.10.0.2633 networkType/WIFI appid/22062 deviceName/Xiaomi_M2004j7ac deviceModel/M2004J7AC OSVersion/13 protocolVersion/301 platformVersion/10 sdkVersion/233100 Oauth2Client/0.9 (Linux 4_9_337-perf-sn-uotan-gd9d488809c3d3d) (JAVA 0)"

	// 获取签名
	timestamp, captchaSign := getCaptchaSign(clientID, clientVersion, packageName, deviceID)

	// 构建 meta（如果 metas 中没有 timestamp 和 captcha_sign，则添加）
	if metas == nil {
		metas = make(map[string]interface{})
	}
	metas["timestamp"] = timestamp
	metas["captcha_sign"] = captchaSign
	metas["client_version"] = clientVersion
	metas["package_name"] = packageName

	// 构建请求体
	requestBody := map[string]interface{}{
		"action":        action,
		"captcha_token": "",
		"client_id":     clientID,
		"device_id":     deviceID,
		"meta":          metas,
		"redirect_uri":  "xlaccsdk01://xunlei.com/callback?state=harbor",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 调用获取 token 的 API（正确的 API 地址）
	tokenAPI := "https://xluser-ssl.xunlei.com/v1/shield/captcha/init"
	req, err := http.NewRequestWithContext(ctx, "POST", tokenAPI, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头（参考参考项目的实现）
	req.Header.Set("Accept", "application/json;charset=UTF-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("x-device-id", deviceID)
	req.Header.Set("x-client-id", clientID)
	req.Header.Set("x-client-version", clientVersion)

	// 发送请求
	httpClient := apphttp.GetClient()
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Printf("[XunleiChecker] 获取 token 请求失败: %v", err)
		return "", fmt.Errorf("获取 token 请求失败: %v", err)
	}
	defer apphttp.CloseResponse(resp)

	// 读取响应
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[XunleiChecker] 读取 token 响应失败: %v", err)
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 自动解压 gzip/deflate
	ce := strings.ToLower(resp.Header.Get("Content-Encoding"))
	if ce == "gzip" {
		if gz, err := gzip.NewReader(bytes.NewReader(rawBody)); err == nil {
			if dec, err2 := io.ReadAll(gz); err2 == nil {
				_ = gz.Close()
				rawBody = dec
			}
		}
	} else if ce == "deflate" {
		if zr, err := zlib.NewReader(bytes.NewReader(rawBody)); err == nil {
			if dec, err2 := io.ReadAll(zr); err2 == nil {
				_ = zr.Close()
				rawBody = dec
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[XunleiChecker] Token API 响应错误，状态码: %d, 响应: %s", resp.StatusCode, string(rawBody))
		return "", fmt.Errorf("验证码token请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var tokenResp struct {
		CaptchaToken string `json:"captcha_token"`
		Url          string `json:"url"`
	}
	if err := json.Unmarshal(rawBody, &tokenResp); err != nil {
		log.Printf("[XunleiChecker] 解析 token 响应失败: %v", err)
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查是否需要验证
	if tokenResp.Url != "" {
		return "", fmt.Errorf("需要验证: %s", tokenResp.Url)
	}

	// 提取 token
	if tokenResp.CaptchaToken == "" {
		log.Printf("[XunleiChecker] 未获取到验证码token")
		return "", fmt.Errorf("未获取到验证码token")
	}

	return tokenResp.CaptchaToken, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Check 检测链接是否有效
func (c *XunleiChecker) Check(link string) (*CheckResult, error) {
	// 应用频率限制
	c.ApplyRateLimit()

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.GetTimeout())
	defer cancel()

	// 提取 share_id
	shareID := extractShareID(link)
	if shareID == "" {
		return &CheckResult{
			Valid:         false,
			FailureReason: "链接格式无效：无法提取 share_id",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 提取 pass_code（密码，可能没有）
	u, err := url.Parse(link)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "链接格式无效：无法解析 URL",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}
	q := u.Query()
	passCode := q.Get("pwd")

	// 获取 captcha token
	action := "get:/drive/v1/share"
	metas := map[string]interface{}{
		"username":       "",
		"phone_number":   "",
		"email":          "",
		"package_name":   "pan.xunlei.com",
		"client_version": "1.92.10",
		"user_id":        "0",
	}

	captchaToken, err := c.getCaptchaToken(ctx, action, metas)
	if err != nil {
		log.Printf("[XunleiChecker] 获取 captcha token 失败: %v", err)
		captchaToken = ""
	}

	// 构建 API URL
	apiURL := fmt.Sprintf("https://api-pan.xunlei.com/drive/v1/share?share_id=%s&pass_code=%s&limit=100&pass_code_token=&page_token=&thumbnail_size=SIZE_SMALL",
		url.QueryEscape(shareID), url.QueryEscape(passCode))

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return &CheckResult{
			Valid:         false,
			FailureReason: "创建请求失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 设置请求头
	req.Header.Set("Accept", "*/*")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", "https://pan.xunlei.com")
	req.Header.Set("referer", "https://pan.xunlei.com/")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("X-Client-Id", "ZUBzD9J_XPXfn7f7")
	req.Header.Set("X-Device-Id", "5505bd0cab8c9469b98e5891d9fb3e0d")
	if captchaToken != "" {
		req.Header.Set("X-Captcha-Token", captchaToken)
	}

	// 发送请求
	httpClient := apphttp.GetClient()
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		// 请求失败，可能是网络问题或链接无效
		if ctx.Err() == context.DeadlineExceeded {
			return &CheckResult{
				Valid:         true, // 超时视为有效，避免误判
				FailureReason: "",
				Duration:      time.Since(start).Milliseconds(),
			}, nil
		}
		log.Printf("[XunleiChecker] API 请求失败: %v", err)
		return &CheckResult{
			Valid:         false,
			FailureReason: "请求失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}
	defer apphttp.CloseResponse(resp)

	// 读取响应
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[XunleiChecker] 读取响应失败: %v", err)
		return &CheckResult{
			Valid:         false,
			FailureReason: "读取响应失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 自动解压 gzip/deflate
	ce := strings.ToLower(resp.Header.Get("Content-Encoding"))
	if ce == "gzip" {
		if gz, err := gzip.NewReader(bytes.NewReader(raw)); err == nil {
			if dec, err2 := io.ReadAll(gz); err2 == nil {
				_ = gz.Close()
				raw = dec
			}
		}
	} else if ce == "deflate" {
		if zr, err := zlib.NewReader(bytes.NewReader(raw)); err == nil {
			if dec, err2 := io.ReadAll(zr); err2 == nil {
				_ = zr.Close()
				raw = dec
			}
		}
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		log.Printf("[XunleiChecker] HTTP 状态码错误: %d, 响应内容: %s", resp.StatusCode, string(raw))

		// 尝试解析 JSON 响应
		var errorResp map[string]interface{}
		isRateLimited := false
		responseContent := string(raw)

		if err := json.Unmarshal(raw, &errorResp); err == nil {
			// 成功解析 JSON，检查 error_code
			if errorCode, ok := errorResp["error_code"].(float64); ok {
				// error_code 为 9 表示被限制
				if int(errorCode) == 9 {
					isRateLimited = true
				}
				// error_code 为 3 或其他值都表示失效链接（IsRateLimited 保持 false）
			}

			// 格式化 JSON 响应内容
			if formattedJSON, err := json.MarshalIndent(errorResp, "", "  "); err == nil {
				responseContent = string(formattedJSON)
			}
		}
		// 如果解析失败，使用原始响应内容

		return &CheckResult{
			Valid:         false,
			FailureReason: fmt.Sprintf("HTTP状态码: %d, 响应内容: %s", resp.StatusCode, responseContent),
			Duration:      time.Since(start).Milliseconds(),
			IsRateLimited: isRateLimited,
		}, nil
	}

	// 解析 JSON 响应
	var apiResp map[string]interface{}
	if err := json.Unmarshal(raw, &apiResp); err != nil {
		log.Printf("[XunleiChecker] 解析 JSON 响应失败: %v", err)
		return &CheckResult{
			Valid:         false,
			FailureReason: "解析响应失败: " + err.Error(),
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 检查 share_status
	shareStatus, ok := apiResp["share_status"].(string)
	if !ok {
		// 如果没有 share_status 字段，检查是否有错误信息
		if errMsg, ok := apiResp["error"].(string); ok && errMsg != "" {
			return &CheckResult{
				Valid:         false,
				FailureReason: errMsg,
				Duration:      time.Since(start).Milliseconds(),
			}, nil
		}
		// 如果既没有 share_status 也没有 error，可能是响应格式异常
		log.Printf("[XunleiChecker] 响应格式异常：缺少 share_status 字段")
		return &CheckResult{
			Valid:         false,
			FailureReason: "响应格式异常：缺少 share_status 字段",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 根据 share_status 判断有效性
	if shareStatus == "OK" {
		return &CheckResult{
			Valid:         true,
			FailureReason: "",
			Duration:      time.Since(start).Milliseconds(),
		}, nil
	}

	// 其他状态都视为无效
	shareStatusText, _ := apiResp["share_status_text"].(string)
	if shareStatusText == "" {
		shareStatusText = fmt.Sprintf("分享状态: %s", shareStatus)
	}

	return &CheckResult{
		Valid:         false,
		FailureReason: shareStatusText,
		Duration:      time.Since(start).Milliseconds(),
	}, nil
}
