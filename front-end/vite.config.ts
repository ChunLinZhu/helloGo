import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const port = Number(process.env.VITE_PORT || 9003)

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: Number.isFinite(port) ? port : 9003,
    strictPort: false,
  },
  build: {
    outDir: 'dist',
  },
})
