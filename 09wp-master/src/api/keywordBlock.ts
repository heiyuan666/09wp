import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface IKeywordBlock {
  id?: number
  keyword: string
  enabled: boolean
  created_at?: string
  updated_at?: string
}

export const keywordBlockList = () =>
  request.get<ICommonResponse<{ list: IKeywordBlock[] }>>('/admin/keyword-blocks')

export const keywordBlockCreate = (data: { keyword: string; enabled?: boolean }) =>
  request.post<ICommonResponse<IKeywordBlock>>('/admin/keyword-blocks', data)

export const keywordBlockUpdate = (id: number, data: { keyword?: string; enabled?: boolean }) =>
  request.put<ICommonResponse<IKeywordBlock>>(`/admin/keyword-blocks/${id}`, data)

export const keywordBlockDelete = (id: number) =>
  request.delete<ICommonResponse<{ deleted: boolean }>>(`/admin/keyword-blocks/${id}`)

