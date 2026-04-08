import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface IGameResourceFeedback {
  id?: number
  game_id: number
  game_resource_id: number
  download_url?: string
  extract_code?: string
  type: string
  content: string
  contact?: string
  status?: string
  created_at?: string
}

export const adminGameFeedbackList = (params?: {
  page?: number
  page_size?: number
  status?: string
  type?: string
  game_id?: string | number
}) =>
  request.get<ICommonResponse<{ list: IGameResourceFeedback[]; total: number }>>('/admin/game-feedbacks', { params })

export const adminGameFeedbackUpdateStatus = (id: number, status: 'pending' | 'processed') =>
  request.put<ICommonResponse<{ id: number; status: string }>>(`/admin/game-feedbacks/${id}/status`, { status })

