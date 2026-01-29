<template>
  <div class="space-y-4">
    <button type="button" :disabled="disabled" class="btn btn-secondary w-full" @click="startLogin">
      <svg
        class="mr-2 h-5 w-5"
        viewBox="0 0 1024 1024"
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden="true"
      >
        <!-- DingTalk Logo -->
        <path
          d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64z"
          fill="#3296FA"
        />
        <path
          d="M715.5 546.1c-5.8-18.3-23.8-33.9-54.4-46.9-28.7-12.2-52.9-11.3-72.3-3.4-7.9 3.2-14.3 7.2-19.3 11.4l-46.6-20.5c19.3-35.1 26.8-77.7 21.3-127.4-7.8-70.5-48.2-132.2-107.9-166.7-59.7-34.5-131.1-37.5-193.5-8.1-56.9 26.8-96.4 76.1-108.8 135.6-12.4 59.5 3.5 119.9 43.8 166.1l-2.7 1.6c-15.5 9.5-28.7 21.3-39 34.4-10.3 13.1-17.9 28.2-22.2 44.8-4.3 16.6-5.3 34.4-2.2 52.9 3.1 18.5 10.6 37 22.8 54.1 25.9 36.4 67.1 61.1 116.2 70.7 49.1 9.6 103.1 3.7 152.1-16.4l147.9 74c9.3 4.6 18.7 7 27.8 7 9.1 0 17.9-2.4 26-7.3 8.1-4.9 15.1-12 20.6-21.4 5.5-9.4 9.3-21.1 10.8-34.6 1.5-13.5 0.7-29-3.4-45.3-4.1-16.3-10.8-32.5-20.4-46.8 15.2-20.9 24.2-45.9 26.5-72.7 2.3-26.8-2.3-54.8-13.1-81.1zm-284.2 144.2c-33.8 13.8-71.3 18-107.2 10.9-35.9-7.1-67.8-23.7-87-46.7-9.6-11.4-15.6-23.5-17.7-35.3-2.1-11.8-1.4-23.2 2.1-33.6 3.5-10.4 9.3-19.8 17.1-27.5 7.8-7.7 17.6-13.7 29.4-17.5l8.6 27.3c-7.3 2.4-13.3 5.9-18.2 10.5-4.9 4.6-8.5 10.3-10.5 16.7-2 6.4-2.3 13.5-0.6 20.8 1.7 7.3 5.4 14.7 11.2 21.6 11.6 13.8 30.9 24.1 53.6 29.5 22.7 5.4 48.5 5.1 72.9-1.9 24.4-7 46.2-20 62.3-37.2 16.1-17.2 26-38.3 27.1-60.3l-59.1 24c-3.4 1.4-7.1 2-10.8 1.9-3.7-0.1-7.3-1-10.5-2.5-6.4-3-11.2-8.5-13-15.5-1.8-7 0.1-14.7 5.4-21l88.7-105.3c5.1-6.1 12.4-9.8 20.5-10.3 8.1-0.5 16.2 2.2 22.5 7.7l1.3 1.1c6.5 5.7 10.2 13.7 10.5 22 0.3 8.3-2.7 16.7-8.5 23.2l-23.2 26.1c31.3-4.7 63.2 2.2 89.7 19.3l41.1 26.5c-14.7-2.6-30.7-0.7-46.5 6.6-15.8 7.3-30.9 19.5-43.3 35.2l-117.4 58.7c-1.4 0.7-3 1.2-4.7 1.5l-0.3 0.1c-8.6 9.7-18.8 18.1-30.6 24.6-11.8 6.5-25.1 11.5-39.6 14.4z"
          fill="#FFFFFF"
        />
      </svg>
      {{ t('auth.dingtalk.signIn') }}
    </button>

    <div class="flex items-center gap-3">
      <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
      <span class="text-xs text-gray-500 dark:text-dark-400">
        {{ t('auth.dingtalk.orContinue') }}
      </span>
      <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

defineProps<{
  disabled?: boolean
}>()

const route = useRoute()
const { t } = useI18n()

function startLogin(): void {
  const redirectTo = (route.query.redirect as string) || '/dashboard'
  const apiBase = (import.meta.env.VITE_API_BASE_URL as string | undefined) || '/api/v1'
  const normalized = apiBase.replace(/\/$/, '')
  const startURL = `${normalized}/auth/oauth/dingtalk/start?redirect=${encodeURIComponent(redirectTo)}`
  window.location.href = startURL
}
</script>
