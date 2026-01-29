<template>
  <div class="code-block overflow-hidden rounded-lg bg-gray-900">
    <div class="flex items-center justify-between bg-gray-800 p-3">
      <span class="font-mono text-sm text-gray-300">{{ title }}</span>
      <button
        @click="copyCode"
        class="text-sm text-gray-400 transition-colors hover:text-white"
        :class="{ 'text-green-400': copied }"
      >
        {{ copied ? displayCopiedText : displayCopyText }}
      </button>
    </div>
    <pre class="overflow-x-auto p-4 text-sm text-green-400"><code>{{ code }}</code></pre>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    code: string
    title?: string
    copyText?: string
    copiedText?: string
  }>(),
  {
    title: 'Code',
    copyText: '',
    copiedText: ''
  }
)

const copied = ref(false)

const displayCopyText = computed(() => props.copyText || t('guide.copy'))
const displayCopiedText = computed(() => props.copiedText || t('guide.copied'))

async function copyCode() {
  try {
    await navigator.clipboard.writeText(props.code)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy code:', err)
  }
}
</script>

<style scoped>
.code-block pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}

.code-block code {
  font-family: ui-monospace, 'Fira Code', 'JetBrains Mono', monospace;
}
</style>
