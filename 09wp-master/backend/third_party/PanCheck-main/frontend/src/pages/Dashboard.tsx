import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { statisticsApi, type StatisticsOverview, type PlatformInvalidCount, type TimeSeriesData } from '@/api/statisticsApi';
import { PLATFORM_NAMES } from '@/utils/constants';
import { TimeRangeSelector, type TimeRange } from '@/components/TimeRangeSelector';
import { toast } from 'sonner';
import { BarChart, Bar, Rectangle, Cell, LineChart, Line, XAxis, YAxis, CartesianGrid } from 'recharts';
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from '@/components/ui/chart';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { linkApi } from '@/api/linkApi';

// 柱状图配置
const createChartConfig = (platforms: string[]): ChartConfig => {
  const colors = ['hsl(var(--chart-1))', 'hsl(var(--chart-2))', 'hsl(var(--chart-3))', 'hsl(var(--chart-4))', 'hsl(var(--chart-5))', 'hsl(var(--chart-6))', 'hsl(var(--chart-7))', 'hsl(var(--chart-8))', 'hsl(var(--chart-9))'];
  
  const config: ChartConfig = {
    count: {
      label: '失效链接数',
    },
  };

  platforms.forEach((platform, index) => {
    config[platform] = {
      label: PLATFORM_NAMES[platform] || platform,
      color: colors[index % colors.length],
    };
  });

  return config;
};

interface RateLimitedLink {
  id: number;
  link: string;
  platform: string;
  failure_reason: string;
  check_duration?: number;
  is_rate_limited: boolean;
  submission_id?: number;
  created_at: string;
}

export function Dashboard() {
  const [overview, setOverview] = useState<StatisticsOverview | null>(null);
  const [platformCounts, setPlatformCounts] = useState<PlatformInvalidCount[]>([]);
  const [timeSeriesData, setTimeSeriesData] = useState<TimeSeriesData[]>([]);
  const [loading, setLoading] = useState(true);
  const [timeSeriesLoading, setTimeSeriesLoading] = useState(false);
  const [timeRange, setTimeRange] = useState<TimeRange>('last7d');
  
  // 受限链接弹窗相关状态
  const [rateLimitedDialogOpen, setRateLimitedDialogOpen] = useState(false);
  const [rateLimitedLinks, setRateLimitedLinks] = useState<RateLimitedLink[]>([]);
  const [rateLimitedLoading, setRateLimitedLoading] = useState(false);
  const [selectedPlatform, setSelectedPlatform] = useState<string>('all');
  const [rateLimitedPage, setRateLimitedPage] = useState(1);
  const [rateLimitedTotal, setRateLimitedTotal] = useState(0);
  const pageSize = 20;

  // 计算时间范围
  const getDateRange = (range: TimeRange): { start?: string; end?: string; granularity: 'hour' | 'day' } => {
    const today = new Date();
    const end = today.toISOString().split('T')[0];
    let start: string;
    let granularity: 'hour' | 'day' = 'day';

    switch (range) {
      case 'today':
        start = end;
        granularity = 'hour';
        break;
      case 'last24h':
        const yesterday = new Date(today);
        yesterday.setDate(yesterday.getDate() - 1);
        start = yesterday.toISOString().split('T')[0];
        granularity = 'hour';
        break;
      case 'thisWeek':
        const weekStart = new Date(today);
        // 获取本周一：如果今天是周日(getDay()=0)，则往前推6天；否则往前推(getDay()-1)天
        const dayOfWeek = today.getDay();
        const daysToMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
        weekStart.setDate(today.getDate() - daysToMonday);
        start = weekStart.toISOString().split('T')[0];
        granularity = 'day';
        break;
      case 'last7d':
        const last7d = new Date(today);
        last7d.setDate(today.getDate() - 7);
        start = last7d.toISOString().split('T')[0];
        granularity = 'day';
        break;
      case 'thisMonth':
        const monthStart = new Date(today.getFullYear(), today.getMonth(), 1);
        start = monthStart.toISOString().split('T')[0];
        granularity = 'day';
        break;
      case 'last30d':
        const last30d = new Date(today);
        last30d.setDate(today.getDate() - 30);
        start = last30d.toISOString().split('T')[0];
        granularity = 'day';
        break;
      case 'last90d':
        const last90d = new Date(today);
        last90d.setDate(today.getDate() - 90);
        start = last90d.toISOString().split('T')[0];
        granularity = 'day';
        break;
      default:
        start = end;
    }

    return { start, end, granularity };
  };

  // 加载概览和饼图数据（只在组件挂载时调用）
  const loadOverviewData = async () => {
    setLoading(true);
    try {
      const [overviewData, platformData] = await Promise.all([
        statisticsApi.getOverview(),
        statisticsApi.getPlatformInvalidCounts(),
      ]);

      setOverview(overviewData);
      setPlatformCounts(platformData.filter(p => p.count > 0)); // 只显示有数据的平台
    } catch (error: any) {
      console.error('加载统计数据失败:', error);
      toast.error('加载统计数据失败: ' + (error.response?.data?.error || error.message));
    } finally {
      setLoading(false);
    }
  };

  // 只加载时间序列数据（在时间范围改变时调用）
  const loadTimeSeriesData = async () => {
    setTimeSeriesLoading(true);
    try {
      const dateRange = getDateRange(timeRange);
      const timeSeries = await statisticsApi.getSubmissionTimeSeries(
        dateRange.start,
        dateRange.end,
        dateRange.granularity
      );
      setTimeSeriesData(timeSeries);
    } catch (error: any) {
      console.error('加载时间序列数据失败:', error);
      toast.error('加载时间序列数据失败: ' + (error.response?.data?.error || error.message));
    } finally {
      setTimeSeriesLoading(false);
    }
  };

  // 组件挂载时加载概览和饼图数据
  useEffect(() => {
    loadOverviewData();
  }, []);

  // 时间范围改变时只加载时间序列数据
  useEffect(() => {
    loadTimeSeriesData();
  }, [timeRange]);

  // 加载受限链接列表
  const loadRateLimitedLinks = async () => {
    setRateLimitedLoading(true);
    try {
      const result = await linkApi.listRateLimitedLinks(
        rateLimitedPage,
        pageSize,
        selectedPlatform === 'all' ? undefined : selectedPlatform
      );
      setRateLimitedLinks(result.data);
      setRateLimitedTotal(result.total);
    } catch (error: any) {
      console.error('加载受限链接失败:', error);
      toast.error('加载受限链接失败: ' + (error.response?.data?.error || error.message));
    } finally {
      setRateLimitedLoading(false);
    }
  };

  // 打开弹窗时加载数据
  useEffect(() => {
    if (rateLimitedDialogOpen) {
      loadRateLimitedLinks();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [rateLimitedDialogOpen, rateLimitedPage, selectedPlatform]);

  // 清空受限链接
  const handleClearRateLimitedLinks = async () => {
    if (!confirm('确定要清空所有受限链接吗？清空后这些链接可以重新检测。')) {
      return;
    }

    try {
      await linkApi.clearRateLimitedLinks();
      toast.success('已清空所有受限链接');
      // 重新加载数据
      await loadRateLimitedLinks();
      // 重新加载概览数据以更新统计
      await loadOverviewData();
    } catch (error: any) {
      console.error('清空受限链接失败:', error);
      toast.error('清空受限链接失败: ' + (error.response?.data?.error || error.message));
    }
  };

  // 平台筛选改变时重置页码
  useEffect(() => {
    if (rateLimitedDialogOpen) {
      setRateLimitedPage(1);
    }
  }, [selectedPlatform, rateLimitedDialogOpen]);

  const barChartData = platformCounts.map(item => ({
    platform: item.platform,
    count: item.count,
    fill: `var(--color-${item.platform})`,
  }));

  const chartConfig = createChartConfig(platformCounts.map(p => p.platform));

  // 折线图配置
  const lineChartConfig: ChartConfig = {
    count: {
      label: '提交记录数',
      color: 'hsl(var(--chart-1))',
    },
  };

  // 格式化时间显示
  const formatTimeLabel = (dateStr: string, timeRange: TimeRange): string => {
    if (timeRange === 'today' || timeRange === 'last24h') {
      // 按小时显示，格式：HH:00
      try {
        // 处理不同的时间格式：2025-11-22T00:00:00+08:00 或 2025-11-22 00:00:00
        let date: Date;
        if (dateStr.includes('T')) {
          date = new Date(dateStr);
        } else if (dateStr.includes(' ')) {
          // 格式：2025-11-22 00:00:00
          date = new Date(dateStr.replace(' ', 'T') + '+08:00');
        } else {
          // 格式：2025-11-22
          date = new Date(dateStr + 'T00:00:00+08:00');
        }
        const hours = date.getHours().toString().padStart(2, '0');
        return `${hours}:00`;
      } catch {
        return dateStr;
      }
    } else {
      // 按天显示，格式：MM-DD
      try {
        // 处理不同的时间格式
        let date: Date;
        if (dateStr.includes('T')) {
          date = new Date(dateStr);
        } else if (dateStr.includes(' ')) {
          date = new Date(dateStr.split(' ')[0]);
        } else {
          date = new Date(dateStr);
        }
        const month = (date.getMonth() + 1).toString().padStart(2, '0');
        const day = date.getDate().toString().padStart(2, '0');
        return `${month}-${day}`;
      } catch {
        return dateStr;
      }
    }
  };

  const lineChartData = timeSeriesData.map(item => ({
    date: item.date,
    displayDate: formatTimeLabel(item.date, timeRange),
    count: item.count,
  }));

  return (
    <div className="space-y-8">

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总失效链接数</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="text-2xl font-bold">加载中...</div>
            ) : (
              <>
                <div className="text-2xl font-bold">{overview?.total_invalid_links.toLocaleString() || 0}</div>
                <p className="text-xs text-muted-foreground">累计失效链接总数</p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总提交记录数</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="text-2xl font-bold">加载中...</div>
            ) : (
              <>
                <div className="text-2xl font-bold">{overview?.total_submissions.toLocaleString() || 0}</div>
                <p className="text-xs text-muted-foreground">所有提交记录总数</p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">已完成检测</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="text-2xl font-bold">加载中...</div>
            ) : (
              <>
                <div className="text-2xl font-bold">{overview?.completed_submissions.toLocaleString() || 0}</div>
                <p className="text-xs text-muted-foreground">已完成检测的记录数</p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待检测记录</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="text-2xl font-bold">加载中...</div>
            ) : (
              <>
                <div className="text-2xl font-bold">{overview?.pending_submissions.toLocaleString() || 0}</div>
                <p className="text-xs text-muted-foreground">等待检测的记录数</p>
              </>
            )}
          </CardContent>
        </Card>

        <Card 
          className="cursor-pointer hover:bg-accent/50 transition-colors"
          onClick={() => setRateLimitedDialogOpen(true)}
        >
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">受限检测链接数</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="text-2xl font-bold">加载中...</div>
            ) : (
              <>
                <div className="text-2xl font-bold">{overview?.rate_limited_links.toLocaleString() || 0}</div>
                <p className="text-xs text-muted-foreground">可能被限制导致检测无效的链接数</p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总定时任务数</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className="text-2xl font-bold">加载中...</div>
            ) : (
              <>
                <div className="text-2xl font-bold">{overview?.total_scheduled_tasks.toLocaleString() || 0}</div>
                <p className="text-xs text-muted-foreground">所有定时任务总数</p>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 统计图表：柱状图和折线图 */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* 柱状图：各大网盘失效记录数 */}
        <Card>
          <CardHeader>
            <CardTitle>各大网盘失效记录数</CardTitle>
            <CardDescription>按平台统计的失效链接分布</CardDescription>
          </CardHeader>
          <CardContent className="flex items-center justify-center">
            {loading ? (
              <div className="flex items-center justify-center h-[250px] w-full">加载中...</div>
            ) : barChartData.length === 0 ? (
              <div className="flex items-center justify-center h-[250px] w-full text-muted-foreground">
                暂无数据
              </div>
            ) : (
              <ChartContainer config={chartConfig} className="h-[250px] w-full flex items-center justify-center">
                <BarChart data={barChartData}>
                  <CartesianGrid vertical={false} />
                  <XAxis
                    dataKey="platform"
                    tickLine={false}
                    tickMargin={10}
                    axisLine={false}
                    tickFormatter={(value) =>
                      chartConfig[value as keyof typeof chartConfig]?.label || value
                    }
                  />
                  <YAxis tick={{ fontSize: 12 }} />
                  <ChartTooltip
                    cursor={false}
                    content={<ChartTooltipContent hideLabel />}
                  />
                  <Bar
                    dataKey="count"
                    strokeWidth={2}
                    radius={8}
                    activeBar={({ ...props }) => {
                      return (
                        <Rectangle
                          {...props}
                          fillOpacity={0.8}
                          strokeDasharray={4}
                          strokeDashoffset={4}
                        />
                      );
                    }}
                  >
                    {barChartData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.fill} />
                    ))}
                  </Bar>
                </BarChart>
              </ChartContainer>
            )}
          </CardContent>
        </Card>

        {/* 折线图：各个时间段提交记录数 */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle>提交记录趋势</CardTitle>
              <CardDescription>按时间统计的提交记录数量</CardDescription>
            </div>
            <TimeRangeSelector
              value={timeRange}
              onChange={setTimeRange}
            />
          </CardHeader>
          <CardContent className="flex items-center justify-center">
            {timeSeriesLoading ? (
              <div className="flex items-center justify-center h-[250px] w-full">加载中...</div>
            ) : lineChartData.length === 0 ? (
              <div className="flex items-center justify-center h-[250px] w-full text-muted-foreground">
                暂无数据
              </div>
            ) : (
              <ChartContainer config={lineChartConfig} className="h-[250px] w-full flex items-center justify-center">
                <LineChart
                  accessibilityLayer
                  data={lineChartData}
                  margin={{
                    left: 12,
                    right: 12,
                  }}
                >
                  <CartesianGrid vertical={false} />
                  <XAxis
                    dataKey="displayDate"
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                    tick={{ fontSize: 12 }}
                    angle={timeRange === 'today' || timeRange === 'last24h' ? 0 : -45}
                    textAnchor={timeRange === 'today' || timeRange === 'last24h' ? 'middle' : 'end'}
                    height={timeRange === 'today' || timeRange === 'last24h' ? 40 : 60}
                  />
                  <ChartTooltip
                    cursor={false}
                    content={<ChartTooltipContent hideLabel />}
                  />
                  <Line
                    dataKey="count"
                    type="natural"
                    stroke="var(--color-count)"
                    strokeWidth={2}
                    dot={false}
                  />
                </LineChart>
              </ChartContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 受限链接明细弹窗 */}
      <Dialog open={rateLimitedDialogOpen} onOpenChange={setRateLimitedDialogOpen}>
        <DialogContent className="max-w-6xl max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>受限检测链接明细</DialogTitle>
            <DialogDescription>
              显示所有可能被限制导致检测无效的链接，清空后可以重新检测这些链接
            </DialogDescription>
          </DialogHeader>
          
          <div className="flex items-center justify-between gap-4 mb-4">
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">平台筛选：</span>
              <Select value={selectedPlatform} onValueChange={setSelectedPlatform}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="选择平台" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部平台</SelectItem>
                  {Object.entries(PLATFORM_NAMES).map(([key, name]) => (
                    <SelectItem key={key} value={key}>
                      {name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button
              variant="destructive"
              onClick={handleClearRateLimitedLinks}
              disabled={rateLimitedLoading || rateLimitedTotal === 0}
            >
              清空所有
            </Button>
          </div>

          <div className="flex-1 overflow-auto">
            {rateLimitedLoading ? (
              <div className="flex items-center justify-center h-[300px]">加载中...</div>
            ) : rateLimitedLinks.length === 0 ? (
              <div className="flex items-center justify-center h-[300px] text-muted-foreground">
                暂无数据
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[120px]">平台</TableHead>
                    <TableHead className="min-w-[400px]">链接</TableHead>
                    <TableHead className="min-w-[200px]">失败原因</TableHead>
                    <TableHead className="w-[120px]">检测耗时</TableHead>
                    <TableHead className="w-[180px]">创建时间</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rateLimitedLinks.map((link) => (
                    <TableRow key={link.id}>
                      <TableCell>{PLATFORM_NAMES[link.platform] || link.platform}</TableCell>
                      <TableCell className="max-w-[500px] break-all">
                        <a
                          href={link.link}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-primary hover:underline"
                          title={link.link}
                        >
                          {link.link}
                        </a>
                      </TableCell>
                      <TableCell className="max-w-[300px] break-words" title={link.failure_reason}>
                        {link.failure_reason || '-'}
                      </TableCell>
                      <TableCell>
                        {link.check_duration ? `${link.check_duration}ms` : '-'}
                      </TableCell>
                      <TableCell>
                        {new Date(link.created_at).toLocaleString('zh-CN')}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>

          <DialogFooter className="flex items-center justify-between">
            <div className="text-sm text-muted-foreground">
              共 {rateLimitedTotal} 条记录
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setRateLimitedPage(p => Math.max(1, p - 1))}
                disabled={rateLimitedPage === 1 || rateLimitedLoading}
              >
                上一页
              </Button>
              <span className="text-sm">
                第 {rateLimitedPage} / {Math.ceil(rateLimitedTotal / pageSize)} 页
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setRateLimitedPage(p => p + 1)}
                disabled={rateLimitedPage >= Math.ceil(rateLimitedTotal / pageSize) || rateLimitedLoading}
              >
                下一页
              </Button>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
