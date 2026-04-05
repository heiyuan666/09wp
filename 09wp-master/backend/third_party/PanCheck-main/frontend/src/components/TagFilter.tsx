import { useState, useEffect } from 'react';
import { Check } from 'lucide-react';
import { Button } from './ui/button';
import { getAllTags, type TagsResponse } from '@/api/scheduledTaskApi';

interface TagFilterProps {
  selectedTags: string[];
  onChange: (tags: string[]) => void;
}

export function TagFilter({ selectedTags, onChange }: TagFilterProps) {
  const [availableTags, setAvailableTags] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadTags();
  }, []);

  const loadTags = async () => {
    try {
      const response: TagsResponse = await getAllTags();
      setAvailableTags(response.tags);
    } catch (error) {
      console.error('Failed to load tags:', error);
    } finally {
      setLoading(false);
    }
  };

  const toggleTag = (tag: string) => {
    if (selectedTags.includes(tag)) {
      onChange(selectedTags.filter(t => t !== tag));
    } else {
      onChange([...selectedTags, tag]);
    }
  };

  if (loading) {
    return <div className="text-sm text-muted-foreground">加载标签中...</div>;
  }

  if (availableTags.length === 0) {
    return <div className="text-sm text-muted-foreground">暂无标签</div>;
  }

  return (
    <div className="space-y-2">
      <div className="text-sm font-medium">标签筛选</div>
      <div className="flex flex-wrap gap-2">
        {availableTags.map((tag) => (
          <Button
            key={tag}
            type="button"
            variant={selectedTags.includes(tag) ? 'default' : 'outline'}
            size="sm"
            onClick={() => toggleTag(tag)}
            className="h-8"
          >
            {selectedTags.includes(tag) && <Check className="mr-1 h-3 w-3" />}
            {tag}
          </Button>
        ))}
      </div>
      {selectedTags.length > 0 && (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => onChange([])}
          className="h-8 text-xs"
        >
          清除筛选
        </Button>
      )}
    </div>
  );
}

