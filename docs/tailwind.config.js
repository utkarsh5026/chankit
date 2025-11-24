/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.md",
    "./.vitepress/**/*.{js,ts,vue}",
    "./guide/**/*.md",
    "./examples/**/*.md",
    "./api/**/*.md",
    "./advanced/**/*.md",
    "./**/*.md"
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
          950: '#082f49',
        },
        dark: {
          bg: '#0a0a0a',
          'bg-alt': '#111111',
          'bg-soft': '#1a1a1a',
          border: '#2a2a2a',
          text: '#e5e5e5',
          'text-muted': '#a3a3a3',
        }
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'Consolas', 'Monaco', 'Courier New', 'monospace'],
      },
    },
  },
  plugins: [],
}
