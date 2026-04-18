import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface ICleanupTaskLogItem {
  id: number
  task: string
  resource_id: number
  platform: string
  action: string
  status: 'success' | 'failed' | 'skipped'
  message?: string
  created_at?: string
}

export const adminCleanupLogList = (params?: {
  page?: number
  page_size?: number
  task?: string
  status?: string
  platform?: string
  resource_id?: string | number
}) =>
  request.get<ICommonResponse<{ list: ICleanupTaskLogItem[]; total: number }>>('/admin/cleanup-logs', { params })

