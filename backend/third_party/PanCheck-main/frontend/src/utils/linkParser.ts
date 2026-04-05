import { Platform } from '@/types';

const PLATFORM_PATTERNS: Record<Platform, RegExp> = {
  quark: /(?:https?:\/\/)?(?:pan\.quark\.cn|quark\.cn|pan\.qoark\.cn)\/s\/[a-zA-Z0-9]+/i,
  uc: /(?:https?:\/\/)?(?:yun\.uc\.cn|uc\.cn)\/s\/[a-zA-Z0-9]+/i,
  baidu: /(?:https?:\/\/)?(?:pan\.baidu\.com)\/s\/[a-zA-Z0-9_-]+/i,
  tianyi: /(?:https?:\/\/)?(?:cloud\.189\.cn|h5\.cloud\.189\.cn)\/(?:t\/[a-zA-Z0-9]+|web\/share\?code=[a-zA-Z0-9]+|share\.html#\/t\/[a-zA-Z0-9]+)/i,
  pan123: /(?:https?:\/\/)?(?:123pan\.com|123pan\.cn|123684\.com|123685\.com|123912\.com|123592\.com|123865\.com)\/s\/[a-zA-Z0-9-]+/i,
  pan115: /(?:https?:\/\/)?(?:115\.com|115cdn\.com|anxia\.com)\/s\/[a-zA-Z0-9]+/i,
  aliyun: /(?:https?:\/\/)?(?:www\.aliyundrive\.com|www\.alipan\.com)\/s\/[a-zA-Z0-9]+/i,
  xunlei: /(?:https?:\/\/)?(?:pan\.xunlei\.com)\/s\/[a-zA-Z0-9]+/i,
  cmcc: /(?:https?:\/\/)?(?:yun\.139\.com\/shareweb\/#\/w\/i\/|caiyun\.139\.com\/m\/i\?)[a-zA-Z0-9]+/i,
  unknown: /./,
};

export function parseLink(link: string): Platform {
  const trimmed = link.trim();
  if (!trimmed) return 'unknown';

  for (const [platform, pattern] of Object.entries(PLATFORM_PATTERNS)) {
    if (pattern.test(trimmed)) {
      return platform as Platform;
    }
  }

  return 'unknown';
}

export function normalizeLink(link: string): string {
  const trimmed = link.trim();
  if (!trimmed) return trimmed;
  
  if (!/^https?:\/\//.test(trimmed)) {
    return 'https://' + trimmed;
  }
  
  return trimmed;
}

