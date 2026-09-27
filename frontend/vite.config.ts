import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // CORS не нужен: в dev фронт ходит в API через прокси Vite
      '/api': 'http://localhost:8080',
    },
  },
})
