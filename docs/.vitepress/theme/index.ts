import { h } from 'vue'
import type { Theme } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import './style.css'
import FeatureCard from './components/FeatureCard.vue'
import ApiCard from './components/ApiCard.vue'
import CodeExample from './components/CodeExample.vue'
import ComparisonTable from './components/ComparisonTable.vue'

export default {
  extends: DefaultTheme,
  Layout: () => {
    return h(DefaultTheme.Layout, null, {
      // https://vitepress.dev/guide/extending-default-theme#layout-slots
    })
  },
  enhanceApp({ app, router, siteData }) {
    // Register custom components
    app.component('FeatureCard', FeatureCard)
    app.component('ApiCard', ApiCard)
    app.component('CodeExample', CodeExample)
    app.component('ComparisonTable', ComparisonTable)
  }
} satisfies Theme
