import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'
import type { IUserDetailResponse } from '@/types/system/user'
import type {
  ILoginParams,
  ILoginResponse,
  IUserPermissionsResponse,
  ILoginLogParams,
} from '@/types/login'

export const login = (params: ILoginParams) => {
  return request.post<ILoginResponse>('/login', params)
}

/** 管理后台扫码登录：创建会话，二维码内容为 App 可识别的 dfannetdisk://qr-admin-login?sid= */
export const adminQrLoginCreate = () => {
  return request.post<
    ICommonResponse<{
      sid: string
      expires_at: string
      qr_payload: string
      qr_payload_alt: { type: string; sid: string }
      for_admin: boolean
    }>
  >('/auth/qr/create', { for_admin: true })
}

/** 轮询管理后台扫码状态：pending | confirmed | expired */
export const adminQrLoginStatus = (sid: string) => {
  return request.get<
    ICommonResponse<{
      status: string
      token?: string
      user?: Record<string, unknown>
      for_admin?: boolean
    }>
  >(`/auth/qr/status/${encodeURIComponent(sid)}`)
}

/**
 * 获取用户权限（菜单权限和按钮权限）
 */
export const userPermissions = () => {
  return request.get<IUserPermissionsResponse>('/users/permissions')
}

/**
 * 获取用户信息
 */
export const userInfoRequest = () => {
  return request.get<IUserDetailResponse>('/users/info')
}

// 记录用户登录日志
export const addLoginLog = (data: ILoginLogParams) => {
  return request.put('/users/log', data)
}
