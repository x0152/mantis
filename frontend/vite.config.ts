import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  envDir: '..',
  server: {
    host: true,
    proxy: {
      '/api': {
        target: process.env.API_URL || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
