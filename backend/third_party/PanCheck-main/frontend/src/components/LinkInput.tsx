import { useState } from 'react';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { toast } from 'sonner';
import { PLATFORM_NAMES } from '@/utils/constants';
import type { Platform } from '@/types';

interface LinkInputProps {
  onCheck: (links: string[], selectedPlatforms?: Platform[]) => void;
  loading?: boolean;
}

export function LinkInput({ onCheck, loading }: LinkInputProps) {
  const [links, setLinks] = useState('');
  const [selectedPlatforms, setSelectedPlatforms] = useState<Platform[]>([]);

  const handleSubmit = () => {
    const linkArray = links
      .split('\n')
      .map(link => link.trim())
      .filter(link => link.length > 0);

    if (linkArray.length === 0) {
      toast.error('请输入至少一个链接');
      return;
    }

    // 传递选中的平台（如果选择了的话）
    onCheck(linkArray, selectedPlatforms.length > 0 ? selectedPlatforms : undefined);
  };

  const togglePlatform = (platform: Platform) => {
    setSelectedPlatforms(prev => {
      if (prev.includes(platform)) {
        return prev.filter(p => p !== platform);
      } else {
        return [...prev, platform];
      }
    });
  };

  // 获取所有可选的平台（排除 unknown）
  const availablePlatforms: Platform[] = ['quark', 'uc', 'baidu', 'tianyi', 'pan123', 'pan115', 'aliyun', 'xunlei', 'cmcc'];

  // 检查是否所有平台都被选中
  const isAllSelected = selectedPlatforms.length === availablePlatforms.length;

  // 全选/取消全选
  const toggleSelectAll = () => {
    if (isAllSelected) {
      setSelectedPlatforms([]);
    } else {
      setSelectedPlatforms([...availablePlatforms]);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>网盘链接检查</CardTitle>
        <CardDescription>
          批量提交网盘分享链接进行有效性检测
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="links">链接列表（每行一个）</Label>
          <Textarea
            id="links"
            placeholder="请输入网盘分享链接，每行一个&#10;例如：&#10;https://pan.baidu.com/s/xxxxx&#10;https://www.aliyundrive.com/s/xxxxx"
            value={links}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setLinks(e.target.value)}
            rows={10}
            disabled={loading}
          />
        </div>

        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <Label>选择要即时检测的网盘平台（可选，多选）</Label>
            <label className="flex items-center space-x-2 cursor-pointer">
              <input
                type="checkbox"
                checked={isAllSelected}
                onChange={toggleSelectAll}
                disabled={loading}
                className="w-4 h-4 rounded border-gray-300"
              />
              <span className="text-sm font-medium">全选</span>
            </label>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-2 p-3 border rounded-md">
            {availablePlatforms.map(platform => (
              <label key={platform} className="flex items-center space-x-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={selectedPlatforms.includes(platform)}
                  onChange={() => togglePlatform(platform)}
                  disabled={loading}
                  className="w-4 h-4 rounded border-gray-300"
                />
                <span className="text-sm">{PLATFORM_NAMES[platform]}</span>
              </label>
            ))}
          </div>
          <p className="text-xs text-muted-foreground">
            选择后，这些平台的链接将立即检测；未选择的平台链接将通过定时任务检测。如果全部选择，则等同于即时检测所有链接。
          </p>
        </div>

        <Button onClick={handleSubmit} disabled={loading} className="w-full">
          {loading ? '检测中...' : '开始检测'}
        </Button>
      </CardContent>
    </Card>
  );
}

