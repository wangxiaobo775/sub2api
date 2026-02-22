/**
 * Admin Request Content Logs API endpoints
 * 请求内容日志管理 API
 */

import { apiClient } from '../client'
import type { PaginatedResponse, FetchOptions } from '@/types'

/**
 * 请求内容日志列表项
 */
export interface RequestContentLogItem {
  id: number
  user_id: number
  api_key_id: number
  model: string
  platform: string
  ip_address: string
  user_agent: string
  created_at: string
  user_email?: string
  api_key_name?: string
  session_fingerprint?: string
  message_offset?: number
  message_count?: number
}

/**
 * 请求内容日志详情（含 messages）
 */
export interface RequestContentLogDetail extends RequestContentLogItem {
  messages: any
}

/**
 * 查询过滤参数
 */
export interface RequestContentLogFilters {
  user_id?: number
  api_key_id?: number
  model?: string
  platform?: string
  session_fingerprint?: string
  start_date?: string
  end_date?: string
}

/**
 * 分页查询请求内容日志
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: RequestContentLogFilters,
  options?: FetchOptions
): Promise<PaginatedResponse<RequestContentLogItem>> {
  const params: Record<string, any> = {
    page,
    page_size: pageSize,
    ...filters
  }

  // 清理空值
  Object.keys(params).forEach(key => {
    if (params[key] === undefined || params[key] === null || params[key] === '') {
      delete params[key]
    }
  })

  const { data } = await apiClient.get<PaginatedResponse<RequestContentLogItem>>(
    '/admin/request-content-logs',
    { params, signal: options?.signal }
  )
  return data
}

/**
 * 查询请求内容日志详情
 */
export async function getById(id: number): Promise<RequestContentLogDetail> {
  const { data } = await apiClient.get<RequestContentLogDetail>(
    `/admin/request-content-logs/${id}`
  )
  return data
}

/**
 * 按会话查询完整对话流
 */
export async function getSession(fingerprint: string): Promise<RequestContentLogDetail[]> {
  const { data } = await apiClient.get<RequestContentLogDetail[]>(
    `/admin/request-content-logs/session/${fingerprint}`
  )
  return data
}

export const requestContentLogsAPI = {
  list,
  getById,
  getSession
}

export default requestContentLogsAPI
