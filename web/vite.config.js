import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// The browser calls /api/*, which is forwarded to the Go REST API with the /api prefix
// removed. One origin for everything means the API needs no CORS setup.
// Point it at another address with API_URL=http://host:port.
const proxy = {
  '/api': {
    target: process.env.API_URL ?? 'http://localhost:8080',
    rewrite: (path) => path.replace(/^\/api/, ''),
  },
}

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: { proxy },
  preview: { proxy },
})
