import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

// 迁移自 Next 的 `lib/utils.ts`
// 用于把 Tailwind className 做合并/去重，便于 1:1 复刻样式。
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

