import { Textarea } from './ui/textarea';
import { Label } from './ui/label';

interface CodeEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function CodeEditor({ value, onChange, placeholder }: CodeEditorProps) {
  return (
    <div className="space-y-2">
      <Label>转换脚本 (JavaScript)</Label>
      <Textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder || `// 示例：将rawData转换为字符串数组\nconst data = JSON.parse(rawData);\nreturn data.map(item => item.url);`}
        className="font-mono text-sm min-h-[200px]"
      />
      <div className="text-xs text-muted-foreground">
        <p>脚本应返回一个字符串数组。变量 <code className="bg-muted px-1 rounded">rawData</code> 包含curl命令的原始输出。</p>
      </div>
    </div>
  );
}

