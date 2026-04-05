import { useState, useEffect } from 'react';
import { Plus, Edit, Trash2, Play, Pause, PlayCircle, History } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { TagInput } from '@/components/TagInput';
import { TagFilter } from '@/components/TagFilter';
import { CodeEditor } from '@/components/CodeEditor';
import { DateTimePicker24h } from '@/components/date-n-time/date-time-picker-24h';
import { toast } from 'sonner';
import {
  listTasks,
  createTask,
  updateTask,
  deleteTask,
  enableTask,
  disableTask,
  runTask,
  getTaskExecutions,
  type ScheduledTask,
  type TaskExecution,
} from '@/api/scheduledTaskApi';

export function ScheduledTasks() {
  const [tasks, setTasks] = useState<ScheduledTask[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<ScheduledTask | null>(null);
  const [executionsDialogOpen, setExecutionsDialogOpen] = useState(false);
  const [executions, setExecutions] = useState<TaskExecution[]>([]);
  const [executionsTaskName, setExecutionsTaskName] = useState<string>('');

  // 表单状态
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    tags: [] as string[],
    curl_command: '',
    transform_script: '',
    cron_expression: '',
    auto_destroy_at: null as Date | null,
    status: 'stopped' as 'active' | 'stopped' | 'expired',
  });

  useEffect(() => {
    loadTasks();
  }, [page, selectedTags, statusFilter]);

  const loadTasks = async () => {
    setLoading(true);
    try {
      const response = await listTasks({
        page,
        page_size: pageSize,
        tags: selectedTags.length > 0 ? selectedTags : undefined,
        status: statusFilter || undefined,
      });
      setTasks(response.data);
      setTotal(response.total);
    } catch (error) {
      console.error('Failed to load tasks:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingTask(null);
    setFormData({
      name: '',
      description: '',
      tags: [],
      curl_command: '',
      transform_script: '',
      cron_expression: '',
      auto_destroy_at: null,
      status: 'stopped',
    });
    setDialogOpen(true);
  };

  const handleEdit = (task: ScheduledTask) => {
    setEditingTask(task);
    setFormData({
      name: task.name,
      description: task.description || '',
      tags: task.tags || [],
      curl_command: task.curl_command,
      transform_script: task.transform_script || '',
      cron_expression: task.cron_expression,
      auto_destroy_at: task.auto_destroy_at ? new Date(task.auto_destroy_at) : null,
      status: task.status,
    });
    setDialogOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const taskData = {
        ...formData,
        auto_destroy_at: formData.auto_destroy_at 
          ? formData.auto_destroy_at.toISOString() 
          : undefined,
      };

      if (editingTask) {
        await updateTask(editingTask.id, taskData);
      } else {
        await createTask(taskData);
      }
      setDialogOpen(false);
      loadTasks();
      toast.success(editingTask ? '任务更新成功' : '任务创建成功');
    } catch (error: any) {
      toast.error(error.message || '操作失败');
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确定要删除这个任务吗？')) return;
    try {
      await deleteTask(id);
      loadTasks();
      toast.success('任务删除成功');
    } catch {
      toast.error('删除失败');
    }
  };

  const handleToggle = async (task: ScheduledTask) => {
    try {
      if (task.status === 'active') {
        await disableTask(task.id);
      } else {
        await enableTask(task.id);
      }
      loadTasks();
      toast.success(task.status === 'active' ? '任务已禁用' : '任务已启用');
    } catch {
      toast.error('操作失败');
    }
  };

  const handleRun = async (id: number) => {
    try {
      await runTask(id);
      toast.success('任务已开始执行');
      loadTasks();
    } catch {
      toast.error('执行失败');
    }
  };

  const handleTest = async () => {
    if (!formData.curl_command) {
      toast.error('请先输入curl命令');
      return;
    }
    try {
      // 统一使用当前弹框里输入框的内容进行测试配置
      const response = await fetch('/api/v1/scheduled-tasks/test/test', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          curl_command: formData.curl_command,
          transform_script: formData.transform_script,
        }),
      });
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || '测试失败');
      }
      const result = await response.json();
      toast.success(`测试成功！找到 ${result.count} 个链接`);
    } catch (error: any) {
      toast.error(error.message || '测试失败');
    }
  };

  const handleViewExecutions = async (task: ScheduledTask) => {
    setExecutionsTaskName(task.name);
    setExecutionsDialogOpen(true);
    try {
      const response = await getTaskExecutions(task.id, { page: 1, page_size: 20 });
      setExecutions(response.data);
    } catch (error) {
      console.error('Failed to load executions:', error);
      toast.error('加载执行历史失败');
    }
  };

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleString('zh-CN');
  };

  const formatDuration = (ms?: number) => {
    if (!ms) return '-';
    if (ms < 1000) {
      return `${ms}ms`;
    } else if (ms < 60000) {
      return `${(ms / 1000).toFixed(2)}s`;
    } else if (ms < 3600000) {
      const minutes = Math.floor(ms / 60000);
      const seconds = ((ms % 60000) / 1000).toFixed(0);
      return `${minutes}分${seconds}秒`;
    } else {
      const hours = Math.floor(ms / 3600000);
      const minutes = Math.floor((ms % 3600000) / 60000);
      return `${hours}小时${minutes}分钟`;
    }
  };

  const getStatusBadge = (status: string) => {
    const colors = {
      active: 'bg-green-100 text-green-800',
      stopped: 'bg-gray-100 text-gray-800',
      expired: 'bg-red-100 text-red-800',
    };
    return (
      <span className={`px-2 py-1 rounded text-xs ${colors[status as keyof typeof colors] || 'bg-gray-100'}`}>
        {status === 'active' ? '运行中' : status === 'stopped' ? '已停止' : '已过期'}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">定时任务管理</h1>
          <p className="text-muted-foreground">管理和配置定时检测任务</p>
        </div>
        <Button onClick={handleCreate}>
          <Plus className="mr-2 h-4 w-4" />
          创建任务
        </Button>
      </div>

      {/* 筛选区域 */}
      <Card>
        <CardHeader>
          <CardTitle>筛选</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <TagFilter selectedTags={selectedTags} onChange={setSelectedTags} />
          <div className="flex gap-2">
            <Button
              variant={statusFilter === '' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setStatusFilter('')}
            >
              全部
            </Button>
            <Button
              variant={statusFilter === 'active' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setStatusFilter('active')}
            >
              运行中
            </Button>
            <Button
              variant={statusFilter === 'stopped' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setStatusFilter('stopped')}
            >
              已停止
            </Button>
            <Button
              variant={statusFilter === 'expired' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setStatusFilter('expired')}
            >
              已过期
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 任务列表 */}
      <Card>
        <CardHeader>
          <CardTitle>任务列表</CardTitle>
          <CardDescription>共 {total} 个任务</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-8">加载中...</div>
          ) : tasks.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">暂无任务</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>名称</TableHead>
                  <TableHead>标签</TableHead>
                  <TableHead>Cron表达式</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>下次执行</TableHead>
                  <TableHead>最后执行</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tasks.map((task) => (
                  <TableRow key={task.id}>
                    <TableCell className="font-medium">{task.name}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {task.tags?.map((tag) => (
                          <span key={tag} className="px-2 py-0.5 bg-secondary text-secondary-foreground rounded text-xs">
                            {tag}
                          </span>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-sm">{task.cron_expression}</TableCell>
                    <TableCell>{getStatusBadge(task.status)}</TableCell>
                    <TableCell>{formatDate(task.next_run_at)}</TableCell>
                    <TableCell>{formatDate(task.last_run_at)}</TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleToggle(task)}
                          title={task.status === 'active' ? '禁用' : '启用'}
                        >
                          {task.status === 'active' ? <Pause className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRun(task.id)}
                          title="手动执行"
                        >
                          <PlayCircle className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleViewExecutions(task)}
                          title="执行历史"
                        >
                          <History className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEdit(task)}
                          title="编辑"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(task.id)}
                          title="删除"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
          {total > pageSize && (
            <div className="flex justify-between items-center mt-4">
              <div className="text-sm text-muted-foreground">
                第 {page} 页，共 {Math.ceil(total / pageSize)} 页
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === 1}
                  onClick={() => setPage(page - 1)}
                >
                  上一页
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= Math.ceil(total / pageSize)}
                  onClick={() => setPage(page + 1)}
                >
                  下一页
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 创建/编辑对话框 */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editingTask ? '编辑任务' : '创建任务'}</DialogTitle>
            <DialogDescription>
              {editingTask ? '修改任务配置信息' : '填写任务配置信息以创建新的定时任务'}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="name">任务名称</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">任务描述</Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <Label>标签</Label>
              <TagInput
                tags={formData.tags}
                onChange={(tags) => setFormData({ ...formData, tags })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="curl_command">Curl命令</Label>
              <Textarea
                id="curl_command"
                value={formData.curl_command}
                onChange={(e) => setFormData({ ...formData, curl_command: e.target.value })}
                placeholder="curl -X GET https://example.com/api/data"
                required
                className="font-mono text-sm"
                rows={4}
              />
            </div>

            <CodeEditor
              value={formData.transform_script}
              onChange={(value) => setFormData({ ...formData, transform_script: value })}
            />

            <div className="space-y-2">
              <Label htmlFor="cron_expression">Cron表达式</Label>
              <Input
                id="cron_expression"
                type="text"
                value={formData.cron_expression}
                onChange={(e) => setFormData({ ...formData, cron_expression: e.target.value })}
                placeholder="0 0/5 * * * ?"
                required
                className="font-mono"
              />
              <p className="text-xs text-muted-foreground">
                需要帮助生成 Cron 表达式？请访问{' '}
                <a
                  href="https://small-stone.github.io/vCrontab/dist/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  vCrontab 工具
                </a>
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="auto_destroy_at">自我销毁时间</Label>
              <DateTimePicker24h
                value={formData.auto_destroy_at}
                onChange={(date) => setFormData({ ...formData, auto_destroy_at: date || null })}
                placeholder="选择自我销毁时间"
              />
              <p className="text-xs text-muted-foreground">任务在此时间后自动停止（保留记录）</p>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="status_active"
                checked={formData.status === 'active'}
                onChange={(e) => {
                  const isActive = e.target.checked;
                  setFormData({ 
                    ...formData, 
                    status: isActive ? 'active' : 'stopped',
                  });
                }}
                className="h-4 w-4"
              />
              <Label htmlFor="status_active">立即激活</Label>
              <p className="text-xs text-muted-foreground ml-2">
                创建后立即开始执行定时任务
              </p>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={handleTest}>
                测试配置
              </Button>
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
                取消
              </Button>
              <Button type="submit">保存</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* 执行历史对话框 */}
      <Dialog open={executionsDialogOpen} onOpenChange={setExecutionsDialogOpen}>
        <DialogContent className="max-w-5xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>执行历史 - {executionsTaskName}</DialogTitle>
            <DialogDescription>最近 20 条执行记录</DialogDescription>
          </DialogHeader>
          <div className="mt-4">
            {executions.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">暂无执行记录</div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>执行时间</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead className="text-right">链接数</TableHead>
                    <TableHead className="text-right">有效数</TableHead>
                    <TableHead className="text-right">失效数</TableHead>
                    <TableHead className="text-right">耗时</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {executions.map((exec) => (
                    <TableRow key={exec.id}>
                      <TableCell className="font-medium">{formatDate(exec.started_at)}</TableCell>
                      <TableCell>
                        <span
                          className={`px-2 py-0.5 rounded text-xs ${
                            exec.status === 'success'
                              ? 'bg-green-100 text-green-800'
                              : exec.status === 'failed'
                              ? 'bg-red-100 text-red-800'
                              : 'bg-blue-100 text-blue-800'
                          }`}
                        >
                          {exec.status === 'success' ? '成功' : exec.status === 'failed' ? '失败' : '运行中'}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">{exec.links_count}</TableCell>
                      <TableCell className="text-right">{exec.valid_count}</TableCell>
                      <TableCell className="text-right">{exec.invalid_count}</TableCell>
                      <TableCell className="text-right font-medium">
                        {exec.execution_duration ? formatDuration(exec.execution_duration) : '-'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

