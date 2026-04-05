import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { settingsApi, PlatformRateConfig, RedisConfig } from '@/api/settingsApi';
import { toast } from 'sonner';
import { PLATFORM_NAMES } from '@/utils/constants';

export function Settings() {
  const [rateConfigSettings, setRateConfigSettings] = useState<Record<string, PlatformRateConfig>>({});
  const [redisConfig, setRedisConfig] = useState<RedisConfig>({
    enabled: false,
    host: 'localhost',
    port: 6379,
    username: '',
    password: '',
    invalid_ttl: 168,
  });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [savingRedis, setSavingRedis] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    setLoading(true);
    try {
      const [rateConfig, redisConfigData] = await Promise.all([
        settingsApi.getRateConfigSettings(),
        settingsApi.getRedisConfig(),
      ]);
      setRateConfigSettings(rateConfig);
      setRedisConfig(redisConfigData);
    } catch (error: any) {
      toast.error('加载设置失败: ' + (error.response?.data?.error || error.message));
    } finally {
      setLoading(false);
    }
  };

  const handleSaveRateConfig = async () => {
    // 验证所有配置
    for (const [platform, config] of Object.entries(rateConfigSettings)) {
      if (config.concurrency < 1 || config.concurrency > 100) {
        toast.error(`${PLATFORM_NAMES[platform] || platform} 的并发数必须在1-100之间`);
        return;
      }
      if (config.request_delay_ms < 0 || config.request_delay_ms > 10000) {
        toast.error(`${PLATFORM_NAMES[platform] || platform} 的请求间隔必须在0-10000毫秒之间`);
        return;
      }
      if (config.max_requests_per_second < 0 || config.max_requests_per_second > 100) {
        toast.error(`${PLATFORM_NAMES[platform] || platform} 的每秒最大请求数必须在0-100之间`);
        return;
      }
      if (config.cache_ttl_hours !== undefined && (config.cache_ttl_hours < 0 || config.cache_ttl_hours > 720)) {
        toast.error(`${PLATFORM_NAMES[platform] || platform} 的缓存有效期必须在0-720小时之间`);
        return;
      }
    }

    setSaving(true);
    try {
      await settingsApi.updateRateConfigSettings(rateConfigSettings);
      toast.success('保存成功！配置已立即生效');
    } catch (error: any) {
      toast.error('保存失败: ' + (error.response?.data?.error || error.message));
    } finally {
      setSaving(false);
    }
  };

  const handleSaveRedisConfig = async () => {
    // 验证配置
    if (redisConfig.port < 1 || redisConfig.port > 65535) {
      toast.error('Redis端口必须在1-65535之间');
      return;
    }
    if (redisConfig.invalid_ttl < 1 || redisConfig.invalid_ttl > 720) {
      toast.error('无效链接缓存过期时间必须在1-720小时之间');
      return;
    }

    setSavingRedis(true);
    try {
      await settingsApi.updateRedisConfig(redisConfig);
      toast.success('保存成功！请重启服务使配置生效');
    } catch (error: any) {
      toast.error('保存失败: ' + (error.response?.data?.error || error.message));
    } finally {
      setSavingRedis(false);
    }
  };

  const updateRateConfig = (platform: string, field: keyof PlatformRateConfig, value: any) => {
    setRateConfigSettings(prev => ({
      ...prev,
      [platform]: {
        ...prev[platform],
        [field]: value,
      },
    }));
  };

  const platforms = ['quark', 'uc', 'baidu', 'tianyi', 'pan123', 'pan115', 'aliyun', 'xunlei', 'cmcc'];

  const defaultRateConfig: PlatformRateConfig = {
    concurrency: 5,
    request_delay_ms: 0,
    max_requests_per_second: 0,
    cache_ttl_hours: 24,
  };

  return (
    <div className="container mx-auto py-8 space-y-8">
      {/* Redis缓存配置 */}
      <Card>
        <CardHeader>
          <CardTitle>Redis缓存配置</CardTitle>
          <CardDescription>
            配置Redis缓存连接信息，启用缓存可以提高检测效率，减少API调用
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-8">加载中...</div>
          ) : (
            <div className="space-y-6">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="redis-enabled">启用Redis缓存</Label>
                  <p className="text-sm text-muted-foreground">启用后将使用Redis缓存检测结果</p>
                </div>
                <Switch
                  id="redis-enabled"
                  checked={redisConfig.enabled}
                  onCheckedChange={(checked) => setRedisConfig({ ...redisConfig, enabled: checked })}
                />
              </div>

              {redisConfig.enabled && (
                <>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="redis-host">Redis地址</Label>
                      <Input
                        id="redis-host"
                        value={redisConfig.host}
                        onChange={(e) => setRedisConfig({ ...redisConfig, host: e.target.value })}
                        placeholder="localhost"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="redis-port">Redis端口</Label>
                      <Input
                        id="redis-port"
                        type="number"
                        min="1"
                        max="65535"
                        value={redisConfig.port}
                        onChange={(e) => setRedisConfig({ ...redisConfig, port: parseInt(e.target.value) || 6379 })}
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="redis-username">Redis用户名</Label>
                      <Input
                        id="redis-username"
                        value={redisConfig.username || ''}
                        onChange={(e) => setRedisConfig({ ...redisConfig, username: e.target.value })}
                        placeholder="(可选)"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="redis-password">Redis密码</Label>
                      <Input
                        id="redis-password"
                        type="password"
                        value={redisConfig.password}
                        onChange={(e) => setRedisConfig({ ...redisConfig, password: e.target.value })}
                        placeholder="(可选)"
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="redis-invalid-ttl">无效链接缓存过期时间（小时）</Label>
                    <Input
                      id="redis-invalid-ttl"
                      type="number"
                      min="1"
                      max="720"
                      value={redisConfig.invalid_ttl}
                      onChange={(e) => setRedisConfig({ ...redisConfig, invalid_ttl: parseInt(e.target.value) || 168 })}
                    />
                    <p className="text-sm text-muted-foreground">无效链接的统一缓存过期时间（1-720小时，默认168小时即7天）</p>
                  </div>
                </>
              )}

              <div className="flex justify-end">
                <Button
                  onClick={handleSaveRedisConfig}
                  disabled={savingRedis || !redisConfig.enabled}
                  className="w-32"
                >
                  {savingRedis ? '保存中...' : '保存配置'}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 频率控制配置 */}
      <Card>
        <CardHeader>
          <CardTitle>频率控制配置</CardTitle>
          <CardDescription>
            配置每个网盘平台的请求频率限制，避免因请求过多导致检测失败
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-8">加载中...</div>
          ) : (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
                {platforms.map((platform) => {
                  const config = rateConfigSettings[platform] || defaultRateConfig;
                  return (
                    <Card key={platform} className="p-4">
                      <div className="space-y-3">
                        <div className="font-semibold text-base mb-3 border-b pb-2">
                          {PLATFORM_NAMES[platform] || platform}
                        </div>
                        
                        <div className="space-y-3">
                          <div className="space-y-1.5">
                            <Label htmlFor={`${platform}-concurrency`} className="text-sm">
                              并发数
                            </Label>
                            <Input
                              id={`${platform}-concurrency`}
                              type="number"
                              min="1"
                              max="100"
                              value={config.concurrency}
                              onChange={(e) => updateRateConfig(platform, 'concurrency', parseInt(e.target.value) || 1)}
                              className="h-9"
                            />
                            <p className="text-xs text-muted-foreground">同时检测的链接数量（1-100）</p>
                          </div>

                          <div className="space-y-1.5">
                            <Label htmlFor={`${platform}-delay`} className="text-sm">
                              请求间隔（毫秒）
                            </Label>
                            <Input
                              id={`${platform}-delay`}
                              type="number"
                              min="0"
                              max="10000"
                              value={config.request_delay_ms}
                              onChange={(e) => updateRateConfig(platform, 'request_delay_ms', parseInt(e.target.value) || 0)}
                              className="h-9"
                            />
                            <p className="text-xs text-muted-foreground">每次请求之间的间隔</p>
                          </div>

                          <div className="space-y-1.5">
                            <Label htmlFor={`${platform}-max-rps`} className="text-sm">
                              每秒最大请求数
                            </Label>
                            <Input
                              id={`${platform}-max-rps`}
                              type="number"
                              min="0"
                              max="100"
                              value={config.max_requests_per_second}
                              onChange={(e) => updateRateConfig(platform, 'max_requests_per_second', parseInt(e.target.value) || 0)}
                              className="h-9"
                            />
                            <p className="text-xs text-muted-foreground">0表示不限制</p>
                          </div>

                          <div className="space-y-1.5">
                            <Label htmlFor={`${platform}-cache-ttl`} className="text-sm">
                              缓存有效期（小时）
                            </Label>
                            <Input
                              id={`${platform}-cache-ttl`}
                              type="number"
                              min="0"
                              max="720"
                              value={config.cache_ttl_hours ?? 24}
                              onChange={(e) => updateRateConfig(platform, 'cache_ttl_hours', parseInt(e.target.value) || 24)}
                              className="h-9"
                            />
                            <p className="text-xs text-muted-foreground">有效链接的缓存过期时间（0-720小时，默认24小时）</p>
                          </div>
                        </div>
                      </div>
                    </Card>
                  );
                })}
              </div>
              <div className="flex justify-end">
                <Button
                  onClick={handleSaveRateConfig}
                  disabled={saving}
                  className="w-32"
                >
                  {saving ? '保存中...' : '保存所有'}
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
