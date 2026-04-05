import { api } from './authApi';

export interface ScheduledTask {
  id: number;
  name: string;
  description?: string;
  tags: string[];
  curl_command: string;
  transform_script?: string;
  cron_expression: string;
  auto_destroy_at?: string;
  status: 'active' | 'stopped' | 'expired';
  last_run_at?: string;
  next_run_at?: string;
  created_at: string;
  updated_at: string;
}

export interface TaskExecution {
  id: number;
  task_id: number;
  status: 'running' | 'success' | 'failed';
  links_count: number;
  checked_count: number;
  valid_count: number;
  invalid_count: number;
  error_message?: string;
  execution_duration?: number;
  started_at: string;
  finished_at?: string;
  created_at: string;
}

export interface ListTasksResponse {
  data: ScheduledTask[];
  total: number;
  page: number;
  page_size: number;
}

export interface ListExecutionsResponse {
  data: TaskExecution[];
  total: number;
  page: number;
  page_size: number;
}

export interface TagsResponse {
  tags: string[];
}

export interface TestTaskResponse {
  links: string[];
  count: number;
}

// 获取任务列表
export async function listTasks(params?: {
  page?: number;
  page_size?: number;
  tags?: string[];
  status?: string;
}): Promise<ListTasksResponse> {
  const queryParams: Record<string, string | number> = {};
  if (params?.page) queryParams.page = params.page;
  if (params?.page_size) queryParams.page_size = params.page_size;
  if (params?.status) queryParams.status = params.status;
  
  const config: { params?: Record<string, string | number | string[]> } = {};
  if (Object.keys(queryParams).length > 0 || params?.tags) {
    config.params = { ...queryParams };
    if (params?.tags) {
      config.params.tags = params.tags;
    }
  }

  const response = await api.get<ListTasksResponse>('/scheduled-tasks', config);
  return response.data;
}

// 获取任务详情
export async function getTask(id: number): Promise<ScheduledTask> {
  const response = await api.get<ScheduledTask>(`/scheduled-tasks/${id}`);
  return response.data;
}

// 创建任务
export async function createTask(task: Omit<ScheduledTask, 'id' | 'created_at' | 'updated_at' | 'enabled'>): Promise<ScheduledTask> {
  const response = await api.post<ScheduledTask>('/scheduled-tasks', task);
  return response.data;
}

// 更新任务
export async function updateTask(id: number, task: Partial<ScheduledTask>): Promise<ScheduledTask> {
  const response = await api.put<ScheduledTask>(`/scheduled-tasks/${id}`, task);
  return response.data;
}

// 删除任务
export async function deleteTask(id: number): Promise<void> {
  await api.delete(`/scheduled-tasks/${id}`);
}

// 测试任务配置
export async function testTaskConfig(id: number): Promise<TestTaskResponse> {
  const response = await api.post<TestTaskResponse>(`/scheduled-tasks/${id}/test`);
  return response.data;
}

// 手动触发执行任务
export async function runTask(id: number): Promise<void> {
  await api.post(`/scheduled-tasks/${id}/run`);
}

// 启用任务
export async function enableTask(id: number): Promise<void> {
  await api.post(`/scheduled-tasks/${id}/enable`);
}

// 禁用任务
export async function disableTask(id: number): Promise<void> {
  await api.post(`/scheduled-tasks/${id}/disable`);
}

// 获取任务执行历史
export async function getTaskExecutions(
  taskId: number,
  params?: { page?: number; page_size?: number }
): Promise<ListExecutionsResponse> {
  const config: { params?: Record<string, number> } = {};
  if (params?.page || params?.page_size) {
    config.params = {};
    if (params.page) config.params.page = params.page;
    if (params.page_size) config.params.page_size = params.page_size;
  }

  const response = await api.get<ListExecutionsResponse>(`/scheduled-tasks/${taskId}/executions`, config);
  return response.data;
}

// 获取所有标签
export async function getAllTags(): Promise<TagsResponse> {
  const response = await api.get<TagsResponse>('/scheduled-tasks/tags');
  return response.data;
}
