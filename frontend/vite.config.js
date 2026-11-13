import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// In development, Vite forwards every /api request to the Go server.
// So the browser thinks it's the same origin, and fetch('/api/...') just works.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': process.env.API_URL || 'http://localhost:8080',
    },
  },
})
