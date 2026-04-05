import request from '@/utils/request'
import publicRequest from '@/utils/publicRequest'
import type { ICommonResponse } from '@/types/common'

export type IPageResult<T> = ICommonResponse<{ list: T[]; total: number } | any>

// 后台管理（管理员）
export const adminCategoryList = (params?: any) =>
  request.get<ICommonResponse<any>>('/admin/categories', { params })
export const adminCategoryCreate = (data: any) =>
  request.post<ICommonResponse<unknown>>('/admin/categories', data)
export const adminCategoryUpdate = (id: string, data: any) =>
  request.put<ICommonResponse<unknown>>(`/admin/categories/${id}`, data)
export const adminCategoryDelete = (id: string) =>
  request.delete<ICommonResponse<unknown>>(`/admin/categories/${id}`)

export const adminResourcePage = (params?: any) =>
  request.get<ICommonResponse<any>>('/admin/resources', { params })
export const adminResourceCreate = (data: any) =>
  request.post<ICommonResponse<unknown>>('/admin/resources', data)
export const adminResourceUpdate = (id: string, data: any) =>
  request.put<ICommonResponse<unknown>>(`/admin/resources/${id}`, data)
export const adminResourceDelete = (id: string) =>
  request.delete<ICommonResponse<unknown>>(`/admin/resources/${id}`)
export const adminResourceRetryTransfer = (id: string) =>
  request.post<ICommonResponse<unknown>>(`/admin/resources/${id}/retry-transfer`)
export const adminResourceTransferLogs = (id: string, params?: any) =>
  request.get<ICommonResponse<any>>(`/admin/resources/${id}/transfer-logs`, { params })

/** 管理端：上传 CSV/XLSX 表格并按固定字段导入资源，不触发自动转存。 */
export const adminResourceImportTable = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return request.post<ICommonResponse<any>>('/admin/resources/import-table', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 管理端：导出资源表格（CSV/XLSX），并把导出文件作为资源记录写入 resources 表。 */
export const adminResourceExportTable = (params: {
  format?: 'csv' | 'xlsx'
  title?: string
  category_id?: number | string
  status?: number | string
  limit?: number
  export_category_id?: number | string
  export_public?: boolean
  export_title?: string
}) =>
  request.get<ICommonResponse<any>>('/admin/resources/export-table', {
    params,
  })

/** 按链接识别网盘并一键转存，需要管理员且已配置对应网盘 Cookie/Token。 */
export const netdiskTransferByLink = (data: { link: string; passcode?: string }) =>
  request.post<ICommonResponse<unknown>>('/netdisk/transfer', data)
export const netdiskTransferBatchByLinks = (data: {
  items: Array<{ link: string; passcode?: string }>
}) =>
  request.post<
    ICommonResponse<{
      total: number
      success: number
      failed: number
      results: Array<{
        index: number
        link: string
        platform: string
        success: boolean
        message?: string
        data?: any
      }>
    }>
  >('/netdisk/transfer/batch', data)
export const adminResourceSyncTelegram = () =>
  request.post<ICommonResponse<{ synced: number; added: number; skipped: number }>>(
    '/admin/resources/sync-telegram',
  )
export const adminResourceBatchDelete = (ids: number[]) =>
  request.post<ICommonResponse<unknown>>('/admin/resources/batch-delete', { ids })
export const adminResourceBatchStatus = (ids: number[], status: number) =>
  request.post<ICommonResponse<unknown>>('/admin/resources/batch-status', { ids, status })
export const adminResourceCheckLinks = (
  ids?: number[],
  selectedPlatforms?: string[],
  oneByOne = false,
) =>
  request.post<
    ICommonResponse<{
      submission_id: number
      valid: number
      invalid: number
      unknown: number
      checked: number
      details?: Array<{ link: string; status: 'valid' | 'invalid' | 'unknown'; msg: string }>
    }>
  >('/admin/resources/check-links', {
    ids: ids || [],
    selectedPlatforms: selectedPlatforms || [],
    one_by_one: oneByOne,
  })

export const adminSubmissionList = (params?: {
  page?: number
  page_size?: number
  status?: string
  q?: string
  user_id?: string
}) => request.get<ICommonResponse<{ list: any[]; total: number }>>('/admin/submissions', { params })
export const adminSubmissionApprove = (id: number | string) =>
  request.post<ICommonResponse<{ resource_id: number; game_resource_id?: number }>>(
    `/admin/submissions/${id}/approve`,
  )
export const adminSubmissionReject = (id: number | string, data?: { reason?: string }) =>
  request.post<ICommonResponse<unknown>>(`/admin/submissions/${id}/reject`, data || {})

// 前台站点（无需管理员 token）
export const siteHome = () => publicRequest.get<ICommonResponse<any>>('/home')
export const siteCategories = () => publicRequest.get<ICommonResponse<any[]>>('/categories')
export const siteResourcePage = (params?: any) =>
  publicRequest.get<ICommonResponse<any>>('/resources', { params })
export const siteResourceDetail = (id: string) =>
  publicRequest.get<ICommonResponse<any>>(`/resources/${id}`)
export const siteResourceAccessLink = (id: string) =>
  publicRequest.post<ICommonResponse<{ status: string; link?: string; message?: string }>>(
    `/resources/${id}/access-link`,
  )
export const siteResourceLatestTransferLog = (id: string) =>
  publicRequest.get<
    ICommonResponse<{
      exists: boolean
      platform?: string
      message?: string
      own_share_url?: string
      created_at?: string
      filter_log?: any
    }>
  >(`/resources/${id}/transfer/latest-log`)
export const siteSearch = (params?: any) =>
  publicRequest.get<ICommonResponse<any>>('/search', { params })

/** 热搜榜，来自用户真实搜索统计。 */
export const siteHotSearch = (params?: { limit?: number }) =>
  publicRequest.get<ICommonResponse<{ list: { keyword: string; search_count: number }[] }>>(
    '/public/hot-search',
    { params },
  )

/** 豆瓣热门榜单 */
export interface IDoubanHotItem {
  title: string
  cover?: string
  url?: string
}

/** 豆瓣热门榜单，点击后跳转到搜索页并沿用 /search?q=...。 */
export const siteDoubanHot = (params?: { limit?: number; type?: string; tag?: string }) =>
  publicRequest.get<ICommonResponse<{ list: IDoubanHotItem[] }>>('/public/douban-hot', { params })

export const siteRegister = (data: any) =>
  publicRequest.post<ICommonResponse<unknown>>('/auth/register', data)
export const siteCaptcha = () =>
  publicRequest.get<ICommonResponse<{ captcha_id: string; svg: string; expires_at: string }>>(
    '/auth/captcha',
  )
export const siteRegisterSendCode = (data: {
  email: string
  captcha_id: string
  captcha_code: string
}) => publicRequest.post<ICommonResponse<{ expires_at: string }>>('/auth/register/send-code', data)
export const siteLogin = (data: any) =>
  publicRequest.post<ICommonResponse<any>>('/auth/login', data)
export const sitePasswordForgot = (data: { email: string }) =>
  publicRequest.post<ICommonResponse<{ reset_token: string; expires_at: string }>>(
    '/auth/password/forgot',
    data,
  )
export const sitePasswordReset = (data: {
  token: string
  newPassword: string
  confirmPassword: string
}) => publicRequest.post<ICommonResponse<unknown>>('/auth/password/reset', data)
export const siteMe = () => publicRequest.get<ICommonResponse<any>>('/user/profile')
export const siteChangePassword = (data: any) =>
  publicRequest.put<ICommonResponse<unknown>>('/user/password', data)
export const siteFavorites = () => publicRequest.get<ICommonResponse<any[]>>('/user/favorites')
export const siteFavoriteAdd = (resourceId: string) =>
  publicRequest.post<ICommonResponse<unknown>>(`/user/favorites/${resourceId}`)
export const siteFavoriteRemove = (resourceId: string) =>
  publicRequest.delete<ICommonResponse<unknown>>(`/user/favorites/${resourceId}`)

export const siteSubmissionCreate = (data: {
  title: string
  link: string
  category_id?: number
  game_id?: number
  description?: string
  extract_code?: string
  tags?: string
}) => publicRequest.post<ICommonResponse<any>>('/user/submissions', data)

export const siteMySubmissions = () =>
  publicRequest.get<ICommonResponse<any[]>>('/user/submissions')

// TG 频道管理（管理员）
export const telegramChannelList = () =>
  request.get<ICommonResponse<{ list: any[]; total: number }>>('/telegram/channels')
export const telegramChannelCreate = (data: any) =>
  request.post<ICommonResponse<any>>('/telegram/channels', data)
export const telegramChannelUpdate = (id: number | string, data: any) =>
  request.put<ICommonResponse<unknown>>(`/telegram/channels/${id}`, data)
export const telegramChannelDelete = (id: number | string) =>
  request.delete<ICommonResponse<unknown>>(`/telegram/channels/${id}`)
export const telegramChannelSync = (id: number | string) =>
  request.post<ICommonResponse<{ added: number; skipped: number }>>(`/telegram/channels/${id}/sync`)
export const telegramChannelBackfill = (id: number | string, data?: { limit?: number }) =>
  request.post<ICommonResponse<{ added: number; skipped: number; scanned: number }>>(
    `/telegram/channels/${id}/backfill`,
    data || {},
    { timeout: 180000 },
  )
export const telegramChannelSyncAll = () =>
  request.post<ICommonResponse<{ synced: number; added: number; skipped: number }>>(
    '/telegram/channels/sync-all',
  )
export const telegramChannelTest = (data: {
  bot_token: string
  channel_chat_id: string
  proxy_url?: string
}) =>
  request.post<ICommonResponse<{ ok: boolean; message: string }>>('/telegram/channels/test', data)

// TG MTProto 登录会话
export const telegramSessionStatus = () =>
  request.get<
    ICommonResponse<{
      has_api: boolean
      has_session: boolean
      need_password: boolean
      phone: string
    }>
  >('/telegram/session/status')
export const telegramSessionSendCode = (data: { phone: string }) =>
  request.post<ICommonResponse<{ message: string }>>('/telegram/session/send-code', data, {
    timeout: 60000,
  })
export const telegramSessionSignIn = (data: { code: string }) =>
  request.post<ICommonResponse<{ need_password: boolean; message: string }>>(
    '/telegram/session/sign-in',
    data,
    {
      timeout: 60000,
    },
  )
export const telegramSessionCheckPassword = (data: { password: string }) =>
  request.post<ICommonResponse<{ message: string }>>('/telegram/session/check-password', data, {
    timeout: 60000,
  })

// RSS 订阅管理（管理员）
export const rssSubscriptionList = () =>
  request.get<ICommonResponse<{ list: any[]; total: number }>>('/rss/subscriptions')
export const rssSubscriptionCreate = (data: any) =>
  request.post<ICommonResponse<any>>('/rss/subscriptions', data)
export const rssSubscriptionUpdate = (id: number | string, data: any) =>
  request.put<ICommonResponse<unknown>>(`/rss/subscriptions/${id}`, data)
export const rssSubscriptionDelete = (id: number | string) =>
  request.delete<ICommonResponse<unknown>>(`/rss/subscriptions/${id}`)
export const rssSubscriptionTest = (data: { feed_url: string }) =>
  request.post<ICommonResponse<{ ok: boolean; message: string }>>('/rss/subscriptions/test', data)
export const rssSubscriptionSync = (id: number | string) =>
  request.post<ICommonResponse<{ added: number; skipped: number }>>(`/rss/subscriptions/${id}/sync`)
export const rssSubscriptionSyncAll = () =>
  request.post<ICommonResponse<{ synced: number; added: number; skipped: number }>>(
    '/rss/subscriptions/sync-all',
  )

