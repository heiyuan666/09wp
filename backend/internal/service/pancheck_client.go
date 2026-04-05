package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/config"
)

var defaultPanCheckBaseURL = config.DefaultPanCheckBaseURL

func SetPanCheckBaseURL(v string) {
	v = strings.TrimSpace(v)
	if v == "" {
		return
	}
	defaultPanCheckBaseURL = strings.TrimRight(v, "/")
}

type PanCheckRequest struct {
	Links             []string `json:"links"`
	SelectedPlatforms []string `json:"selectedPlatforms,omitempty"`
}

type PanCheckResponse struct {
	SubmissionID int64    `json:"submission_id"`
	ValidLinks   []string `json:"valid_links"`
	InvalidLinks []string `json:"invalid_links"`
	PendingLinks []string `json:"pending_links"`
	TotalDuration float64 `json:"total_duration"`
}

func PanCheckLinks(req PanCheckRequest, baseURL string) (PanCheckResponse, error) {
	if len(req.Links) == 0 {
		return PanCheckResponse{}, fmt.Errorf("links 不能为空")
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultPanCheckBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/links/check", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return PanCheckResponse{}, fmt.Errorf("请求地址 %s 失败: %w", baseURL+"/api/v1/links/check", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(raw) > 0 {
			return PanCheckResponse{}, fmt.Errorf("PanCheck 返回异常: %s", string(raw))
		}
		return PanCheckResponse{}, fmt.Errorf("PanCheck 返回状态码: %d", resp.StatusCode)
	}

	var out PanCheckResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return PanCheckResponse{}, fmt.Errorf("解析 PanCheck 响应失败")
	}
	return out, nil
}

// PanCheckLinksWithPolling 在首轮检测后，对 pending_links 进行短轮询重试，减少 unknown。
func PanCheckLinksWithPolling(
	req PanCheckRequest,
	baseURL string,
	maxPollRounds int,
pollInterval time.Duration,
) (PanCheckResponse, error) {
	resp, err := PanCheckLinks(req, baseURL)
	if err != nil {
		return PanCheckResponse{}, err
	}
	if maxPollRounds <= 0 || len(resp.PendingLinks) == 0 {
		return resp, nil
	}
	if pollInterval <= 0 {
		pollInterval = 3 * time.Second
	}

	validSet := make(map[string]struct{}, len(resp.ValidLinks))
	invalidSet := make(map[string]struct{}, len(resp.InvalidLinks))
	pendingSet := make(map[string]struct{}, len(resp.PendingLinks))
	for _, v := range resp.ValidLinks {
		validSet[strings.TrimSpace(v)] = struct{}{}
	}
	for _, v := range resp.InvalidLinks {
		invalidSet[strings.TrimSpace(v)] = struct{}{}
	}
	for _, v := range resp.PendingLinks {
		v = strings.TrimSpace(v)
		if v != "" {
			pendingSet[v] = struct{}{}
		}
	}

	lastSubmissionID := resp.SubmissionID
	for i := 0; i < maxPollRounds && len(pendingSet) > 0; i++ {
		time.Sleep(pollInterval)
		pendingLinks := make([]string, 0, len(pendingSet))
		for v := range pendingSet {
			pendingLinks = append(pendingLinks, v)
		}
		retryResp, retryErr := PanCheckLinks(PanCheckRequest{
			Links:             pendingLinks,
			SelectedPlatforms: req.SelectedPlatforms,
		}, baseURL)
		if retryErr != nil {
			// 轮询失败不阻断整体结果，保留当前已得结果。
			break
		}
		lastSubmissionID = retryResp.SubmissionID

		for _, v := range retryResp.ValidLinks {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			validSet[v] = struct{}{}
			delete(pendingSet, v)
		}
		for _, v := range retryResp.InvalidLinks {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			invalidSet[v] = struct{}{}
			delete(pendingSet, v)
		}

		// 保留本轮仍 pending 的链接，避免平台归一化导致的遗漏。
		nextPending := map[string]struct{}{}
		for _, v := range retryResp.PendingLinks {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			if _, ok := validSet[v]; ok {
				continue
			}
			if _, ok := invalidSet[v]; ok {
				continue
			}
			nextPending[v] = struct{}{}
		}
		pendingSet = nextPending
	}

	out := PanCheckResponse{
		SubmissionID:  lastSubmissionID,
		ValidLinks:    mapKeys(validSet),
		InvalidLinks:  mapKeys(invalidSet),
		PendingLinks:  mapKeys(pendingSet),
		TotalDuration: resp.TotalDuration,
	}
	return out, nil
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

