import request from '@/utils/request'

export interface IHaokaCategory {
  id: number
  name: string
  slug: string
  status: number
}

export interface IHaokaSku {
  sku_id: number
  sku_name: string
  desc: string
}

// 外部接口返回字段名与后端 service.HaokaExternalProduct 保持一致（camelCase）
export interface IHaokaExternalSku {
  SkuID: number
  SkuName: string
  Desc: string
}

export interface IHaokaExternalProduct {
  productID: number
  productName: string
  mainPic: string
  area: string
  disableArea: string
  littlepicture: string
  netAddr: string
  flag: boolean
  numberSel: number
  operator: string
  BackMoneyType: string
  Taocan: string
  Rule: string
  Age1: number
  Age2: number
  PriceTime: string
  Skus: IHaokaExternalSku[]
}

export interface IHaokaProductItem {
  id: number
  category_id: number
  category_name?: string
  product_id: number
  product_name: string
  operator?: string
  flag: boolean
  main_pic?: string
  area?: string
  disable_area?: string
  little_picture?: string
  net_addr?: string
  number_sel?: number
  back_money_type?: string
  taocan?: string
  rule?: string
  age1?: number
  age2?: number
  price_time?: string
}

export function haokaCategories() {
  return request.get<{ code: number; message: string; data: IHaokaCategory[] }>('/admin/haoka/categories')
}

export function haokaQueryProducts(payload: { user_id: string; secret: string; product_id?: string }) {
  return request.post<{ code: number; message: string; data: { list: IHaokaExternalProduct[] } }>(
    '/admin/haoka/query-products',
    payload,
  )
}

export function haokaSync(payload: { user_id: string; secret: string; product_id?: string }) {
  return request.post<{ code: number; message: string; data: any }>('/admin/haoka/sync', payload)
}

export function haokaListProducts(params?: {
  category_id?: string
  operator?: string
  flag?: string
  page?: number
  page_size?: number
}) {
  return request.get<{
    code: number
    message: string
    data: { list: IHaokaProductItem[]; total: number; page: number; page_size: number }
  }>(
    '/admin/haoka/products',
    { params },
  )
}

export function haokaUpsertFromExternal(product: IHaokaExternalProduct) {
  return request.post<{ code: number; message: string; data: any }>('/admin/haoka/products/upsert', product)
}

export interface IHaokaSkuItem {
  sku_id: number
  sku_name: string
  desc: string
}

export interface IHaokaProductDetail extends IHaokaProductItem {
  category_name?: string
  operator?: string
  skus?: IHaokaSkuItem[]
}

export function haokaProductDetail(id: number) {
  return request.get<{
    code: number
    message: string
    data: { product: IHaokaProductDetail; category_name: string; skus: IHaokaSkuItem[] }
  }>(`/admin/haoka/products/${id}`)
}

export function haokaSetProductFlag(id: number, payload: { flag: boolean }) {
  return request.put<{ code: number; message: string; data: any }>(`/admin/haoka/products/${id}/flag`, payload)
}

export function haokaUpdateProduct(id: number, payload: any) {
  return request.put<{ code: number; message: string; data: any }>(`/admin/haoka/products/${id}`, payload)
}

export function haokaCreateProduct(payload: any) {
  return request.post<{ code: number; message: string; data: { id: number } }>('/admin/haoka/products', payload)
}

