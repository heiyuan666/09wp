import { api } from './authApi';

export interface PlatformRateConfig {
  concurrency: number;
  request_delay_ms: number;
  max_requests_per_second: number;
  cache_ttl_hours?: number; // 有效链接缓存过期时间（小时）
}

export interface RedisConfig {
  enabled: boolean;
  host: string;
  port: number;
  username?: string; // Redis用户名（Redis 6.0+ ACL支持，留空则只使用密码）
  password: string;
  invalid_ttl: number; // 无效链接缓存过期时间（小时）
}

export const settingsApi = {
  // 获取所有平台的频率配置
  getRateConfigSettings: async () => {
    const response = await api.get<{ data: Record<string, PlatformRateConfig> }>('/settings/rate-config');
    return response.data.data;
  },

  // 批量更新平台频率配置
  updateRateConfigSettings: async (settings: Record<string, PlatformRateConfig>) => {
    const response = await api.put<{
      message: string;
      data: Record<string, PlatformRateConfig>;
    }>('/settings/rate-config', { settings });
    return response.data;
  },

  // 获取Redis配置
  getRedisConfig: async () => {
    const response = await api.get<{ data: RedisConfig }>('/settings/redis-config');
    return response.data.data;
  },

  // 更新Redis配置
  updateRedisConfig: async (config: RedisConfig) => {
    const response = await api.put<{
      message: string;
      data: RedisConfig;
    }>('/settings/redis-config', { config });
    return response.data;
  },
};

