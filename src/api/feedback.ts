import request from '@/utils/request'
import publicRequest from '@/utils/publicRequest'
import type { ICommonResponse } from '@/types/common'

export interface IResourceFeedback {
  id?: number
  resource_id: number
  type: string
  content: string
  contact?: string
  status?: string
  created_at?: string
}

// 前台公开提交反馈
export const feedbackCreate = (data: Omit<IResourceFeedback, 'id' | 'status' | 'created_at'>) =>
  publicRequest.post<ICommonResponse<IResourceFeedback>>('/feedbacks', data)

// 后台：反馈列表
export const feedbackAdminList = (params?: { page?: number; page_size?: number; status?: string; type?: string }) =>
  request.get<ICommonResponse<{ list: IResourceFeedback[]; total: number }>>('/admin/feedbacks', { params })

// 后台：更新反馈处理状态
export const feedbackAdminUpdateStatus = (id: number, status: 'pending' | 'processed') =>
  request.put<ICommonResponse<{ id: number; status: string }>>(`/admin/feedbacks/${id}/status`, { status })
