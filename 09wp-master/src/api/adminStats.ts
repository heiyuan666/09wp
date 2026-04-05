import request from '@/utils/request'

export interface IAdminStats {
  resources_total: number
  resources_online: number
  users_total: number
  categories_total: number
  submissions_pending: number
  feedbacks_total: number
  searches_total: number
}

export function adminStats() {
  return request.get<{ code: number; message: string; data: IAdminStats }>('/admin/stats')
}

