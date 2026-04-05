export type Platform = 
  | 'quark' 
  | 'uc' 
  | 'baidu' 
  | 'tianyi' 
  | 'pan123' 
  | 'pan115' 
  | 'aliyun' 
  | 'xunlei' 
  | 'cmcc' 
  | 'unknown';

export type SubmissionStatus = 'pending' | 'checked';

export interface CheckLinksRequest {
  links: string[];
  selected_platforms?: Platform[]; // 选择的平台（多选），如果全部选择则等同于即时检测所有链接
}

export interface CheckLinksResponse {
  submission_id: number;
  invalid_links: string[];
  pending_links: string[];
  valid_links: string[];
  total_duration?: number;
  invalid_format_count: number; // 不规范链接数量
  duplicate_count: number; // 重复链接数量
}

export interface SubmissionRecord {
  id: number;
  original_links: string[];
  pending_links: string[];
  valid_links: string[];
  selected_platforms?: Platform[]; // 用户提交时选择的网盘平台类型
  status: SubmissionStatus;
  total_duration?: number;
  total_links: number;
  client_ip: string;
  browser?: string;
  os?: string;
  device?: string;
  language?: string;
  country?: string;
  region?: string;
  city?: string;
  created_at: string;
  updated_at: string;
  checked_at?: string;
}

export interface LinkInfo {
  link: string;
  platform: Platform;
  status?: 'valid' | 'invalid' | 'pending';
}

