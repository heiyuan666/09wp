package scheduler

import (
	"PanCheck/internal/model"
	"PanCheck/internal/repository"
	"PanCheck/internal/service"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler 任务调度器
type Scheduler struct {
	cron            *cron.Cron
	taskService     *service.ScheduledTaskService
	schedulerService *service.TaskSchedulerService
	taskRepo        *repository.ScheduledTaskRepository
	ctx             context.Context
	cancel          context.CancelFunc
	entries         map[uint]cron.EntryID // 任务ID到cron entry ID的映射
}

// NewScheduler 创建调度器
func NewScheduler(taskService *service.ScheduledTaskService, schedulerService *service.TaskSchedulerService) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		cron:            cron.New(cron.WithSeconds()), // 支持秒级精度
		taskService:     taskService,
		schedulerService: schedulerService,
		taskRepo:        repository.NewScheduledTaskRepository(),
		ctx:             ctx,
		cancel:          cancel,
		entries:         make(map[uint]cron.EntryID),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	log.Println("Starting task scheduler...")

	// 加载所有活跃任务
	if err := s.loadActiveTasks(); err != nil {
		return fmt.Errorf("failed to load active tasks: %v", err)
	}

	// 启动cron调度器
	s.cron.Start()

	// 启动过期检查协程
	go s.checkExpiredTasks()

	log.Println("Task scheduler started")
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	log.Println("Stopping task scheduler...")
	s.cancel()
	s.cron.Stop()
	log.Println("Task scheduler stopped")
}

// loadActiveTasks 加载所有活跃任务
func (s *Scheduler) loadActiveTasks() error {
	tasks, err := s.taskRepo.GetAllActive()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if err := s.addTask(&task); err != nil {
			log.Printf("Failed to add task %d: %v", task.ID, err)
		}
	}

	log.Printf("Loaded %d active tasks", len(tasks))
	return nil
}

// addTask 添加任务到调度器
func (s *Scheduler) addTask(task *model.ScheduledTask) error {
	// 计算下次执行时间
	nextRun, err := s.calculateNextRun(task.CronExpression)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %v", err)
	}

	// 如果任务已存在，先移除
	if entryID, exists := s.entries[task.ID]; exists {
		s.cron.Remove(entryID)
	}

	// 创建任务执行函数
	taskCopy := *task // 复制任务以避免闭包问题
	entryID, err := s.cron.AddFunc(task.CronExpression, func() {
		// 重新获取任务以确保使用最新数据
		currentTask, err := s.taskRepo.GetByID(taskCopy.ID)
		if err != nil {
			log.Printf("Failed to get task %d: %v", taskCopy.ID, err)
			return
		}
		s.executeTask(currentTask)
	})

	if err != nil {
		return err
	}

	s.entries[task.ID] = entryID

	// 更新任务的下次执行时间
	task.NextRunAt = nextRun
	if err := s.taskRepo.Update(task); err != nil {
		log.Printf("Failed to update task %d next run time: %v", task.ID, err)
	}

	log.Printf("Added task %d with cron expression: %s, next run: %v", task.ID, task.CronExpression, nextRun)
	return nil
}

// removeTask 从调度器移除任务
func (s *Scheduler) removeTask(taskID uint) {
	if entryID, exists := s.entries[taskID]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, taskID)
		log.Printf("Removed task %d from scheduler", taskID)
	}
}

// executeTask 执行任务
func (s *Scheduler) executeTask(task *model.ScheduledTask) {
	log.Printf("Executing task %d: %s", task.ID, task.Name)

	// 检查任务是否仍然启用
	currentTask, err := s.taskRepo.GetByID(task.ID)
	if err != nil {
		log.Printf("Failed to get task %d: %v", task.ID, err)
		return
	}

	if currentTask.Status != "active" {
		log.Printf("Task %d is not active, removing from scheduler", task.ID)
		s.removeTask(task.ID)
		return
	}

	// 检查是否过期
	if currentTask.AutoDestroyAt != nil && time.Now().After(*currentTask.AutoDestroyAt) {
		log.Printf("Task %d has expired, stopping", task.ID)
		currentTask.Status = "expired"
		s.taskRepo.Update(currentTask)
		s.removeTask(task.ID)
		return
	}

	// 执行任务
	if err := s.schedulerService.ExecuteTask(currentTask); err != nil {
		log.Printf("Task %d execution failed: %v", task.ID, err)
	}

	// 更新最后执行时间和下次执行时间
	now := time.Now()
	currentTask.LastRunAt = &now
	nextRun, err := s.calculateNextRun(currentTask.CronExpression)
	if err == nil {
		currentTask.NextRunAt = nextRun
	}
	s.taskRepo.Update(currentTask)
}

// calculateNextRun 计算下次执行时间
func (s *Scheduler) calculateNextRun(cronExpr string) (*time.Time, error) {
	// 使用标准cron解析器（支持5位和6位表达式）
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return nil, err
	}

	next := schedule.Next(time.Now())
	return &next, nil
}

// checkExpiredTasks 定期检查过期任务
func (s *Scheduler) checkExpiredTasks() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.taskService.CheckExpiredTasks(); err != nil {
				log.Printf("Failed to check expired tasks: %v", err)
			}

			// 移除已过期的任务
			expiredTasks, err := s.taskRepo.GetExpiredTasks()
			if err != nil {
				log.Printf("Failed to get expired tasks: %v", err)
				continue
			}

			for _, task := range expiredTasks {
				s.removeTask(task.ID)
			}
		}
	}
}

// ReloadTask 重新加载任务（用于任务更新后）
func (s *Scheduler) ReloadTask(taskID uint) error {
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return err
	}

	// 如果任务已存在，先移除
	s.removeTask(taskID)

	// 如果任务处于活跃状态，重新添加
	if task.Status == "active" {
		return s.addTask(task)
	}

	return nil
}

