package service

import (
	"PanCheck/internal/model"
	"PanCheck/internal/repository"
	"PanCheck/pkg/utils"
	"PanCheck/pkg/validator"
	"strings"
)

// LinkService 链接服务
type LinkService struct {
	submissionRepo  *repository.SubmissionRepository
	invalidLinkRepo *repository.InvalidLinkRepository
}

// NewLinkService 创建链接服务
func NewLinkService() *LinkService {
	return &LinkService{
		submissionRepo:  repository.NewSubmissionRepository(),
		invalidLinkRepo: repository.NewInvalidLinkRepository(),
	}
}

// CheckLinksRequest 检测链接请求
type CheckLinksRequest struct {
	Links             []string         `json:"links" binding:"required"`
	SelectedPlatforms []model.Platform `json:"selected_platforms"` // 选择的平台（多选），如果全部选择则等同于即时检测所有链接
}

// CheckLinksResponse 检测链接响应
type CheckLinksResponse struct {
	SubmissionID       uint     `json:"submission_id"`
	InvalidLinks       []string `json:"invalid_links"`        // 已知失效链接
	PendingLinks       []string `json:"pending_links"`        // 待检测链接
	ValidLinks         []string `json:"valid_links"`          // 有效链接（即时检测后）
	TotalDuration      *int64   `json:"total_duration"`       // 总耗时（即时检测后）
	InvalidFormatCount int      `json:"invalid_format_count"` // 不规范链接数量
	DuplicateCount     int      `json:"duplicate_count"`      // 重复链接数量
}

// CheckLinks 检测链接
func (s *LinkService) CheckLinks(req *CheckLinksRequest, clientIP string, deviceInfo utils.DeviceInfo) (*CheckLinksResponse, error) {
	// 1. 去重处理：统计重复链接数量
	linkMap := make(map[string]bool)
	uniqueLinks := make([]string, 0)
	duplicateCount := 0

	for _, link := range req.Links {
		// 规范化链接用于去重（去除首尾空格）
		normalizedLink := strings.TrimSpace(link)
		if normalizedLink == "" {
			continue
		}

		if linkMap[normalizedLink] {
			duplicateCount++
		} else {
			linkMap[normalizedLink] = true
			uniqueLinks = append(uniqueLinks, normalizedLink)
		}
	}

	// 2. 解析链接，识别平台
	linkInfos := validator.ParseLinks(uniqueLinks)

	// 统计不规范链接数量（无法识别的链接）
	invalidFormatCount := 0
	validLinkMap := make(map[string]bool)
	for _, info := range linkInfos {
		validLinkMap[info.Link] = true
	}
	for _, link := range uniqueLinks {
		if !validLinkMap[link] {
			invalidFormatCount++
		}
	}

	if len(linkInfos) == 0 {
		return &CheckLinksResponse{
			InvalidLinks:       []string{},
			PendingLinks:       []string{},
			ValidLinks:         []string{},
			InvalidFormatCount: invalidFormatCount,
			DuplicateCount:     duplicateCount,
		}, nil
	}

	// 3. 提取所有链接（使用去重后的链接）
	allLinks := make([]string, 0, len(linkInfos))
	for _, info := range linkInfos {
		allLinks = append(allLinks, info.Link)
	}

	// 3. 查询失效链接表，过滤已知失效链接
	invalidLinksFromDB, err := s.invalidLinkRepo.FindByLinks(allLinks)
	if err != nil {
		return nil, err
	}

	invalidLinkMap := make(map[string]bool)
	invalidLinks := make([]string, 0)
	for _, il := range invalidLinksFromDB {
		invalidLinkMap[il.Link] = true
		invalidLinks = append(invalidLinks, il.Link)
	}

	// 4. 过滤出待检测的链接
	pendingLinks := make([]string, 0)
	for _, link := range allLinks {
		if !invalidLinkMap[link] {
			pendingLinks = append(pendingLinks, link)
		}
	}

	// 5. 创建提交记录（使用去重后的链接）
	record := &model.SubmissionRecord{
		OriginalLinks:     model.StringArray(uniqueLinks),
		PendingLinks:      model.StringArray(pendingLinks),
		SelectedPlatforms: model.PlatformArray(req.SelectedPlatforms),
		Status:            "pending",
		TotalLinks:        len(uniqueLinks), // 总提交的链接数量（去重后）
		ClientIP:          clientIP,
		Browser:           deviceInfo.Browser,
		OS:                deviceInfo.OS,
		Device:            deviceInfo.Device,
		Language:          deviceInfo.Language,
		// Country, Region, City 由 IP 反查得到，后续会补充实现
	}

	// 保存记录
	err = s.submissionRepo.Create(record)
	if err != nil {
		return nil, err
	}

	return &CheckLinksResponse{
		SubmissionID:       record.ID,
		InvalidLinks:       invalidLinks,
		PendingLinks:       pendingLinks,
		ValidLinks:         []string{},
		InvalidFormatCount: invalidFormatCount,
		DuplicateCount:     duplicateCount,
	}, nil
}

// GetSubmission 获取提交记录
func (s *LinkService) GetSubmission(id uint) (*model.SubmissionRecord, error) {
	return s.submissionRepo.GetByID(id)
}

// ListSubmissions 分页查询提交记录
func (s *LinkService) ListSubmissions(page, pageSize int) ([]model.SubmissionRecord, int64, error) {
	return s.submissionRepo.List(page, pageSize)
}

// ListRateLimitedLinks 分页查询被限制的失效链接
func (s *LinkService) ListRateLimitedLinks(page, pageSize int, platform *model.Platform) ([]model.InvalidLink, int64, error) {
	return s.invalidLinkRepo.ListRateLimited(page, pageSize, platform)
}

// ClearRateLimitedLinks 清空所有被限制的失效链接
func (s *LinkService) ClearRateLimitedLinks() error {
	return s.invalidLinkRepo.DeleteRateLimited()
}
