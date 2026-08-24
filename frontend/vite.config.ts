import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        timeout: 300000,
        proxyTimeout: 300000,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
