<template>
  <div class="code-example bg-dark-bg-soft border border-dark-border rounded-lg overflow-hidden mb-6">
    <div v-if="title" class="bg-dark-bg-alt px-4 py-2 border-b border-dark-border">
      <span class="text-sm font-semibold text-dark-text">{{ title }}</span>
    </div>

    <div class="relative">
      <slot />
      <button
        v-if="copyable"
        @click="copyCode"
        class="absolute top-2 right-2 px-3 py-1 text-xs bg-primary-600 hover:bg-primary-700 text-white rounded transition-colors"
      >
        {{ copied ? 'Copied!' : 'Copy' }}
      </button>
    </div>

    <div v-if="output" class="bg-dark-bg px-4 py-3 border-t border-dark-border">
      <div class="text-xs text-dark-text-muted mb-1">Output:</div>
      <pre class="text-sm text-green-400 font-mono">{{ output }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  title?: string
  output?: string
  copyable?: boolean
  code?: string
}>()

const copied = ref(false)

const copyCode = async () => {
  if (props.code) {
    await navigator.clipboard.writeText(props.code)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  }
}
</script>
