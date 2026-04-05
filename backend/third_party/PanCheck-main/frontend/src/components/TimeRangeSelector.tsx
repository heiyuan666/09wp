import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

export type TimeRange = 'today' | 'last24h' | 'thisWeek' | 'last7d' | 'thisMonth' | 'last30d' | 'last90d';

interface TimeRangeSelectorProps {
  value: TimeRange;
  onChange: (range: TimeRange) => void;
}

export function TimeRangeSelector({
  value,
  onChange,
}: TimeRangeSelectorProps) {
  return (
    <Select value={value} onValueChange={(val) => onChange(val as TimeRange)}>
      <SelectTrigger className="w-[140px]">
        <SelectValue placeholder="选择时间范围" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="today">今天</SelectItem>
        <SelectItem value="last24h">最近24小时</SelectItem>
        <SelectItem value="thisWeek">本周</SelectItem>
        <SelectItem value="last7d">最近7天</SelectItem>
        <SelectItem value="thisMonth">本月</SelectItem>
        <SelectItem value="last30d">最近30天</SelectItem>
        <SelectItem value="last90d">最近90天</SelectItem>
      </SelectContent>
    </Select>
  );
}

