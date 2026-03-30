import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'prompt',
      includeAssets: ['favicon.svg'],
      devOptions: {
        enabled: true  // 开发模式也生成 manifest，不加这个 iOS 添加到主屏幕不生效
      },
      manifest: {
        name: '易账 - Expense Log',
        short_name: '易账',
        description: '基于视觉AI的下一代易账应用',
        theme_color: '#faf8f5',
        background_color: '#faf8f5',
        display: 'standalone',
        icons: [
          {
            src: '/icon-192.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: '/icon-512.png',
            sizes: '512x512',
            type: 'image/png'
          }
        ]
      }
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8808',
        changeOrigin: true
      }
    }
  }
})
