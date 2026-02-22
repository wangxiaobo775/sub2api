<template>
  <BaseDialog
    :show="show"
    :title="t('admin.requestContentLogs.sessionDetail')"
    width="extra-wide"
    @close="$emit('close')"
  >
    <div v-if="loading" class="flex items-center justify-center py-12">
      <svg class="h-8 w-8 animate-spin text-primary-500" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="opacity-25" />
        <path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" class="opacity-75" />
      </svg>
    </div>

    <div v-else-if="sessionLogs.length > 0" class="space-y-4">
      <!-- 会话概要 -->
      <div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800/50">
        <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <div>
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.session') }}</span>
            <p class="font-mono text-sm">{{ fingerprint }}</p>
          </div>
          <div>
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.sessionRequests') }}</span>
            <p class="text-sm">{{ sessionLogs.length }}</p>
          </div>
          <div>
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.user') }}</span>
            <p class="text-sm">{{ sessionLogs[0]?.user_email || '-' }}</p>
          </div>
          <div>
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.sessionTotalMessages') }}</span>
            <p class="text-sm">{{ lastMessageCount }}</p>
          </div>
        </div>
      </div>

      <!-- 完整对话流 -->
      <div>
        <div class="mb-2 flex items-center justify-between">
          <!-- Tab 切换 -->
          <div class="flex items-center gap-1 rounded-lg bg-gray-100 p-0.5 dark:bg-gray-800">
            <button
              class="rounded-md px-3 py-1 text-xs font-medium transition-colors"
              :class="viewMode === 'chat'
                ? 'bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-gray-100'
                : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
              @click="viewMode = 'chat'"
            >
              {{ t('admin.requestContentLogs.chatView') }}
            </button>
            <button
              class="rounded-md px-3 py-1 text-xs font-medium transition-colors"
              :class="viewMode === 'json'
                ? 'bg-white text-gray-900 shadow-sm dark:bg-gray-700 dark:text-gray-100'
                : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
              @click="viewMode = 'json'"
            >
              {{ t('admin.requestContentLogs.rawJson') }}
            </button>
          </div>
          <button
            class="text-xs text-primary-600 hover:text-primary-800 dark:text-primary-400"
            @click="copyAll"
          >
            {{ copied ? t('common.copied') : t('common.copy') }}
          </button>
        </div>

        <!-- 聊天气泡视图 -->
        <div v-if="viewMode === 'chat'" class="max-h-[60vh] overflow-auto rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800">
          <ChatMessageView :messages="mergedMessages" :dividers="dividers" />
        </div>

        <!-- 原始 JSON 视图 -->
        <div v-else class="max-h-[60vh] overflow-auto rounded-lg bg-gray-900 p-4 space-y-3">
          <div
            v-for="(entry, idx) in sessionLogs"
            :key="entry.id"
            class="border-l-2 pl-3"
            :class="idx === 0 ? 'border-green-500' : 'border-blue-500'"
          >
            <div class="mb-1 flex items-center gap-2 text-xs text-gray-500">
              <span class="font-mono">
                #{{ idx + 1 }}
              </span>
              <span>{{ formatTime(entry.created_at) }}</span>
              <span class="rounded bg-gray-700 px-1.5 py-0.5 text-gray-300">
                offset={{ entry.message_offset }} count={{ entry.message_count }}
              </span>
              <span v-if="entry.model" class="rounded bg-blue-900/50 px-1.5 py-0.5 text-blue-300">
                {{ entry.model }}
              </span>
            </div>
            <pre class="text-xs text-green-400 whitespace-pre-wrap break-words">{{ formatMessages(entry.messages) }}</pre>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="py-8 text-center text-gray-500">
      {{ t('common.noData') }}
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { RequestContentLogDetail } from '@/api/admin/requestContentLogs'
import { normalizeMessages } from '@/api/admin/requestContentLogs'
import type { ChatMessage } from '@/api/admin/requestContentLogs'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ChatMessageView from './ChatMessageView.vue'
import type { DividerInfo } from './ChatMessageView.vue'

const props = defineProps<{
  show: boolean
  fingerprint: string
}>()

defineEmits(['close'])

const { t } = useI18n()

const sessionLogs = ref<RequestContentLogDetail[]>([])
const loading = ref(false)
const copied = ref(false)
const viewMode = ref<'chat' | 'json'>('chat')

const lastMessageCount = computed(() => {
  if (sessionLogs.value.length === 0) return 0
  const last = sessionLogs.value[sessionLogs.value.length - 1]
  return last.message_count || 0
})

/**
 * 展示最后一条日志的完整对话（每条日志现在包含完整 messages）
 * 在同一 session 中，最后一条日志的 messages 是最完整的
 */
const mergedMessages = computed<ChatMessage[]>(() => {
  if (sessionLogs.value.length === 0) return []
  const last = sessionLogs.value[sessionLogs.value.length - 1]
  return normalizeMessages(last.messages)
})

/**
 * 在每次请求的新增消息起始位置插入时间戳分隔线
 * 通过 message_count 反推每次请求新增了哪些消息
 */
const dividers = computed<DividerInfo[]>(() => {
  const result: DividerInfo[] = []
  const totalMessages = mergedMessages.value.length

  for (let i = 0; i < sessionLogs.value.length; i++) {
    const entry = sessionLogs.value[i]

    // 找到这条请求之前的消息数（即上一条请求的 message_count）
    let prevCount = 0
    if (i > 0) {
      prevCount = sessionLogs.value[i - 1].message_count || 0
    }

    // 只在有新增消息时插入分隔线（跳过第一条）
    if (i > 0 && prevCount < totalMessages) {
      result.push({
        beforeIndex: prevCount,
        label: `#${i + 1} · ${formatTime(entry.created_at)}${entry.model ? ' · ' + entry.model : ''}`
      })
    }
  }
  return result
})

const loadSession = async (fp: string) => {
  if (!fp) return
  loading.value = true
  sessionLogs.value = []
  try {
    sessionLogs.value = await adminAPI.requestContentLogs.getSession(fp)
  } catch (error) {
    console.error('Failed to load session:', error)
  } finally {
    loading.value = false
  }
}

const formatMessages = (messages: any): string => {
  if (!messages) return ''
  try {
    return JSON.stringify(messages, null, 2)
  } catch {
    return String(messages)
  }
}

const formatTime = (iso: string) => {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

const copyAll = async () => {
  try {
    const allMessages = sessionLogs.value.map(log => formatMessages(log.messages)).join('\n\n---\n\n')
    await navigator.clipboard.writeText(allMessages)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // fallback
  }
}

watch(
  () => props.show,
  (show) => {
    if (show && props.fingerprint) {
      loadSession(props.fingerprint)
    }
  }
)
</script>
