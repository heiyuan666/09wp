import { api } from './authApi';

export interface StatisticsOverview {
  total_invalid_links: number;
  total_submissions: number;
  completed_submissions: number;
  pending_submissions: number;
  rate_limited_links: number;
  total_scheduled_tasks: number;
}

export interface PlatformInvalidCount {
  platform: string;
  count: number;
}

export interface TimeSeriesData {
  date: string;
  count: number;
}

export const statisticsApi = {
  // 获取统计概览
  getOverview: async (): Promise<StatisticsOverview> => {
    const response = await api.get<{ data: StatisticsOverview }>('/statistics/overview');
    return response.data.data;
  },

  // 获取各大网盘失效记录数
  getPlatformInvalidCounts: async (): Promise<PlatformInvalidCount[]> => {
    const response = await api.get<{ data: PlatformInvalidCount[] }>('/statistics/platform-invalid-counts');
    return response.data.data;
  },

  // 获取各个时间段提交记录数
  getSubmissionTimeSeries: async (startTime?: string, endTime?: string, granularity?: 'hour' | 'day'): Promise<TimeSeriesData[]> => {
    const params: Record<string, string> = {};
    if (startTime) params.start_time = startTime;
    if (endTime) params.end_time = endTime;
    if (granularity) params.granularity = granularity;
    
    const response = await api.get<{ data: TimeSeriesData[] }>('/statistics/submission-time-series', { params });
    return response.data.data;
  },
};

