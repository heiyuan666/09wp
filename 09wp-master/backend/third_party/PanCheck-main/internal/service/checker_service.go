package service

import (
	"PanCheck/internal/checker"
	"PanCheck/internal/model"
	"PanCheck/internal/repository"
	"PanCheck/pkg/cache"
	apphttp "PanCheck/pkg/http"
	"PanCheck/pkg/validator"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// CheckerService 检测服务
type CheckerService struct {
	checkerFactory  *checker.CheckerFactory
	submissionRepo  *repository.SubmissionRepository
	invalidLinkRepo *repository.InvalidLinkRepository
	settingsRepo    *repository.SettingsRepository
	cacheRepo       cache.CacheRepository
	invalidTTL      int                    // 无效链接缓存过期时间（小时）
	platformTTLMap  map[model.Platform]int // 平台缓存TTL映射
	ttlMu           sync.RWMutex           // 保护TTL配置的读写锁
}

// NewCheckerService 创建检测服务
func NewCheckerService(factory *checker.CheckerFactory, cacheRepo cache.CacheRepository, invalidTTL int, platformTTLMap map[model.Platform]int) *CheckerService {
	return &CheckerService{
		checkerFactory:  factory,
		submissionRepo:  repository.NewSubmissionRepository(),
		invalidLinkRepo: repository.NewInvalidLinkRepository(),
		settingsRepo:    repository.NewSettingsRepository(),
		cacheRepo:       cacheRepo,
		invalidTTL:      invalidTTL,
		platformTTLMap:  platformTTLMap,
	}
}

// UpdateTTLConfig 更新TTL配置（用于动态更新）
func (s *CheckerService) UpdateTTLConfig(invalidTTL int, platformTTLMap map[model.Platform]int) {
	s.ttlMu.Lock()
	defer s.ttlMu.Unlock()
	s.invalidTTL = invalidTTL
	s.platformTTLMap = platformTTLMap
}

// CheckRealtime 实时检测链接
func (s *CheckerService) CheckRealtime(submissionID uint, links []string) (*model.SubmissionRecord, error) {
	log.Printf("CheckRealtime: Starting check for submission %d with %d links", submissionID, len(links))

	// 尝试原子性地将状态从 pending 更新为 checking，避免重复处理
	rowsAffected, err := s.submissionRepo.UpdateStatusToChecking(submissionID)
	if err != nil {
		log.Printf("CheckRealtime: Failed to update status to checking for submission %d: %v", submissionID, err)
		return nil, fmt.Errorf("failed to update status: %v", err)
	}
	if rowsAffected == 0 {
		log.Printf("CheckRealtime: Submission %d is already being processed or not in pending status, skipping", submissionID)
		return nil, fmt.Errorf("submission %d is already being processed", submissionID)
	}
	log.Printf("CheckRealtime: Successfully updated submission %d status to checking", submissionID)

	startTime := time.Now()

	// 如果检测失败，恢复状态为 pending，以便重试
	// 注意：只有在发生 panic 或返回错误时才恢复状态
	// 正常情况下，状态会在最后更新为 checked
	defer func() {
		if r := recover(); r != nil {
			log.Printf("CheckRealtime: Panic recovered for submission %d: %v, restoring status to pending", submissionID, r)
			// 尝试恢复状态为 pending
			record, err := s.submissionRepo.GetByID(submissionID)
			if err == nil && record.Status == "checking" {
				record.Status = "pending"
				s.submissionRepo.Update(record)
			}
		}
	}()

	// 去重处理：确保检测时使用去重后的链接
	linkMap := make(map[string]bool)
	uniqueLinks := make([]string, 0)

	for _, link := range links {
		// 规范化链接用于去重（去除首尾空格）
		normalizedLink := strings.TrimSpace(link)
		if normalizedLink == "" {
			continue
		}

		if !linkMap[normalizedLink] {
			linkMap[normalizedLink] = true
			uniqueLinks = append(uniqueLinks, normalizedLink)
		}
	}

	log.Printf("CheckRealtime: After deduplication, %d unique links", len(uniqueLinks))

	// 解析链接（包含所有链接，包括未识别的）
	linkInfos := make([]validator.LinkInfo, 0, len(uniqueLinks))
	linkInfoMap := make(map[string]validator.LinkInfo) // 用于查找

	for _, link := range uniqueLinks {
		info := validator.ParseLink(link)
		linkInfos = append(linkInfos, info)
		linkInfoMap[info.Link] = info
	}

	// 如果没有链接需要检测（所有链接都是已知失效的），直接更新状态为 checked
	if len(linkInfos) == 0 {
		log.Printf("CheckRealtime: No links to check for submission %d (all links are already known invalid), updating status to checked", submissionID)
		record, err := s.submissionRepo.GetByID(submissionID)
		if err != nil {
			log.Printf("CheckRealtime: Failed to get submission record %d: %v", submissionID, err)
			return nil, fmt.Errorf("failed to get submission record: %v", err)
		}

		duration := time.Since(startTime).Milliseconds()
		now := time.Now()
		record.ValidLinks = model.StringArray([]string{})
		record.PendingLinks = model.StringArray([]string{})
		record.Status = "checked"
		record.TotalDuration = &duration
		record.CheckedAt = &now

		if err := s.submissionRepo.Update(record); err != nil {
			log.Printf("CheckRealtime: Failed to update submission record %d: %v", submissionID, err)
			return nil, fmt.Errorf("failed to update submission record: %v", err)
		}

		log.Printf("CheckRealtime: Successfully updated submission %d status to checked", submissionID)
		return record, nil
	}

	// 按平台分组
	linksByPlatform := make(map[model.Platform][]string)
	unknownLinks := make([]string, 0)

	for _, info := range linkInfos {
		if info.Platform == model.PlatformUnknown {
			unknownLinks = append(unknownLinks, info.Link)
		} else {
			linksByPlatform[info.Platform] = append(linksByPlatform[info.Platform], info.Link)
		}
	}

	log.Printf("CheckRealtime: Grouped links by platform for submission %d, platforms: %v, unknown: %d",
		submissionID, getPlatformKeys(linksByPlatform), len(unknownLinks))

	// 并发检测
	var wg sync.WaitGroup
	var mu sync.Mutex
	validLinks := make([]string, 0)
	invalidLinks := make([]model.InvalidLink, 0)

	// 检测已知平台的链接
	for platform, platformLinks := range linksByPlatform {
		linkChecker, ok := s.checkerFactory.GetChecker(platform)
		if !ok {
			// 没有检测器，标记为无效
			log.Printf("CheckRealtime: No checker found for platform %s, marking %d links as invalid", platform, len(platformLinks))
			mu.Lock()
			for _, link := range platformLinks {
				invalidLinks = append(invalidLinks, model.InvalidLink{
					Link:          link,
					Platform:      platform,
					FailureReason: "该平台检测器未实现",
					IsRateLimited: false,
					SubmissionID:  &submissionID,
					CreatedAt:     time.Now(),
				})
			}
			mu.Unlock()
			continue
		}

		// 获取并发限制
		concurrencyLimit := linkChecker.GetConcurrencyLimit()
		if concurrencyLimit <= 0 {
			concurrencyLimit = 5 // 默认值
		}

		log.Printf("CheckRealtime: Starting check for platform %s with %d links, concurrency: %d",
			platform, len(platformLinks), concurrencyLimit)

		// 使用worker pool模式
		wg.Add(1)
		go func(p model.Platform, plinks []string, ch checker.LinkChecker, limit int) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("CheckRealtime: Panic recovered in platform %s check: %v", p, r)
				}
				wg.Done()
			}()
			log.Printf("CheckRealtime: Starting goroutine for platform %s", p)
			s.checkLinksWithConcurrency(context.Background(), ch, plinks, limit, &mu, &validLinks, &invalidLinks, submissionID)
			log.Printf("CheckRealtime: Completed goroutine for platform %s", p)
		}(platform, platformLinks, linkChecker, concurrencyLimit)
	}

	// 处理未识别的链接（标记为无效）
	if len(unknownLinks) > 0 {
		log.Printf("CheckRealtime: Processing %d unknown links for submission %d", len(unknownLinks), submissionID)
		mu.Lock()
		for _, link := range unknownLinks {
			invalidLinks = append(invalidLinks, model.InvalidLink{
				Link:          link,
				Platform:      model.PlatformUnknown,
				FailureReason: "无法识别网盘平台类型",
				IsRateLimited: false,
				SubmissionID:  &submissionID,
				CreatedAt:     time.Now(),
			})
		}
		mu.Unlock()
	}

	log.Printf("CheckRealtime: Waiting for all platform checks to complete for submission %d", submissionID)
	wg.Wait()
	log.Printf("CheckRealtime: All platform checks completed for submission %d", submissionID)

	log.Printf("CheckRealtime: Detection completed for submission %d, valid: %d, invalid: %d", submissionID, len(validLinks), len(invalidLinks))

	// 更新提交记录
	record, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		log.Printf("CheckRealtime: Failed to get submission record %d: %v", submissionID, err)
		// 如果获取记录失败，尝试恢复状态为 pending
		if record != nil {
			record.Status = "pending"
			s.submissionRepo.Update(record)
		}
		return nil, err
	}

	duration := time.Since(startTime).Milliseconds()
	now := time.Now()

	record.ValidLinks = model.StringArray(validLinks)
	// 检测完所有链接后，清空待检测链接（因为所有链接都已被处理）
	record.PendingLinks = model.StringArray([]string{})
	record.Status = "checked"
	record.TotalDuration = &duration
	record.CheckedAt = &now

	log.Printf("CheckRealtime: Updating submission record %d, status: checked", submissionID)
	err = s.submissionRepo.Update(record)
	if err != nil {
		log.Printf("CheckRealtime: Failed to update submission record %d: %v", submissionID, err)
		// 如果更新失败，尝试恢复状态为 pending
		record.Status = "pending"
		s.submissionRepo.Update(record)
		return nil, err
	}

	log.Printf("CheckRealtime: Successfully updated submission record %d", submissionID)

	// 保存失效链接
	log.Printf("CheckRealtime: Saving %d invalid links for submission %d", len(invalidLinks), submissionID)
	for _, il := range invalidLinks {
		err = s.invalidLinkRepo.CreateOrUpdate(&il)
		if err != nil {
			// 记录错误但不影响主流程
			log.Printf("CheckRealtime: Failed to save invalid link %s: %v", il.Link, err)
		}
	}

	log.Printf("CheckRealtime: Successfully completed check for submission %d", submissionID)
	return record, nil
}

// CheckRealtimeWithPlatformFilter 实时检测链接（按平台过滤）
func (s *CheckerService) CheckRealtimeWithPlatformFilter(submissionID uint, links []string, selectedPlatforms []model.Platform) (*model.SubmissionRecord, error) {
	log.Printf("CheckRealtimeWithPlatformFilter: Starting check for submission %d with %d links, selected platforms: %v", submissionID, len(links), selectedPlatforms)

	// 尝试原子性地将状态从 pending 更新为 checking，避免重复处理
	rowsAffected, err := s.submissionRepo.UpdateStatusToChecking(submissionID)
	if err != nil {
		log.Printf("CheckRealtimeWithPlatformFilter: Failed to update status to checking for submission %d: %v", submissionID, err)
		return nil, fmt.Errorf("failed to update status: %v", err)
	}
	if rowsAffected == 0 {
		log.Printf("CheckRealtimeWithPlatformFilter: Submission %d is already being processed or not in pending status, skipping", submissionID)
		return nil, fmt.Errorf("submission %d is already being processed", submissionID)
	}
	log.Printf("CheckRealtimeWithPlatformFilter: Successfully updated submission %d status to checking", submissionID)

	startTime := time.Now()

	// 去重处理：确保检测时使用去重后的链接
	linkMap := make(map[string]bool)
	uniqueLinks := make([]string, 0)

	for _, link := range links {
		// 规范化链接用于去重（去除首尾空格）
		normalizedLink := strings.TrimSpace(link)
		if normalizedLink == "" {
			continue
		}

		if !linkMap[normalizedLink] {
			linkMap[normalizedLink] = true
			uniqueLinks = append(uniqueLinks, normalizedLink)
		}
	}

	log.Printf("CheckRealtimeWithPlatformFilter: After deduplication, %d unique links", len(uniqueLinks))

	// 解析链接（包含所有链接，包括未识别的）
	linkInfos := make([]validator.LinkInfo, 0, len(uniqueLinks))
	linkInfoMap := make(map[string]validator.LinkInfo) // 用于查找

	for _, link := range uniqueLinks {
		info := validator.ParseLink(link)
		linkInfos = append(linkInfos, info)
		linkInfoMap[info.Link] = info
	}

	// 如果没有链接需要检测，直接更新状态为 checked
	// 注意：即使没有链接需要检测，也可能有未选中平台的链接需要保留在 pending_links 中
	if len(linkInfos) == 0 {
		log.Printf("CheckRealtimeWithPlatformFilter: No links to check for submission %d, updating status to checked", submissionID)
		record, err := s.submissionRepo.GetByID(submissionID)
		if err != nil {
			log.Printf("CheckRealtimeWithPlatformFilter: Failed to get submission record %d: %v", submissionID, err)
			return nil, fmt.Errorf("failed to get submission record: %v", err)
		}

		duration := time.Since(startTime).Milliseconds()
		now := time.Now()
		record.ValidLinks = model.StringArray([]string{})
		record.PendingLinks = model.StringArray([]string{})
		record.Status = "checked"
		record.TotalDuration = &duration
		record.CheckedAt = &now

		if err := s.submissionRepo.Update(record); err != nil {
			log.Printf("CheckRealtimeWithPlatformFilter: Failed to update submission record %d: %v", submissionID, err)
			return nil, fmt.Errorf("failed to update submission record: %v", err)
		}

		log.Printf("CheckRealtimeWithPlatformFilter: Successfully updated submission %d status to checked", submissionID)
		return record, nil
	}

	// 构建选中平台的映射
	selectedPlatformMap := make(map[model.Platform]bool)
	for _, platform := range selectedPlatforms {
		selectedPlatformMap[platform] = true
	}

	// 按平台分组，只检测选中的平台
	linksByPlatform := make(map[model.Platform][]string)
	unknownLinks := make([]string, 0)
	skippedLinks := make([]string, 0) // 未选中平台的链接

	for _, info := range linkInfos {
		if info.Platform == model.PlatformUnknown {
			unknownLinks = append(unknownLinks, info.Link)
		} else if selectedPlatformMap[info.Platform] {
			// 只处理选中的平台
			linksByPlatform[info.Platform] = append(linksByPlatform[info.Platform], info.Link)
		} else {
			// 未选中的平台链接，跳过检测
			skippedLinks = append(skippedLinks, info.Link)
		}
	}

	// 并发检测
	var wg sync.WaitGroup
	var mu sync.Mutex
	validLinks := make([]string, 0)
	invalidLinks := make([]model.InvalidLink, 0)

	// 检测选中平台的链接
	for platform, platformLinks := range linksByPlatform {
		linkChecker, ok := s.checkerFactory.GetChecker(platform)
		if !ok {
			// 没有检测器，标记为无效
			mu.Lock()
			for _, link := range platformLinks {
				invalidLinks = append(invalidLinks, model.InvalidLink{
					Link:          link,
					Platform:      platform,
					FailureReason: "该平台检测器未实现",
					IsRateLimited: false,
					SubmissionID:  &submissionID,
					CreatedAt:     time.Now(),
				})
			}
			mu.Unlock()
			continue
		}

		// 获取并发限制
		concurrencyLimit := linkChecker.GetConcurrencyLimit()
		if concurrencyLimit <= 0 {
			concurrencyLimit = 5 // 默认值
		}

		// 使用worker pool模式
		wg.Add(1)
		go func(p model.Platform, plinks []string, ch checker.LinkChecker, limit int) {
			defer wg.Done()
			s.checkLinksWithConcurrency(context.Background(), ch, plinks, limit, &mu, &validLinks, &invalidLinks, submissionID)
		}(platform, platformLinks, linkChecker, concurrencyLimit)
	}

	// 处理未识别的链接（保留在待检测列表中，不标记为无效）
	// 在 local_only 模式下，未识别的链接应该保留在 pending_links 中

	wg.Wait()

	log.Printf("CheckRealtimeWithPlatformFilter: Detection completed for submission %d, valid: %d, invalid: %d, skipped: %d, unknown: %d",
		submissionID, len(validLinks), len(invalidLinks), len(skippedLinks), len(unknownLinks))

	// 更新提交记录
	record, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		log.Printf("CheckRealtimeWithPlatformFilter: Failed to get submission record %d: %v", submissionID, err)
		return nil, err
	}

	duration := time.Since(startTime).Milliseconds()
	now := time.Now()

	// 更新有效链接（只包含检测过的）
	record.ValidLinks = model.StringArray(validLinks)
	// 更新待检测链接：保留未选中平台的链接、未识别的链接
	remainingPendingLinks := make([]string, 0)
	remainingPendingLinks = append(remainingPendingLinks, skippedLinks...)
	remainingPendingLinks = append(remainingPendingLinks, unknownLinks...)
	record.PendingLinks = model.StringArray(remainingPendingLinks)

	// 如果还有待检测的链接，状态保持为 pending；否则设置为 checked
	if len(remainingPendingLinks) > 0 {
		record.Status = "pending"
		log.Printf("CheckRealtimeWithPlatformFilter: Updating submission record %d, status: pending (remaining %d links to check)", submissionID, len(remainingPendingLinks))
	} else {
		record.Status = "checked"
		record.CheckedAt = &now
		log.Printf("CheckRealtimeWithPlatformFilter: Updating submission record %d, status: checked", submissionID)
	}

	record.TotalDuration = &duration
	err = s.submissionRepo.Update(record)
	if err != nil {
		log.Printf("CheckRealtimeWithPlatformFilter: Failed to update submission record %d: %v", submissionID, err)
		return nil, err
	}

	log.Printf("CheckRealtimeWithPlatformFilter: Successfully updated submission record %d", submissionID)

	// 保存失效链接
	log.Printf("CheckRealtimeWithPlatformFilter: Saving %d invalid links for submission %d", len(invalidLinks), submissionID)
	for _, il := range invalidLinks {
		err = s.invalidLinkRepo.CreateOrUpdate(&il)
		if err != nil {
			// 记录错误但不影响主流程
			log.Printf("CheckRealtimeWithPlatformFilter: Failed to save invalid link %s: %v", il.Link, err)
		}
	}

	log.Printf("CheckRealtimeWithPlatformFilter: Successfully completed check for submission %d", submissionID)
	return record, nil
}

// GetInvalidLinksFromRecord 从提交记录中获取失效链接列表
func (s *CheckerService) GetInvalidLinksFromRecord(submissionID uint) ([]string, error) {
	record, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return nil, err
	}

	// 从数据库中查询该提交记录相关的失效链接
	allLinks := append([]string(record.OriginalLinks), []string(record.PendingLinks)...)
	allLinks = append(allLinks, []string(record.ValidLinks)...)

	invalidLinksFromDB, err := s.invalidLinkRepo.FindByLinks(allLinks)
	if err != nil {
		return nil, err
	}

	invalidLinks := make([]string, 0, len(invalidLinksFromDB))
	for _, il := range invalidLinksFromDB {
		invalidLinks = append(invalidLinks, il.Link)
	}

	return invalidLinks, nil
}

// checkLinksWithConcurrency 使用并发控制检测链接
func (s *CheckerService) checkLinksWithConcurrency(
	ctx context.Context,
	ch checker.LinkChecker,
	links []string,
	concurrencyLimit int,
	mu *sync.Mutex,
	validLinks *[]string,
	invalidLinks *[]model.InvalidLink,
	submissionID uint,
) {
	platform := ch.GetPlatform()
	log.Printf("checkLinksWithConcurrency: Starting check for platform %s, submission %d, links: %d, concurrency: %d",
		platform, submissionID, len(links), concurrencyLimit)

	startTime := time.Now()
	// 创建worker pool
	sem := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup
	checkedCount := 0
	var checkedMu sync.Mutex

	for _, link := range links {
		wg.Add(1)
		sem <- struct{}{} // 获取信号量

		go func(l string) {
			defer func() {
				wg.Done()
				<-sem // 释放信号量
				checkedMu.Lock()
				checkedCount++
				currentCount := checkedCount
				checkedMu.Unlock()
				// 每检测10个链接输出一次进度
				if currentCount%10 == 0 || currentCount == len(links) {
					log.Printf("checkLinksWithConcurrency: Platform %s, submission %d, progress: %d/%d",
						platform, submissionID, currentCount, len(links))
				}
			}()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("checkLinksWithConcurrency: Panic recovered while checking link %s (platform %s, submission %d): %v",
						l, platform, submissionID, r)
				}
			}()

			// 规范化链接
			linkInfo := validator.ParseLink(l)
			normalizedLink := linkInfo.Link

			var result *checker.CheckResult
			var err error
			var fromCache bool

			// 1. 先查询Redis缓存
			if s.cacheRepo != nil && s.cacheRepo.IsEnabled() {
				cachedResult, cacheErr := s.cacheRepo.Get(ctx, normalizedLink)
				if cacheErr == nil && cachedResult != nil {
					result = cachedResult
					fromCache = true
					log.Printf("checkLinksWithConcurrency: Cache hit for link %s (platform %s)", normalizedLink, platform)
				}
			}

			// 2. 缓存未命中，查询无效链接数据库
			if result == nil {
				exists, dbErr := s.invalidLinkRepo.Exists(normalizedLink)
				if dbErr == nil && exists {
					// 从数据库查询详细信息
					invalidLinks, dbErr := s.invalidLinkRepo.FindByLinks([]string{normalizedLink})
					if dbErr == nil && len(invalidLinks) > 0 {
						il := invalidLinks[0]
						var duration int64
						if il.CheckDuration != nil {
							duration = *il.CheckDuration
						}
						result = &checker.CheckResult{
							Valid:         false,
							FailureReason: il.FailureReason,
							Duration:      duration,
							IsRateLimited: il.IsRateLimited,
						}
						log.Printf("checkLinksWithConcurrency: Found invalid link in database: %s (platform %s)", normalizedLink, platform)
					}
				}
			}

			// 3. 缓存和数据库都未命中，调用网盘接口检测
			if result == nil {
				result, err = ch.Check(l)
				if err != nil {
					// 检查是否为429错误（频率限制错误）
					if apphttp.IsRateLimitError(err) {
						// 429错误也要保存到invalid_links表，标记IsRateLimited为true
						var duration int64
						if result != nil {
							duration = result.Duration
						}
						mu.Lock()
						*invalidLinks = append(*invalidLinks, model.InvalidLink{
							Link:          normalizedLink,
							Platform:      ch.GetPlatform(),
							FailureReason: fmt.Sprintf("API频率限制（429错误）: %v", err),
							CheckDuration: &duration,
							IsRateLimited: true,
							SubmissionID:  &submissionID,
							CreatedAt:     time.Now(),
						})
						mu.Unlock()
						log.Printf("检测链接时遇到频率限制（429错误），已保存到数据库: %s, 平台: %s, 错误: %v", normalizedLink, ch.GetPlatform(), err)

						// 存入缓存
						if s.cacheRepo != nil && s.cacheRepo.IsEnabled() {
							rateLimitResult := &checker.CheckResult{
								Valid:         false,
								FailureReason: fmt.Sprintf("API频率限制（429错误）: %v", err),
								Duration:      duration,
								IsRateLimited: true,
							}
							s.ttlMu.RLock()
							invalidTTL := s.invalidTTL
							platformTTLMap := s.platformTTLMap
							s.ttlMu.RUnlock()
							if cacheErr := s.cacheRepo.Set(ctx, normalizedLink, rateLimitResult, platform, invalidTTL, platformTTLMap); cacheErr != nil {
								log.Printf("Failed to cache rate limit result for link %s: %v", normalizedLink, cacheErr)
							}
						}
						return
					}

					// 其他检测错误，视为无效
					var duration int64
					var isRateLimited bool
					if result != nil {
						duration = result.Duration
						isRateLimited = result.IsRateLimited
					}
					mu.Lock()
					*invalidLinks = append(*invalidLinks, model.InvalidLink{
						Link:          normalizedLink,
						Platform:      ch.GetPlatform(),
						FailureReason: fmt.Sprintf("检测错误: %v", err),
						CheckDuration: &duration,
						IsRateLimited: isRateLimited,
						SubmissionID:  &submissionID,
						CreatedAt:     time.Now(),
					})
					mu.Unlock()

					// 存入缓存
					if s.cacheRepo != nil && s.cacheRepo.IsEnabled() {
						errorResult := &checker.CheckResult{
							Valid:         false,
							FailureReason: fmt.Sprintf("检测错误: %v", err),
							Duration:      duration,
							IsRateLimited: isRateLimited,
						}
						s.ttlMu.RLock()
						invalidTTL := s.invalidTTL
						platformTTLMap := s.platformTTLMap
						s.ttlMu.RUnlock()
						if cacheErr := s.cacheRepo.Set(ctx, normalizedLink, errorResult, platform, invalidTTL, platformTTLMap); cacheErr != nil {
							log.Printf("Failed to cache error result for link %s: %v", normalizedLink, cacheErr)
						}
					}
					return
				}
			}

			// 4. 将检测结果存入Redis缓存（如果未从缓存获取）
			if !fromCache && s.cacheRepo != nil && s.cacheRepo.IsEnabled() && result != nil {
				s.ttlMu.RLock()
				invalidTTL := s.invalidTTL
				platformTTLMap := s.platformTTLMap
				s.ttlMu.RUnlock()
				if cacheErr := s.cacheRepo.Set(ctx, normalizedLink, result, platform, invalidTTL, platformTTLMap); cacheErr != nil {
					log.Printf("Failed to cache result for link %s: %v", normalizedLink, cacheErr)
				}
			}
			if err != nil {
				// 检查是否为429错误（频率限制错误）
				if apphttp.IsRateLimitError(err) {
					// 429错误也要保存到invalid_links表，标记IsRateLimited为true
					var duration int64
					if result != nil {
						duration = result.Duration
					}
					mu.Lock()
					*invalidLinks = append(*invalidLinks, model.InvalidLink{
						Link:          l,
						Platform:      ch.GetPlatform(),
						FailureReason: fmt.Sprintf("API频率限制（429错误）: %v", err),
						CheckDuration: &duration,
						IsRateLimited: true,
						SubmissionID:  &submissionID,
						CreatedAt:     time.Now(),
					})
					mu.Unlock()
					log.Printf("检测链接时遇到频率限制（429错误），已保存到数据库: %s, 平台: %s, 错误: %v", l, ch.GetPlatform(), err)
					return
				}

				// 其他检测错误，视为无效
				var duration int64
				var isRateLimited bool
				if result != nil {
					duration = result.Duration
					isRateLimited = result.IsRateLimited
				}
				mu.Lock()
				*invalidLinks = append(*invalidLinks, model.InvalidLink{
					Link:          l,
					Platform:      ch.GetPlatform(),
					FailureReason: fmt.Sprintf("检测错误: %v", err),
					CheckDuration: &duration,
					IsRateLimited: isRateLimited,
					SubmissionID:  &submissionID,
					CreatedAt:     time.Now(),
				})
				mu.Unlock()
				return
			}

			mu.Lock()
			if result.Valid {
				*validLinks = append(*validLinks, normalizedLink)
			} else {
				invalidLink := model.InvalidLink{
					Link:          normalizedLink,
					Platform:      ch.GetPlatform(),
					FailureReason: result.FailureReason,
					CheckDuration: &result.Duration,
					IsRateLimited: result.IsRateLimited,
					SubmissionID:  &submissionID,
					CreatedAt:     time.Now(),
				}
				*invalidLinks = append(*invalidLinks, invalidLink)
			}
			mu.Unlock()
		}(link)
	}

	log.Printf("checkLinksWithConcurrency: Waiting for all checks to complete for platform %s, submission %d", platform, submissionID)
	wg.Wait()
	duration := time.Since(startTime)
	log.Printf("checkLinksWithConcurrency: Completed all checks for platform %s, submission %d, total: %d, duration: %v",
		platform, submissionID, len(links), duration)
}

// CheckPendingSubmissions 检测待检测的提交记录
func (s *CheckerService) CheckPendingSubmissions(limit int) error {
	records, err := s.submissionRepo.GetPendingRecords(limit)
	if err != nil {
		log.Printf("CheckPendingSubmissions: Failed to get pending records: %v", err)
		return err
	}

	log.Printf("CheckPendingSubmissions: Found %d pending records", len(records))

	for _, record := range records {
		if len(record.PendingLinks) > 0 {
			log.Printf("CheckPendingSubmissions: Processing submission %d with %d pending links", record.ID, len(record.PendingLinks))
			// 自动扫描时，检测所有pending_links，不进行平台过滤
			// 因为pending_links中可能包含未选中平台的链接，需要全部检测
			go func(r model.SubmissionRecord) {
				defer func() {
					if rec := recover(); rec != nil {
						log.Printf("CheckPendingSubmissions: Panic recovered in CheckRealtime for submission %d: %v", r.ID, rec)
					}
				}()
				log.Printf("CheckPendingSubmissions: Starting CheckRealtime for submission %d", r.ID)
				_, err := s.CheckRealtime(r.ID, []string(r.PendingLinks))
				if err != nil {
					log.Printf("CheckPendingSubmissions: Failed to check submission %d: %v", r.ID, err)
				} else {
					log.Printf("CheckPendingSubmissions: Successfully checked submission %d", r.ID)
				}
			}(record)
		} else {
			// 没有待检测链接，将任务状态更新为 checked，表示已完成检测
			log.Printf("CheckPendingSubmissions: No pending links for submission %d, updating status to checked", record.ID)

			// 重新获取记录以确保获取最新状态
			updatedRecord, err := s.submissionRepo.GetByID(record.ID)
			if err != nil {
				log.Printf("CheckPendingSubmissions: Failed to get submission record %d: %v", record.ID, err)
				continue
			}

			// 如果已经是 checked 状态，跳过
			if updatedRecord.Status == "checked" {
				log.Printf("CheckPendingSubmissions: Submission %d is already checked, skipping", record.ID)
				continue
			}

			// 更新状态为 checked
			now := time.Now()
			updatedRecord.Status = "checked"
			updatedRecord.CheckedAt = &now
			updatedRecord.PendingLinks = model.StringArray([]string{})

			if err := s.submissionRepo.Update(updatedRecord); err != nil {
				log.Printf("CheckPendingSubmissions: Failed to update submission %d status to checked: %v", record.ID, err)
			} else {
				log.Printf("CheckPendingSubmissions: Successfully updated submission %d status to checked", record.ID)
			}
		}
	}

	return nil
}

// GetConcurrencyLimit 获取平台的并发限制
func (s *CheckerService) GetConcurrencyLimit(platform model.Platform) int {
	key := fmt.Sprintf("platform_concurrency_%s", platform.String())
	_, err := s.settingsRepo.GetByKey(key)
	if err != nil {
		// 如果不存在，返回默认值
		ch, ok := s.checkerFactory.GetChecker(platform)
		if ok {
			return ch.GetConcurrencyLimit()
		}
		return 5
	}

	// 解析配置值（这里简化处理，实际应该解析JSON）
	// 暂时返回检测器的默认值
	ch, ok := s.checkerFactory.GetChecker(platform)
	if ok {
		return ch.GetConcurrencyLimit()
	}
	return 5
}

// getPlatformKeys 获取平台映射的键列表（用于日志）
func getPlatformKeys(linksByPlatform map[model.Platform][]string) []string {
	keys := make([]string, 0, len(linksByPlatform))
	for platform := range linksByPlatform {
		keys = append(keys, platform.String())
	}
	return keys
}
