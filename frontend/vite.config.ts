import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // 开发期把 /api 转发到后端，前端代码里写相对路径即可，避免跨域。
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
})
