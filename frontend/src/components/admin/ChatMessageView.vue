<template>
  <div class="space-y-3">
    <template v-for="(item, idx) in renderItems" :key="idx">
      <!-- 时间分隔线 -->
      <div v-if="item.type === 'divider'" class="flex items-center gap-3 py-1">
        <div class="h-px flex-1 bg-gray-200 dark:bg-gray-700" />
        <span class="shrink-0 text-xs text-gray-400 dark:text-gray-500">
          {{ item.label }}
        </span>
        <div class="h-px flex-1 bg-gray-200 dark:bg-gray-700" />
      </div>

      <!-- system 消息：居中横幅 -->
      <div
        v-else-if="item.msg?.role === 'system'"
        class="mx-auto max-w-[85%] rounded-lg border border-gray-200 bg-gray-50 px-4 py-2 dark:border-gray-700 dark:bg-gray-800/50"
      >
        <div class="mb-1 text-center text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">
          System
        </div>
        <div class="text-center text-xs leading-relaxed text-gray-500 dark:text-gray-400 whitespace-pre-wrap break-words">
          {{ item.msg.content }}
        </div>
      </div>

      <!-- user 消息：右对齐 -->
      <div v-else-if="item.msg?.role === 'user'" class="flex justify-end">
        <div class="max-w-[80%]">
          <div class="mb-0.5 text-right text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">
            User
          </div>
          <div class="rounded-2xl rounded-tr-sm bg-primary-500 px-4 py-2.5 text-sm leading-relaxed text-white whitespace-pre-wrap break-words">
            {{ item.msg.content }}
          </div>
        </div>
      </div>

      <!-- assistant 消息：左对齐，markdown 渲染 -->
      <div v-else-if="item.msg?.role === 'assistant'" class="flex justify-start">
        <div class="max-w-[80%]">
          <div class="mb-0.5 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">
            Assistant
          </div>
          <div
            class="chat-markdown rounded-2xl rounded-tl-sm bg-gray-100 px-4 py-2.5 text-sm leading-relaxed text-gray-900 dark:bg-gray-700 dark:text-gray-100"
            v-html="renderMarkdown(item.msg.content)"
          />
        </div>
      </div>
    </template>

    <!-- 空状态 -->
    <div v-if="renderItems.length === 0" class="py-8 text-center text-sm text-gray-400">
      {{ t('common.noData') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import type { ChatMessage } from '@/api/admin/requestContentLogs'

export interface DividerInfo {
  /** 插在第 beforeIndex 条消息之前 */
  beforeIndex: number
  label: string
}

type RenderItem =
  | { type: 'message'; msg: ChatMessage }
  | { type: 'divider'; label: string }

const props = defineProps<{
  messages: ChatMessage[]
  dividers?: DividerInfo[]
}>()

const { t } = useI18n()

// 配置 marked
marked.setOptions({ breaks: true, gfm: true })

/**
 * 将消息和分隔线合并为渲染列表
 */
const renderItems = computed<RenderItem[]>(() => {
  const items: RenderItem[] = []
  const dividerMap = new Map<number, string>()

  if (props.dividers) {
    for (const d of props.dividers) {
      dividerMap.set(d.beforeIndex, d.label)
    }
  }

  props.messages.forEach((msg, idx) => {
    const divLabel = dividerMap.get(idx)
    if (divLabel !== undefined) {
      items.push({ type: 'divider', label: divLabel })
    }
    items.push({ type: 'message', msg })
  })

  return items
})

/**
 * 渲染 markdown 并净化 HTML
 */
function renderMarkdown(content: string): string {
  if (!content) return ''
  try {
    const html = marked.parse(content) as string
    return DOMPurify.sanitize(html)
  } catch {
    return DOMPurify.sanitize(content)
  }
}
</script>

<style scoped>
/* markdown 内容样式 */
.chat-markdown :deep(p) {
  margin: 0.25em 0;
}
.chat-markdown :deep(p:first-child) {
  margin-top: 0;
}
.chat-markdown :deep(p:last-child) {
  margin-bottom: 0;
}

.chat-markdown :deep(pre) {
  margin: 0.5em 0;
  padding: 0.75em;
  border-radius: 0.5rem;
  overflow-x: auto;
  font-size: 0.8em;
  background-color: rgba(0, 0, 0, 0.05);
}
:root.dark .chat-markdown :deep(pre) {
  background-color: rgba(0, 0, 0, 0.3);
}

.chat-markdown :deep(code) {
  font-size: 0.85em;
  padding: 0.1em 0.3em;
  border-radius: 0.25rem;
  background-color: rgba(0, 0, 0, 0.05);
}
:root.dark .chat-markdown :deep(code) {
  background-color: rgba(255, 255, 255, 0.1);
}

.chat-markdown :deep(pre code) {
  padding: 0;
  background: none;
}

.chat-markdown :deep(ul),
.chat-markdown :deep(ol) {
  margin: 0.25em 0;
  padding-left: 1.5em;
}

.chat-markdown :deep(li) {
  margin: 0.15em 0;
}

.chat-markdown :deep(blockquote) {
  margin: 0.5em 0;
  padding-left: 0.75em;
  border-left: 3px solid rgba(0, 0, 0, 0.15);
  color: inherit;
  opacity: 0.8;
}

.chat-markdown :deep(h1),
.chat-markdown :deep(h2),
.chat-markdown :deep(h3),
.chat-markdown :deep(h4) {
  margin: 0.5em 0 0.25em;
  font-weight: 600;
}

.chat-markdown :deep(table) {
  border-collapse: collapse;
  margin: 0.5em 0;
  font-size: 0.9em;
}

.chat-markdown :deep(th),
.chat-markdown :deep(td) {
  border: 1px solid rgba(0, 0, 0, 0.1);
  padding: 0.3em 0.6em;
}
:root.dark .chat-markdown :deep(th),
:root.dark .chat-markdown :deep(td) {
  border-color: rgba(255, 255, 255, 0.1);
}

.chat-markdown :deep(a) {
  color: #3b82f6;
  text-decoration: underline;
}
</style>
