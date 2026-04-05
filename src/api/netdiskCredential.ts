import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

/** 多账号轮询项：转存时按顺序轮流使用；勾选「暂停」则跳过 */
export interface INetdiskCookieAccount {
  name?: string
  cookie?: string
  /** 为 true 时该账号不参与轮询 */
  disabled?: boolean
  /** 该账号单独转存目录 ID（留空则用页面全局「转存目录」）；百度见 target_path */
  target_folder_id?: string
  /** 仅百度：该账号单独转存路径（留空则用全局路径） */
  target_path?: string
}

/** 各网盘转存 Cookie / Token 与自动转存开关（独立表 netdisk_credentials） */
export interface INetdiskCredential {
  id?: number
  quark_cookie?: string
  /** 夸克多账号轮询；有可用项时优先于单独 quark_cookie */
  quark_cookie_accounts?: INetdiskCookieAccount[]
  quark_auto_save?: boolean
  quark_target_folder_id?: string
  quark_ad_filter_enabled?: boolean
  quark_banned_keywords?: string
  pan115_cookie?: string
  pan115_cookie_accounts?: INetdiskCookieAccount[]
  pan115_auto_save?: boolean
  pan115_target_folder_id?: string
  tianyi_cookie?: string
  tianyi_cookie_accounts?: INetdiskCookieAccount[]
  tianyi_auto_save?: boolean
  tianyi_target_folder_id?: string
  pan123_cookie?: string
  pan123_cookie_accounts?: INetdiskCookieAccount[]
  pan123_auto_save?: boolean
  pan123_target_folder_id?: string
  baidu_cookie?: string
  baidu_cookie_accounts?: INetdiskCookieAccount[]
  baidu_auto_save?: boolean
  baidu_target_path?: string
  xunlei_cookie?: string
  /** 多账号迅雷 refresh_token 轮询 */
  xunlei_cookie_accounts?: INetdiskCookieAccount[]
  xunlei_auto_save?: boolean
  xunlei_target_folder_id?: string
  uc_cookie?: string
  uc_cookie_accounts?: INetdiskCookieAccount[]
  uc_auto_save?: boolean
  uc_target_folder_id?: string
  aliyun_refresh_token?: string
  /** 多账号 refresh_token 轮询；cookie 字段填 token 文本 */
  aliyun_refresh_token_accounts?: INetdiskCookieAccount[]
  aliyun_auto_save?: boolean
  aliyun_target_parent_file_id?: string
  /** 转存成功后是否将资源链接替换为您本人网盘生成的分享链接（各已接入网盘均会尝试） */
  replace_link_after_transfer?: boolean
}

export const getNetdiskCredentials = () =>
  request.get<ICommonResponse<INetdiskCredential>>('/system/netdisk-credentials')

export const updateNetdiskCredentials = (data: INetdiskCredential) =>
  request.put<ICommonResponse<unknown>>('/system/netdisk-credentials', data)
