<template>
  <div class="api-card bg-dark-bg-soft border border-dark-border rounded-lg p-6 mb-6 transition-all duration-200 hover:border-primary-500/50">
    <div class="flex items-start justify-between mb-3">
      <h3 class="text-2xl font-bold text-primary-400">{{ method }}</h3>
      <span v-if="tag" class="px-3 py-1 text-xs font-semibold rounded-full bg-primary-900/30 text-primary-300 border border-primary-500/30">
        {{ tag }}
      </span>
    </div>

    <p class="text-dark-text-muted mb-4">{{ description }}</p>

    <div v-if="signature" class="bg-dark-bg p-4 rounded-lg mb-4 overflow-x-auto">
      <code class="text-sm text-primary-300 font-mono">{{ signature }}</code>
    </div>

    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  method: string
  description: string
  signature?: string
  tag?: string
}>()
</script>
