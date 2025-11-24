<template>
  <div class="feature-card bg-dark-bg-soft border border-dark-border rounded-lg p-6 transition-all duration-300 hover:border-primary-500 hover:shadow-lg hover:shadow-primary-900/20 hover:-translate-y-1">
    <div v-if="icon" class="text-4xl mb-4">{{ icon }}</div>
    <h3 class="text-xl font-bold text-primary-400 mb-3">{{ title }}</h3>
    <p class="text-dark-text-muted leading-relaxed">{{ description }}</p>
    <a v-if="link" :href="link" class="inline-flex items-center mt-4 text-primary-400 hover:text-primary-300 transition-colors">
      Learn more →
    </a>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  icon?: string
  title: string
  description: string
  link?: string
}>()
</script>

<style scoped>
.feature-card {
  cursor: pointer;
}
</style>
