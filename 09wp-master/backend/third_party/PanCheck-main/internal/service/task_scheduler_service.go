package service

import (
	"PanCheck/internal/model"
	"PanCheck/internal/repository"
	"PanCheck/pkg/utils"
	"context"
	"fmt"
	"log"
	"time"
)

// TaskSchedulerService 任务调度服务
type TaskSchedulerService struct {
	taskService    *ScheduledTaskService
	linkService    *LinkService
	checkerService *CheckerService
	executionRepo  *repository.TaskExecutionRepository
}

// NewTaskSchedulerService 创建任务调度服务
func NewTaskSchedulerService(taskService *ScheduledTaskService, linkService *LinkService, checkerService *CheckerService) *TaskSchedulerService {
	return &TaskSchedulerService{
		taskService:    taskService,
		linkService:    linkService,
		checkerService: checkerService,
		executionRepo:  repository.NewTaskExecutionRepository(),
	}
}

// ExecuteTask 执行任务
func (s *TaskSchedulerService) ExecuteTask(task *model.ScheduledTask) error {
	log.Printf("ExecuteTask: Starting task %d: %s", task.ID, task.Name)
	
	// 创建执行记录
	execution := &model.TaskExecution{
		TaskID:    task.ID,
		Status:    "running",
		StartedAt: time.Now(),
	}
	if err := s.executionRepo.Create(execution); err != nil {
		log.Printf("ExecuteTask: Failed to create execution record for task %d: %v", task.ID, err)
		return fmt.Errorf("failed to create execution record: %v", err)
	}
	log.Printf("ExecuteTask: Created execution record for task %d, execution ID: %d", task.ID, execution.ID)

	startTime := time.Now()
	var links []string
	var err error

	defer func() {
		// 捕获 panic，确保即使发生 panic 也能更新状态
		if r := recover(); r != nil {
			log.Printf("ExecuteTask: Panic recovered in task %d: %v", task.ID, r)
			err = fmt.Errorf("panic: %v", r)
			execution.Status = "failed"
			execution.ErrorMessage = fmt.Sprintf("panic: %v", r)
		}

		// 更新执行记录
		duration := time.Since(startTime).Milliseconds()
		execution.ExecutionDuration = &duration
		now := time.Now()
		execution.FinishedAt = &now

		if err != nil {
			execution.Status = "failed"
			execution.ErrorMessage = err.Error()
			log.Printf("ExecuteTask: Task %d failed: %v", task.ID, err)
		} else {
			execution.Status = "success"
			log.Printf("ExecuteTask: Task %d completed successfully", task.ID)
		}

		log.Printf("ExecuteTask: Updating execution record for task %d, status: %s", task.ID, execution.Status)
		if updateErr := s.executionRepo.Update(execution); updateErr != nil {
			log.Printf("ExecuteTask: Failed to update execution record for task %d: %v", task.ID, updateErr)
		} else {
			log.Printf("ExecuteTask: Successfully updated execution record for task %d", task.ID)
		}
	}()

	// 1. 执行curl命令获取数据
	log.Printf("ExecuteTask: Step 1 - Executing curl command for task %d", task.ID)
	rawData, err := s.taskService.ExecuteCurlCommand(task.CurlCommand)
	if err != nil {
		log.Printf("ExecuteTask: Step 1 failed for task %d: %v", task.ID, err)
		return fmt.Errorf("curl execution failed: %v", err)
	}
	log.Printf("ExecuteTask: Step 1 completed for task %d, rawData length: %d", task.ID, len(rawData))

	// 2. 执行转换脚本转换为JSON数组
	log.Printf("ExecuteTask: Step 2 - Transforming data for task %d", task.ID)
	links, err = s.taskService.TransformData(rawData, task.TransformScript)
	if err != nil {
		log.Printf("ExecuteTask: Step 2 failed for task %d: %v", task.ID, err)
		return fmt.Errorf("data transformation failed: %v", err)
	}
	log.Printf("ExecuteTask: Step 2 completed for task %d, links count: %d", task.ID, len(links))

	execution.LinksCount = len(links)
	if len(links) == 0 {
		log.Printf("ExecuteTask: Task %d: No links found, completing execution", task.ID)
		return nil
	}

	// 3. 调用LinkService创建提交记录（和用户提交链接的逻辑一样）
	log.Printf("ExecuteTask: Step 3 - Creating submission record for task %d with %d links", task.ID, len(links))
	checkReq := &CheckLinksRequest{
		Links: links,
		// 定时任务不选择平台，检测所有链接
		SelectedPlatforms: []model.Platform{},
	}

	// 使用系统标识作为客户端IP
	resp, checkErr := s.linkService.CheckLinks(checkReq, "system", utils.DeviceInfo{})
	if checkErr != nil {
		log.Printf("ExecuteTask: Step 3 failed for task %d: %v", task.ID, checkErr)
		err = fmt.Errorf("link check failed: %v", checkErr)
		return err
	}
	log.Printf("ExecuteTask: Step 3 completed for task %d, submission ID: %d", task.ID, resp.SubmissionID)

	// 4. 调用CheckerService执行实际检测（和用户提交链接的检测逻辑一样）
	// 这会执行实际检测并将失效链接保存到数据库
	log.Printf("ExecuteTask: Step 4 - Executing realtime check for task %d, submission ID: %d", task.ID, resp.SubmissionID)
	submissionRecord, checkErr := s.checkerService.CheckRealtime(resp.SubmissionID, resp.PendingLinks)
	if checkErr != nil {
		log.Printf("ExecuteTask: Step 4 failed for task %d: %v", task.ID, checkErr)
		err = fmt.Errorf("link check execution failed: %v", checkErr)
		return err
	}
	log.Printf("ExecuteTask: Step 4 completed for task %d", task.ID)

	// 5. 更新执行记录统计
	log.Printf("ExecuteTask: Step 5 - Updating execution statistics for task %d", task.ID)
	execution.CheckedCount = len(submissionRecord.OriginalLinks)
	execution.ValidCount = len(submissionRecord.ValidLinks)
	
	// 从提交记录中获取失效链接数量
	// 失效链接已经保存到 invalid_links 表中
	invalidLinksFromDB, _ := s.checkerService.GetInvalidLinksFromRecord(resp.SubmissionID)
	execution.InvalidCount = len(invalidLinksFromDB)

	log.Printf("ExecuteTask: Task %d executed successfully: %d links checked, %d valid, %d invalid",
		task.ID, execution.CheckedCount, execution.ValidCount, execution.InvalidCount)

	return nil
}

// ExecuteTaskAsync 异步执行任务
func (s *TaskSchedulerService) ExecuteTaskAsync(ctx context.Context, task *model.ScheduledTask) {
	go func() {
		if err := s.ExecuteTask(task); err != nil {
			log.Printf("Task %d execution failed: %v", task.ID, err)
		}
	}()
}

