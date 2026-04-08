import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface IGameReview {
  id: number
  game_id: number
  user_id: number
  rating: number
  content: string
  status: number
  helpful_count: number
  unhelpful_count: number
  created_at: string
  updated_at: string
}

export const adminGameReviewPage = (params?: any) =>
  request.get<ICommonResponse<{ list: IGameReview[]; total: number }>>('/game/admin/reviews', { params })

export const adminGameReviewSetStatus = (id: number, status: 0 | 1) =>
  request.put<ICommonResponse<unknown>>(`/game/admin/reviews/${id}/status`, { status })

export const adminGameReviewDelete = (id: number) =>
  request.delete<ICommonResponse<unknown>>(`/game/admin/reviews/${id}`)

