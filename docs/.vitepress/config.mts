import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "chankit",
  description: "Elegant, type-safe channel operations for Go with functional programming patterns",
  base: '/chankit/',

  head: [
    ['link', { rel: 'icon', href: '/chankit/logo.svg' }],
    ['meta', { name: 'theme-color', content: '#0a0a0a' }],
    ['meta', { name: 'og:type', content: 'website' }],
    ['meta', { name: 'og:title', content: 'chankit - Elegant channel operations for Go' }],
    ['meta', { name: 'og:description', content: 'Type-safe channel operations with functional programming patterns' }],
  ],

  appearance: 'dark',

  themeConfig: {
    logo: '/logo.svg',

    siteTitle: 'chankit',

    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Examples', link: '/examples/data-pipeline' },
      { text: 'API Reference', link: '/api/pipeline' },
      { text: 'Contributing', link: '/contributing' }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Why chankit?', link: '/guide/why-chankit' },
            { text: 'Core Concepts', link: '/guide/core-concepts' },
          ]
        }
      ],
      '/examples/': [
        {
          text: 'Examples',
          items: [
            { text: 'Data Processing Pipeline', link: '/examples/data-pipeline' },
            { text: 'Event Stream Processing', link: '/examples/event-stream' },
            { text: 'Search Results Processing', link: '/examples/search-results' },
            { text: 'Rate Limiting', link: '/examples/rate-limiting' },
            { text: 'Data Transformation Stream', link: '/examples/transformation' },
            { text: 'Real-time Analytics', link: '/examples/analytics' },
          ]
        }
      ],
      '/api/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Pipeline', link: '/api/pipeline' },
            { text: 'Flow Control', link: '/api/flow-control' },
            { text: 'Transformers', link: '/api/transformers' },
            { text: 'Generators', link: '/api/generators' },
            { text: 'Combiners', link: '/api/combiners' },
            { text: 'Selectors', link: '/api/selectors' },
          ]
        }
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/yourusername/chankit' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present chankit contributors'
    },

    search: {
      provider: 'local'
    },

    editLink: {
      pattern: 'https://github.com/yourusername/chankit/edit/main/docs/:path',
      text: 'Edit this page on GitHub'
    },

    lastUpdated: {
      text: 'Updated at',
      formatOptions: {
        dateStyle: 'full',
        timeStyle: 'medium'
      }
    }
  },

  markdown: {
    theme: {
      light: 'github-dark',
      dark: 'github-dark'
    },
    lineNumbers: true
  }
})
