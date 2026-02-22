<template>
  <BaseDialog
    :show="show"
    :title="t('admin.requestContentLogs.detail')"
    width="extra-wide"
    @close="$emit('close')"
  >
    <div v-if="loading" class="flex items-center justify-center py-12">
      <svg class="h-8 w-8 animate-spin text-primary-500" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="opacity-25" />
        <path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" class="opacity-75" />
      </svg>
    </div>

    <div v-else-if="logDetail" class="space-y-4">
      <!-- 基本信息 -->
      <div class="grid grid-cols-2 gap-4 rounded-lg bg-gray-50 p-4 dark:bg-gray-800/50">
        <div>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.time') }}</span>
          <p class="text-sm">{{ formatTime(logDetail.created_at) }}</p>
        </div>
        <div>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.user') }}</span>
          <p class="text-sm">{{ logDetail.user_email || '-' }}</p>
        </div>
        <div>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.apiKey') }}</span>
          <p class="text-sm">{{ logDetail.api_key_name || '-' }}</p>
        </div>
        <div>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.model') }}</span>
          <p class="text-sm">
            <span class="inline-flex items-center rounded-md bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
              {{ logDetail.model || '-' }}
            </span>
          </p>
        </div>
        <div>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.platform') }}</span>
          <p class="text-sm">{{ logDetail.platform }}</p>
        </div>
        <div>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">IP</span>
          <p class="text-sm font-mono">{{ logDetail.ip_address }}</p>
        </div>
        <div class="col-span-2">
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">User-Agent</span>
          <p class="text-xs font-mono break-all text-gray-600 dark:text-gray-300">{{ logDetail.user_agent }}</p>
        </div>
      </div>

      <!-- Messages 内容 -->
      <div>
        <div class="mb-2 flex items-center justify-between">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Messages</span>
          <button
            class="text-xs text-primary-600 hover:text-primary-800 dark:text-primary-400"
            @click="copyMessages"
          >
            {{ copied ? t('common.copied') : t('common.copy') }}
          </button>
        </div>
        <div class="max-h-[60vh] overflow-auto rounded-lg bg-gray-900 p-4">
          <pre class="text-xs text-green-400 whitespace-pre-wrap break-words">{{ formattedMessages }}</pre>
        </div>
      </div>
    </div>

    <div v-else class="py-8 text-center text-gray-500">
      {{ t('common.loadFailed') }}
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { RequestContentLogDetail as LogDetail } from '@/api/admin/requestContentLogs'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{
  show: boolean
  logId: number
}>()

defineEmits(['close'])

const { t } = useI18n()

const logDetail = ref<LogDetail | null>(null)
const loading = ref(false)
const copied = ref(false)

const formattedMessages = computed(() => {
  if (!logDetail.value?.messages) return ''
  try {
    return JSON.stringify(logDetail.value.messages, null, 2)
  } catch {
    return String(logDetail.value.messages)
  }
})

const loadDetail = async (id: number) => {
  if (id <= 0) return
  loading.value = true
  logDetail.value = null
  try {
    logDetail.value = await adminAPI.requestContentLogs.getById(id)
  } catch (error) {
    console.error('Failed to load request content log detail:', error)
  } finally {
    loading.value = false
  }
}

const formatTime = (iso: string) => {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

const copyMessages = async () => {
  try {
    await navigator.clipboard.writeText(formattedMessages.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // fallback
  }
}

watch(
  () => props.show,
  (show) => {
    if (show && props.logId > 0) {
      loadDetail(props.logId)
    }
  }
)
</script>
