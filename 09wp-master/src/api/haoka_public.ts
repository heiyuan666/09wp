import request from '@/utils/request'
import type { IHaokaCategory, IHaokaProductItem, IHaokaProductDetail, IHaokaSkuItem } from './haoka'

export type { IHaokaCategory, IHaokaProductItem, IHaokaProductDetail, IHaokaSkuItem }

export function haokaPublicCategories() {
  return request.get<{ code: number; message: string; data: IHaokaCategory[] }>('/public/haoka/categories')
}

export function haokaPublicProducts(params?: {
  category_id?: number
  operator?: string
  flag?: string
  q?: string
  page?: number
  page_size?: number
}) {
  return request.get<{
    code: number
    message: string
    data: { list: IHaokaProductItem[]; total: number; page: number; page_size: number }
  }>('/public/haoka/products', { params })
}

export function haokaPublicProductDetail(id: number) {
  return request.get<{
    code: number
    message: string
    data: { product: IHaokaProductDetail; category_name: string; skus: IHaokaSkuItem[] }
  }>(`/public/haoka/products/${id}`)
}

