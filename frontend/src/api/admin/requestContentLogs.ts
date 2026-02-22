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

/**
 * 标准化聊天消息
 */
export interface ChatMessage {
  role: 'system' | 'user' | 'assistant'
  content: string
}

/**
 * 从 Anthropic content blocks 中提取文本
 * content 可能是 string 或 [{type: "text", text: "..."}]
 */
function extractAnthropicContent(content: any): string {
  if (typeof content === 'string') return content
  if (Array.isArray(content)) {
    return content
      .filter((block: any) => block.type === 'text' && typeof block.text === 'string')
      .map((block: any) => block.text)
      .join('\n')
  }
  return String(content ?? '')
}

/**
 * 从 Gemini parts 中提取文本
 * parts: [{text: "..."}, ...]
 */
function extractGeminiParts(parts: any): string {
  if (!Array.isArray(parts)) return String(parts ?? '')
  return parts
    .filter((p: any) => typeof p.text === 'string')
    .map((p: any) => p.text)
    .join('\n')
}

/**
 * 将角色名统一映射为标准角色
 */
function normalizeRole(role: string): ChatMessage['role'] {
  const r = (role || '').toLowerCase()
  if (r === 'model') return 'assistant'
  if (r === 'system') return 'system'
  if (r === 'user') return 'user'
  if (r === 'assistant') return 'assistant'
  // 未知角色默认当 user
  return 'user'
}

/**
 * 将各平台 messages 格式标准化为 ChatMessage[]
 * 自动检测 OpenAI / Anthropic content blocks / Gemini 格式
 */
export function normalizeMessages(messages: any): ChatMessage[] {
  if (!messages || !Array.isArray(messages)) return []

  return messages.map((msg: any) => {
    const role = normalizeRole(msg.role)

    // Gemini 格式: { role: "user"|"model", parts: [{text: "..."}] }
    if (msg.parts && !msg.content) {
      return { role, content: extractGeminiParts(msg.parts) }
    }

    // Anthropic content blocks: { role, content: [{type:"text", text:"..."}] }
    // 也兼容 OpenAI 的 string content
    return { role, content: extractAnthropicContent(msg.content) }
  })
}

export const requestContentLogsAPI = {
  list,
  getById,
  getSession
}

export default requestContentLogsAPI
