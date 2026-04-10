import request from '@/utils/request'
import type { ICommonResponse } from '@/types/common'

export interface ISoftwareCategory {
  id: number
  name: string
  slug: string
  sort_order: number
  status: number
  created_at: string
  updated_at: string
}

export interface ISoftwareItem {
  id: number
  category_id: number
  name: string
  summary: string
  version: string
  cover: string
  cover_thumb: string
  screenshots: string[]
  size: string
  platforms: string
  website: string
  download_direct: string[]
  download_pan: string[]
  download_extract: string
  published_at?: string
  updated_at_override?: string
  status: number
  created_at: string
  updated_at: string
}

export interface ISoftwareVersionItem {
  id: number
  software_id: number
  version: string
  release_notes: string
  published_at?: string
  download_direct: string[]
  download_pan: string[]
  download_extract: string
  created_at: string
  updated_at: string
}

export const softwareUploadCover = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return request.post<ICommonResponse<{ url: string; thumb_url: string; preview_url?: string }>>(
    '/game/software/upload-cover',
    form,
    {
      headers: { 'Content-Type': 'multipart/form-data' },
    },
  )
}

// 分类
export const softwareCategoryList = () =>
  request.get<ICommonResponse<{ list: ISoftwareCategory[] }>>('/game/software/categories')
export const softwareCategoryCreate = (data: Partial<ISoftwareCategory>) =>
  request.post<ICommonResponse<unknown>>('/game/software/categories', data)
export const softwareCategoryUpdate = (id: number | string, data: Partial<ISoftwareCategory>) =>
  request.put<ICommonResponse<unknown>>(`/game/software/categories/${id}`, data)
export const softwareCategoryDelete = (id: number | string) =>
  request.delete<ICommonResponse<unknown>>(`/game/software/categories/${id}`)
export const softwareCategorySort = (id: number | string, sort_order: number) =>
  request.put<ICommonResponse<unknown>>(`/game/software/categories/${id}/sort`, { sort_order })

// 软件
export const softwareList = (params?: Record<string, unknown>) =>
  request.get<ICommonResponse<{ list: ISoftwareItem[]; total: number }>>('/game/software', { params })

export const softwareDetail = (id: number | string) =>
  request.get<ICommonResponse<{ software: ISoftwareItem; versions: ISoftwareVersionItem[] }>>(
    `/game/software/${id}`,
  )
export const softwareCreate = (data: Partial<ISoftwareItem>) =>
  request.post<ICommonResponse<unknown>>('/game/software', data)
export const softwareUpdate = (id: number | string, data: Partial<ISoftwareItem>) =>
  request.put<ICommonResponse<unknown>>(`/game/software/${id}`, data)
export const softwareDelete = (id: number | string) =>
  request.delete<ICommonResponse<unknown>>(`/game/software/${id}`)

// 版本（按软件）
export const softwareVersionList = (softwareId: number | string) =>
  request.get<ICommonResponse<{ list: ISoftwareVersionItem[] }>>(`/game/software/${softwareId}/versions`)
export const softwareVersionCreate = (softwareId: number | string, data: Partial<ISoftwareVersionItem>) =>
  request.post<ICommonResponse<unknown>>(`/game/software/${softwareId}/versions`, data)
export const softwareVersionUpdate = (versionId: number | string, data: Partial<ISoftwareVersionItem>) =>
  request.put<ICommonResponse<unknown>>(`/game/software/versions/${versionId}`, data)
export const softwareVersionDelete = (versionId: number | string) =>
  request.delete<ICommonResponse<unknown>>(`/game/software/versions/${versionId}`)

