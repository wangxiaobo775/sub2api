<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- 过滤器 -->
      <div class="card p-4">
        <div class="flex flex-wrap items-end gap-3">
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.model') }}</label>
            <input
              v-model="filters.model"
              type="text"
              :placeholder="t('admin.requestContentLogs.filterModel')"
              class="input w-40"
              @keyup.enter="applyFilters"
            />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.platform') }}</label>
            <select v-model="filters.platform" class="input w-36" @change="applyFilters">
              <option value="">{{ t('common.all') }}</option>
              <option value="anthropic">Anthropic</option>
              <option value="openai">OpenAI</option>
              <option value="gemini">Gemini</option>
              <option value="antigravity">Antigravity</option>
            </select>
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.userId') }}</label>
            <input
              v-model.number="filters.user_id"
              type="number"
              :placeholder="t('admin.requestContentLogs.filterUserId')"
              class="input w-28"
              @keyup.enter="applyFilters"
            />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.session') }}</label>
            <input
              v-model="filters.session_fingerprint"
              type="text"
              :placeholder="t('admin.requestContentLogs.filterSession')"
              class="input w-40"
              @keyup.enter="applyFilters"
            />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.startDate') }}</label>
            <input
              v-model="filters.start_date"
              type="datetime-local"
              class="input w-48"
            />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.requestContentLogs.endDate') }}</label>
            <input
              v-model="filters.end_date"
              type="datetime-local"
              class="input w-48"
            />
          </div>
          <button class="btn btn-primary h-9" @click="applyFilters">
            {{ t('common.search') }}
          </button>
          <button class="btn btn-secondary h-9" @click="resetFilters">
            {{ t('common.reset') }}
          </button>
        </div>
      </div>

      <!-- 表格 -->
      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead class="bg-gray-50 dark:bg-gray-800">
              <tr>
                <th class="table-th">ID</th>
                <th class="table-th">{{ t('admin.requestContentLogs.time') }}</th>
                <th class="table-th">{{ t('admin.requestContentLogs.user') }}</th>
                <th class="table-th">{{ t('admin.requestContentLogs.apiKey') }}</th>
                <th class="table-th">{{ t('admin.requestContentLogs.model') }}</th>
                <th class="table-th">{{ t('admin.requestContentLogs.platform') }}</th>
                <th class="table-th">{{ t('admin.requestContentLogs.session') }}</th>
                <th class="table-th">IP</th>
                <th class="table-th">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
              <tr v-if="loading">
                <td colspan="9" class="px-4 py-8 text-center text-gray-500">
                  <div class="flex items-center justify-center gap-2">
                    <svg class="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none">
                      <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="opacity-25" />
                      <path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" class="opacity-75" />
                    </svg>
                    {{ t('common.loading') }}
                  </div>
                </td>
              </tr>
              <tr v-else-if="logs.length === 0">
                <td colspan="9" class="px-4 py-8 text-center text-gray-500">
                  {{ t('common.noData') }}
                </td>
              </tr>
              <tr
                v-for="log in logs"
                :key="log.id"
                class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              >
                <td class="table-td font-mono text-xs">{{ log.id }}</td>
                <td class="table-td text-xs whitespace-nowrap">{{ formatTime(log.created_at) }}</td>
                <td class="table-td text-sm">{{ log.user_email || '-' }}</td>
                <td class="table-td text-sm">{{ log.api_key_name || '-' }}</td>
                <td class="table-td">
                  <span class="inline-flex items-center rounded-md bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
                    {{ log.model || '-' }}
                  </span>
                </td>
                <td class="table-td">
                  <span
                    class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium"
                    :class="platformClass(log.platform)"
                  >
                    {{ log.platform }}
                  </span>
                </td>
                <td class="table-td">
                  <button
                    v-if="log.session_fingerprint"
                    class="inline-flex items-center gap-1 font-mono text-xs text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 dark:hover:text-indigo-300"
                    :title="log.session_fingerprint"
                    @click="viewSession(log.session_fingerprint!)"
                  >
                    <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M7.5 8.25h9m-9 3H12m-9.75 1.51c0 1.6 1.123 2.994 2.707 3.227 1.129.166 2.27.293 3.423.379.35.026.67.21.865.501L12 21l2.755-4.133a1.14 1.14 0 01.865-.501 48.172 48.172 0 003.423-.379c1.584-.233 2.707-1.626 2.707-3.228V6.741c0-1.602-1.123-2.995-2.707-3.228A48.394 48.394 0 0012 3c-2.392 0-4.744.175-7.043.513C3.373 3.746 2.25 5.14 2.25 6.741v6.018z" />
                    </svg>
                    {{ log.session_fingerprint.slice(0, 8) }}
                    <span v-if="log.message_count" class="text-gray-400">({{ log.message_offset }}-{{ log.message_count }})</span>
                  </button>
                  <span v-else class="text-gray-400">-</span>
                </td>
                <td class="table-td font-mono text-xs">{{ log.ip_address }}</td>
                <td class="table-td">
                  <button
                    class="text-primary-600 hover:text-primary-800 dark:text-primary-400 dark:hover:text-primary-300 text-sm font-medium"
                    @click="viewDetail(log.id)"
                  >
                    {{ t('common.view') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 分页 -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>
  </AppLayout>

  <!-- 详情弹窗 -->
  <RequestContentLogDetail
    :show="detailVisible"
    :log-id="selectedLogId"
    @close="detailVisible = false"
  />

  <!-- 会话详情弹窗 -->
  <SessionDetailDialog
    :show="sessionVisible"
    :fingerprint="selectedFingerprint"
    @close="sessionVisible = false"
  />
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { RequestContentLogItem } from '@/api/admin/requestContentLogs'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import RequestContentLogDetail from '@/components/admin/RequestContentLogDetail.vue'
import SessionDetailDialog from '@/components/admin/SessionDetailDialog.vue'

const { t } = useI18n()

const logs = ref<RequestContentLogItem[]>([])
const loading = ref(false)
let abortController: AbortController | null = null

const filters = reactive({
  model: '',
  platform: '',
  user_id: undefined as number | undefined,
  session_fingerprint: '',
  start_date: '',
  end_date: ''
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const detailVisible = ref(false)
const selectedLogId = ref(0)

const sessionVisible = ref(false)
const selectedFingerprint = ref('')

const loadLogs = async () => {
  abortController?.abort()
  const c = new AbortController()
  abortController = c
  loading.value = true

  try {
    // 构造过滤参数，将 datetime-local 格式转为 RFC3339
    const queryFilters: Record<string, any> = {}
    if (filters.model) queryFilters.model = filters.model
    if (filters.platform) queryFilters.platform = filters.platform
    if (filters.user_id) queryFilters.user_id = filters.user_id
    if (filters.session_fingerprint) queryFilters.session_fingerprint = filters.session_fingerprint
    if (filters.start_date) queryFilters.start_date = new Date(filters.start_date).toISOString()
    if (filters.end_date) queryFilters.end_date = new Date(filters.end_date).toISOString()

    const res = await adminAPI.requestContentLogs.list(
      pagination.page,
      pagination.page_size,
      queryFilters,
      { signal: c.signal }
    )
    if (!c.signal.aborted) {
      logs.value = res.items || []
      pagination.total = res.total
    }
  } catch (error: any) {
    if (error?.name !== 'AbortError' && error?.code !== 'ERR_CANCELED') {
      console.error('Failed to load request content logs:', error)
    }
  } finally {
    if (abortController === c) {
      loading.value = false
    }
  }
}

const applyFilters = () => {
  pagination.page = 1
  loadLogs()
}

const resetFilters = () => {
  filters.model = ''
  filters.platform = ''
  filters.user_id = undefined
  filters.session_fingerprint = ''
  filters.start_date = ''
  filters.end_date = ''
  pagination.page = 1
  loadLogs()
}

const handlePageChange = (p: number) => {
  pagination.page = p
  loadLogs()
}

const handlePageSizeChange = (s: number) => {
  pagination.page_size = s
  pagination.page = 1
  loadLogs()
}

const viewDetail = (id: number) => {
  selectedLogId.value = id
  detailVisible.value = true
}

const viewSession = (fingerprint: string) => {
  selectedFingerprint.value = fingerprint
  sessionVisible.value = true
}

const formatTime = (iso: string) => {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

const platformClass = (platform: string) => {
  const classes: Record<string, string> = {
    anthropic: 'bg-orange-50 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    openai: 'bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    gemini: 'bg-purple-50 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
    antigravity: 'bg-cyan-50 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400'
  }
  return classes[platform] || 'bg-gray-50 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
}

onMounted(() => {
  loadLogs()
})

onUnmounted(() => {
  abortController?.abort()
})
</script>
