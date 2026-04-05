package http

import (
	"io"
	"net/http"
	"time"
)

var (
	client *http.Client
)

// GetClient 获取HTTP客户端单例
func GetClient() *http.Client {
	if client == nil {
		transport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			Proxy:               http.ProxyFromEnvironment,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		}

		client = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
	}
	return client
}

// SetDefaultHeaders 设置默认请求头
func SetDefaultHeaders(req *http.Request) {
	req.Header.Set("accept", "application/json;charset=UTF-8")
	req.Header.Set("accept-language", "en,zh-CN;q=0.9,zh;q=0.8")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
}

// CloseResponse 安全关闭HTTP响应
func CloseResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// 错误类型定义
type TimeoutError struct {
	Message string
}

func (e *TimeoutError) Error() string {
	return e.Message
}

type NetworkError struct {
	Message string
	Err     error
}

func (e *NetworkError) Error() string {
	return e.Message + ": " + e.Err.Error()
}

type RequestError struct {
	Message string
	Err     error
}

func (e *RequestError) Error() string {
	return e.Message + ": " + e.Err.Error()
}

type ResponseError struct {
	StatusCode int
	Message    string
}

func (e *ResponseError) Error() string {
	return e.Message
}

// RateLimitError 频率限制错误（429）
type RateLimitError struct {
	StatusCode int
	Message    string
	RetryAfter int // 建议重试延迟（秒）
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// IsTimeoutError 检查错误是否为超时错误
func IsTimeoutError(err error) bool {
	_, ok := err.(*TimeoutError)
	return ok
}

// IsRateLimitError 检查错误是否为频率限制错误（429）
func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}
